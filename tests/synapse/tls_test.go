package synapse_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"math/big"
	"net"
	"net/url"
	"testing"
	"time"

	"brain/internal/synapse"
)

type authority struct {
	cert *x509.Certificate
	key  *ecdsa.PrivateKey
	pool *x509.CertPool
}

func ca(t *testing.T) authority {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	cert := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "test CA"}, NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour), KeyUsage: x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature, IsCA: true, BasicConstraintsValid: true}
	der, err := x509.CreateCertificate(rand.Reader, cert, cert, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	pool := x509.NewCertPool()
	pool.AddCert(parsed)
	return authority{parsed, key, pool}
}
func leaf(t *testing.T, a authority, server bool, identities ...string) tls.Certificate {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	cert := &x509.Certificate{SerialNumber: big.NewInt(2), NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour), KeyUsage: x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}}
	if server {
		cert.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}
		cert.DNSNames = []string{"brain.test"}
	}
	for _, identity := range identities {
		u, err := url.Parse(identity)
		if err != nil {
			t.Fatal(err)
		}
		cert.URIs = append(cert.URIs, u)
	}
	der, err := x509.CreateCertificate(rand.Reader, cert, a.cert, &key.PublicKey, a.key)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key, Leaf: parsed}
}
func handshake(t *testing.T, server, client *tls.Config) (tls.ConnectionState, error, error) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	right, err := net.DialTimeout("tcp", listener.Addr().String(), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer right.Close()
	left, err := listener.Accept()
	if err != nil {
		t.Fatal(err)
	}
	defer left.Close()
	deadline := time.Now().Add(2 * time.Second)
	_ = left.SetDeadline(deadline)
	_ = right.SetDeadline(deadline)
	s, c := tls.Server(left, server), tls.Client(right, client)
	done := make(chan error, 1)
	go func() { done <- s.Handshake() }()
	clientErr := c.Handshake()
	if clientErr != nil {
		_ = right.Close()
	}
	serverErr := <-done
	return s.ConnectionState(), serverErr, clientErr
}
func TestMutualTLSAndIdentity(t *testing.T) {
	a := ca(t)
	server := leaf(t, a, true)
	worker := leaf(t, a, false, "https://example.test/other", "spiffe://brain/neurons/n-1")
	s, err := synapse.ServerTLS(server, a.pool)
	if err != nil {
		t.Fatal(err)
	}
	c, err := synapse.ClientTLS(worker, a.pool, "brain.test")
	if err != nil {
		t.Fatal(err)
	}
	state, se, ce := handshake(t, s, c)
	if se != nil || ce != nil {
		t.Fatal(se, ce)
	}
	id, err := synapse.NeuronIdentity(state)
	if err != nil || id != "n-1" || state.NegotiatedProtocol != "h2" {
		t.Fatal(id, err, state.NegotiatedProtocol)
	}
	if _, err := synapse.ServerTLS(tls.Certificate{}, a.pool); !errors.Is(err, synapse.ErrTLSIdentity) {
		t.Fatal(err)
	}
	if _, err := synapse.ServerTLS(server, nil); !errors.Is(err, synapse.ErrTLSIdentity) {
		t.Fatal(err)
	}
	if _, err := synapse.ClientTLS(worker, a.pool, ""); !errors.Is(err, synapse.ErrTLSIdentity) {
		t.Fatal(err)
	}
	if _, err := synapse.ClientTLS(worker, nil, "brain.test"); !errors.Is(err, synapse.ErrTLSIdentity) {
		t.Fatal(err)
	}
	for _, uris := range [][]string{nil, {"spiffe://brain/other/n"}, {"spiffe://brain/neurons/n%2Fother"}, {"spiffe://brain/neurons/n", "spiffe://brain/neurons/x"}, {"spiffe://other/neurons/n"}} {
		bad := leaf(t, a, false, uris...)
		cfg, err := synapse.ClientTLS(bad, a.pool, "brain.test")
		if err != nil {
			t.Fatal(err)
		}
		_, se, _ := handshake(t, s, cfg)
		if !errors.Is(se, synapse.ErrTLSIdentity) {
			t.Fatalf("bad identity accepted %v: %v", uris, se)
		}
	}
	untrusted := ca(t)
	bad := leaf(t, untrusted, false, "spiffe://brain/neurons/n")
	cfg, err := synapse.ClientTLS(bad, a.pool, "brain.test")
	if err != nil {
		t.Fatal(err)
	}
	_, se, ce = handshake(t, s, cfg)
	if se == nil {
		t.Fatal("foreign CA accepted", ce)
	}
	wrongName, err := synapse.ClientTLS(worker, a.pool, "wrong.test")
	if err != nil {
		t.Fatal(err)
	}
	_, _, ce = handshake(t, s, wrongName)
	if ce == nil {
		t.Fatal("wrong server name accepted")
	}
	for _, v := range []tls.ConnectionState{{}, {Version: tls.VersionTLS13, PeerCertificates: []*x509.Certificate{worker.Leaf}}, {Version: tls.VersionTLS13, PeerCertificates: []*x509.Certificate{worker.Leaf}, VerifiedChains: [][]*x509.Certificate{{server.Leaf}}}} {
		if _, err := synapse.NeuronIdentity(v); !errors.Is(err, synapse.ErrTLSIdentity) {
			t.Fatal(err)
		}
	}
}
