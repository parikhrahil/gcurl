package transport_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/parikhrahil/gcurl/pkg/transport"
	"github.com/stretchr/testify/require"
)

// Helper: Dynamically generates a valid, cryptographically signed PEM certificate
// in memory to prevent testing breakages caused by file expiration invariants.
func generateTestPEM(t *testing.T) []byte {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	template := x509.Certificate{
		SerialNumber: big.NewInt(1337),
		Subject: pkix.Name{
			Organization: []string{"gcurl Architecture Test Authority"},
		},
		NotBefore:             time.Now().Add(-1 * time.Hour),
		NotAfter:              time.Now().Add(1 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	// Self-sign the template certificate structure
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	require.NoError(t, err)

	// Encode the raw DER bytes straight into a readable PEM format block
	pemBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: derBytes,
	}

	return pem.EncodeToMemory(pemBlock)
}

func TestBuildTLSConfig_ValidParameters_Succeeds(t *testing.T) {
	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "ca.crt")

	validPEM := generateTestPEM(t)
	err := os.WriteFile(certPath, validPEM, 0o644)
	require.NoError(t, err)

	tlsConfig, err := transport.BuildTlsConfig(certPath, "1.3")

	if tlsConfig == nil {
		t.Fatal("Structural Invariant Violated: Security factory returned a nil tls.Config pointer")
	}

	if tlsConfig.MinVersion != tls.VersionTLS13 {
		t.Errorf("Protocol Mismatch: Factory failed to bind min version to TLS 1.3. Got: 0x%X", tlsConfig.MinVersion)
	}

	if tlsConfig.RootCAs == nil {
		t.Error("Data Layer Omission: Root certificate authority trust pool remained uninitialized")
	}
}

func TestBuildTLSConfig_UnsupportedVersion_FailsFast(t *testing.T) {
	tlsConfig, err := transport.BuildTlsConfig("", "1.4")
	require.Error(t, err)
	if tlsConfig != nil {
		t.Error("Resource Leak: Factory returned a hydrated configuration block under a failing state parameter")
	}
}

func TestBuildTLSConfig_MissingFile_ReturnsWrappedError(t *testing.T) {
	_, err := transport.BuildTlsConfig("/tmp/ghost_path/non_existent.crt", "1.2")
	require.Error(t, err)
}

func TestBuildTLSConfig_MalformedPEM_FailsValidation(t *testing.T) {
	tmpDir := t.TempDir()
	corruptedCertPath := filepath.Join(tmpDir, "corrupted.crt")

	garbagePEM := []byte("-----BEGIN CERTIFICATE-----\nTHIS-IS-NOT-VALID-CRYPTO-DATA\n-----END CERTIFICATE-----")
	if err := os.WriteFile(corruptedCertPath, garbagePEM, 0o644); err != nil {
		t.Fatalf("Setup failure writing mock data: %v", err)
	}

	_, err := transport.BuildTlsConfig(corruptedCertPath, "1.2")
	require.Error(t, err)
}
