package synapse_test

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"net"
	"testing"
	"time"

	"brain/internal/cortex"
	"brain/internal/execution"
	"brain/internal/neuron"
	"brain/internal/synapse"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func fingerprint(cert tls.Certificate) string {
	sum := sha256.Sum256(cert.Certificate[0])
	return "sha256:" + hex.EncodeToString(sum[:])
}

func TestEnrolledTLSRejectsUnlistedCertificate(t *testing.T) {
	a := ca(t)
	brain := leaf(t, a, true)
	worker := leaf(t, a, false, "spiffe://brain/neurons/worker")
	replacement := leaf(t, a, false, "spiffe://brain/neurons/worker")
	other := leaf(t, a, false, "spiffe://brain/neurons/other")
	pins := map[string][]string{"worker": {fingerprint(worker)}}
	cfg, err := synapse.EnrolledServerTLS(brain, a.pool, pins)
	if err != nil {
		t.Fatal(err)
	}
	pins["worker"][0] = fingerprint(replacement)
	for _, tc := range []struct {
		cert tls.Certificate
		ok   bool
	}{{worker, true}, {replacement, false}, {other, false}} {
		client, err := synapse.ClientTLS(tc.cert, a.pool, "brain.test")
		if err != nil {
			t.Fatal(err)
		}
		_, se, ce := handshake(t, cfg, client)
		if (se == nil && ce == nil) != tc.ok {
			t.Fatalf("unexpected enrollment result: %v %v", se, ce)
		}
	}
	for _, invalid := range []map[string][]string{nil, {"bad/id": {fingerprint(worker)}}, {"worker": {"bad"}}, {"worker": {fingerprint(worker), fingerprint(worker)}}} {
		if _, err := synapse.EnrolledServerTLS(brain, a.pool, invalid); err == nil {
			t.Fatal("invalid enrollment accepted")
		}
	}
}

func TestEnrolledGRPCRequiresRuntimeIdentity(t *testing.T) {
	a := ca(t)
	cert := leaf(t, a, false, "spiffe://brain/neurons/worker")
	serverTLS, err := synapse.EnrolledServerTLS(leaf(t, a, true), a.pool, map[string][]string{"worker": {fingerprint(cert)}})
	if err != nil {
		t.Fatal(err)
	}
	clientTLS, err := synapse.ClientTLS(cert, a.pool, "brain.test")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := execution.NewCatalog(4, 4096)
	if err != nil {
		t.Fatal(err)
	}
	scheduler, err := cortex.New(catalog, cortex.Config{MaxNeurons: 4, MaxProjectRunning: 2, HeartbeatTTL: 8 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	digest := execution.ArtifactDigest([]byte("runtime"))
	if _, err := synapse.NewEnrolledGRPCServer(serverTLS, scheduler, catalog, time.Second, ""); err == nil {
		t.Fatal("missing runtime identity accepted")
	}
	server, err := synapse.NewEnrolledGRPCServer(serverTLS, scheduler, catalog, time.Second, digest)
	if err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go server.Serve(listener)
	defer server.Stop()
	adapter := wireRuntime(func(context.Context, execution.Request, []execution.BundleFile) (execution.Result, error) {
		t.Error("unexpected execution")
		return execution.Result{}, nil
	})
	cfg := neuron.Config{ID: "worker", Slots: 1, MaxArtifactBytes: 4096, Adapters: []neuron.Adapter{{Capability: execution.Runtime{Kind: "typescript", Profile: "isolated"}, Runtime: adapter}}}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	for _, values := range [][]string{nil, {execution.ArtifactDigest([]byte("other"))}, {digest, digest}} {
		call := ctx
		for _, value := range values {
			call = metadata.AppendToOutgoingContext(call, synapse.RuntimeDigestHeader, value)
		}
		err := synapse.RunNeuron(call, listener.Addr().String(), clientTLS, cfg)
		if status.Code(err) != codes.FailedPrecondition || len(scheduler.Snapshot().Neurons) != 0 {
			t.Fatalf("incompatible runtime registered: %v", err)
		}
	}
	done := make(chan error, 1)
	go func() {
		done <- synapse.RunNeuron(metadata.AppendToOutgoingContext(ctx, synapse.RuntimeDigestHeader, digest), listener.Addr().String(), clientTLS, cfg)
	}()
	for len(scheduler.Snapshot().Neurons) != 1 {
		select {
		case err := <-done:
			t.Fatal(err)
		case <-ctx.Done():
			t.Fatal("compatible enrolled Neuron did not register")
		case <-time.After(10 * time.Millisecond):
		}
	}
	cancel()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("enrolled Neuron did not disconnect")
	}
}
