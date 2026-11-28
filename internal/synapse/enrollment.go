package synapse

import (
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"strings"
)

// EnrolledServerTLS adds an explicit leaf-certificate allowlist to CA trust.
// Two fingerprints per identity permit a controlled certificate rollover. The
// map is copied and cannot be mutated by callers after startup.
func EnrolledServerTLS(certificate tls.Certificate, neuronCAs *x509.CertPool, enrolled map[string][]string) (*tls.Config, error) {
	if len(enrolled) == 0 || len(enrolled) > 4096 {
		return nil, ErrTLSIdentity
	}
	pins := make(map[string]map[string]bool, len(enrolled))
	for id, fingerprints := range enrolled {
		if !neuronName.MatchString(id) || len(fingerprints) == 0 || len(fingerprints) > 2 {
			return nil, ErrTLSIdentity
		}
		pins[id] = make(map[string]bool, len(fingerprints))
		for _, fingerprint := range fingerprints {
			if !validDigest(fingerprint) || pins[id][fingerprint] {
				return nil, ErrTLSIdentity
			}
			pins[id][fingerprint] = true
		}
	}
	cfg, err := ServerTLS(certificate, neuronCAs)
	if err != nil {
		return nil, err
	}
	cfg.VerifyConnection = func(state tls.ConnectionState) error {
		id, err := NeuronIdentity(state)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(state.PeerCertificates[0].Raw)
		if !pins[id]["sha256:"+hex.EncodeToString(sum[:])] {
			return ErrTLSIdentity
		}
		return nil
	}
	return cfg, nil
}

func validDigest(value string) bool {
	raw, ok := strings.CutPrefix(value, "sha256:")
	if !ok || len(raw) != 64 || raw != strings.ToLower(raw) {
		return false
	}
	_, err := hex.DecodeString(raw)
	return err == nil
}
