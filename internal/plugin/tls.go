package plugin

import (
	"crypto/tls"
	"crypto/x509"
)

func ExternalTLSConfig(cert tls.Certificate, caPool *x509.CertPool) *tls.Config {
	return &tls.Config{Certificates: []tls.Certificate{cert}, RootCAs: caPool, ClientCAs: caPool, MinVersion: tls.VersionTLS13, ClientAuth: tls.RequireAndVerifyClientCert}
}
