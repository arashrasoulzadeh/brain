package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"brain/internal/execution"
	"brain/internal/neuron"
	"brain/internal/synapse"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func neuronConnect(args []string) error {
	flags := flag.NewFlagSet("neuron connect", flag.ContinueOnError)
	address := flags.String("brain", "", "Synapse gRPC host:port")
	id := flags.String("id", "", "certificate-bound Neuron identity")
	certificate := flags.String("cert", "", "Neuron client certificate")
	key := flags.String("key", "", "Neuron private key")
	ca := flags.String("ca", "", "Brain CA certificate")
	serverName := flags.String("server-name", "", "Brain TLS server name")
	slots := flags.Int("slots", 1, "concurrent isolated invocations")
	root := flags.String("runtime-root", ".", "trusted Brain runtime directory")
	localMacOS := flags.Bool("local-macos", false, "allow local Seatbelt development only")
	pluginRuntime := flags.Bool("plugin-runtime", false, "enable prepared-plugin runtime for managed loopback development")
	if err := flags.Parse(args); err != nil {
		return err
	}
	host, _, err := net.SplitHostPort(*address)
	if err != nil || *id == "" || *certificate == "" || *key == "" || *ca == "" || *serverName == "" || *slots < 1 || *slots > 8 || flags.NArg() != 0 || ((*localMacOS || *pluginRuntime) && host != "127.0.0.1") {
		return errors.New("neuron connect requires brain, id, cert, key, ca, server-name and 1–8 slots; macOS must use loopback")
	}
	cert, err := tls.LoadX509KeyPair(*certificate, *key)
	if err != nil {
		return err
	}
	pem, err := os.ReadFile(*ca)
	if err != nil {
		return err
	}
	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(pem) {
		return errors.New("invalid Brain CA")
	}
	tlsConfig, err := synapse.ClientTLS(cert, roots, *serverName)
	if err != nil {
		return err
	}
	runner, err := neuron.NewTypeScriptRuntime(*root, *localMacOS)
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err = runner.Probe(ctx); err != nil {
		return err
	}
	adapters := []neuron.Adapter{{Capability: runner.Capability(), Runtime: runner}}
	if *pluginRuntime {
		plugins, err := neuron.NewPluginTypeScriptRuntime(*root, *localMacOS)
		if err != nil {
			return err
		}
		if err := plugins.Probe(ctx); err != nil {
			return err
		}
		adapters = append(adapters, neuron.Adapter{Capability: plugins.Capability(), Runtime: plugins})
	}
	cfg := neuron.Config{ID: *id, Slots: *slots, MaxArtifactBytes: execution.MaxBundleBytes, Adapters: adapters}
	log.Printf("%s: isolated runtime probe passed (%s)", *id, runner.Capability().Profile)
	for ctx.Err() == nil {
		err = synapse.RunNeuron(ctx, *address, tlsConfig, cfg)
		if ctx.Err() != nil {
			return nil
		}
		switch status.Code(err) {
		case codes.Unauthenticated, codes.PermissionDenied, codes.InvalidArgument, codes.FailedPrecondition:
			return err
		}
		if errors.Is(err, synapse.ErrProtocol) {
			return err
		}
		log.Printf("%s: Synapse disconnected; reconnecting without replay", *id)
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(time.Second):
		}
	}
	return nil
}
