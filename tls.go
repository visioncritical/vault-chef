package main

import (
	"crypto/tls"
)

// ParseX509KeyPair parses a public/private key pair from a pair of PEM encoded data
func ParseX509KeyPair(certPEM, keyPEM string) (tls.Certificate, error) {
	cert, err := tls.X509KeyPair([]byte(certPEM), []byte(keyPEM))
	return cert, err
}

// TLSConfig creates a tls.Config struct
// LS_ECDHE_RSA_WITH_AES_128_GCM_SHA256 is required:
// https://github.com/golang/net/commit/42ad50856d0131521b13983610e75c16da945ae3
func TLSConfig(cert tls.Certificate) *tls.Config {
	return &tls.Config{
		Certificates:             []tls.Certificate{cert},
		MinVersion:               tls.VersionTLS12,
		CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		},
	}
}
