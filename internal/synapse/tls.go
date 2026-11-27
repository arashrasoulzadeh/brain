package synapse

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"regexp"
	"strings"
)

var ErrTLSIdentity = errors.New("unverified synapse TLS identity")
var neuronName = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$`)

const neuronURIPrefix = "spiffe://brain/neurons/"

// NeuronIdentity extracts a single canonical identity from the verified leaf
// certificate. A registration message must match it exactly. Do not construct
// ConnectionState from untrusted wire data; use the authenticated TLS transport.
func NeuronIdentity(state tls.ConnectionState) (string, error) {
	if state.Version < tls.VersionTLS13 || len(state.PeerCertificates) == 0 || len(state.VerifiedChains) == 0 || len(state.VerifiedChains[0]) == 0 || !bytes.Equal(state.PeerCertificates[0].Raw, state.VerifiedChains[0][0].Raw) {
		return "", ErrTLSIdentity
	}
	id := ""
	for _, uri := range state.PeerCertificates[0].URIs {
		if uri.Scheme != "spiffe" || uri.Host != "brain" {
			continue
		}
		// Reject ambiguous, escaped, noncanonical or differently scoped Brain IDs.
		raw := uri.String()
		candidate, ok := strings.CutPrefix(raw, neuronURIPrefix)
		if !ok || !neuronName.MatchString(candidate) || id != "" {
			return "", ErrTLSIdentity
		}
		id = candidate
	}
	if id == "" {
		return "", ErrTLSIdentity
	}
	return id, nil
}

// ServerTLS requires TLS 1.3 and a client certificate from the explicitly
// configured Neuron CA. The URI identity is checked during the TLS handshake.
func ServerTLS(certificate tls.Certificate, neuronCAs *x509.CertPool) (*tls.Config, error) {
	if len(certificate.Certificate) == 0 || certificate.PrivateKey == nil || neuronCAs == nil {
		return nil, ErrTLSIdentity
	}
	return &tls.Config{MinVersion: tls.VersionTLS13, Certificates: []tls.Certificate{certificate}, ClientAuth: tls.RequireAndVerifyClientCert, ClientCAs: neuronCAs.Clone(), NextProtos: []string{"h2"}, VerifyConnection: func(state tls.ConnectionState) error { _, err := NeuronIdentity(state); return err }}, nil
}

// ClientTLS requires both an explicit server trust root and the intended server
// name. It never uses InsecureSkipVerify or silently falls back to plaintext.
func ClientTLS(certificate tls.Certificate, brainCAs *x509.CertPool, serverName string) (*tls.Config, error) {
	if len(certificate.Certificate) == 0 || certificate.PrivateKey == nil || brainCAs == nil || strings.TrimSpace(serverName) == "" {
		return nil, ErrTLSIdentity
	}
	return &tls.Config{MinVersion: tls.VersionTLS13, Certificates: []tls.Certificate{certificate}, RootCAs: brainCAs.Clone(), ServerName: serverName, NextProtos: []string{"h2"}}, nil
}
