package transport

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
)

var tlsVersionMap = map[string]uint16{
	"1.0": tls.VersionTLS10,
	"1.1": tls.VersionTLS11,
	"1.2": tls.VersionTLS12,
	"1.3": tls.VersionTLS13,
}

func BuildTlsConfig(caCertPath, tlsMinVer string) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	if caCertPath != "" {
		pemCerts, err := os.ReadFile(caCertPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read target CA certificate file descriptor: %w", err)
		}

		certPool := x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(pemCerts) {
			return nil, fmt.Errorf("cryptographic failure: malformed or unparseable PEM blocks inside CA certificate")
		}

		tlsConfig.RootCAs = certPool
	}

	if tlsMinVer != "" {
		versionCode, exists := tlsVersionMap[tlsMinVer]
		if !exists {
			return nil, fmt.Errorf("unsupported cryptographic protocol version: %s (choose: 1.2, 1.3)", tlsMinVer)
		}
		tlsConfig.MinVersion = versionCode
	}

	return tlsConfig, nil
}
