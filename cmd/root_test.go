package cmd_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/parikhrahil/gcurl/cmd"
	"github.com/parikhrahil/gcurl/pkg/config"
	"github.com/spf13/cobra"
)

func TestRootCommand_ConcurrencyCoercionInvariant(t *testing.T) {
	rootCmd := cmd.NewRootCommand()

	stdoutBuf := new(bytes.Buffer)
	stderrBuf := new(bytes.Buffer)

	rootCmd.SetOut(stdoutBuf)
	rootCmd.SetErr(stderrBuf)

	rootCmd.SetArgs([]string{"https://localhost:8080", "-c", "5", "-n", "2"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Unexpected execution barrier encountered: %v", err)
	}

	stderrOutput := stderrBuf.String()
	expectedNotice := "Auto-adjusting concurrency pool to 2"

	if !strings.Contains(stderrOutput, expectedNotice) {
		t.Errorf("Boundary Failure: Expected stderr to emit coercion warning notice, but recorded:\n%s",
			stderrOutput)
	}
}

func TestRootCommand_InvalidProtocolScheme_FailsFast(t *testing.T) {
	rootCmd := cmd.NewRootCommand()

	stdoutBuf := new(bytes.Buffer)
	stderrBuf := new(bytes.Buffer)
	rootCmd.SetOut(stdoutBuf)
	rootCmd.SetErr(stderrBuf)

	// Provide an invalid protocol schema string to hit the fail-fast check
	rootCmd.SetArgs([]string{"ftp://malformed-target.local"})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("Invariant Guard Broken: Expected application to fail fast on invalid protocol scheme, but it passed silently")
	}

	expectedErrorMsg := "invalid protocol"
	if !strings.Contains(err.Error(), expectedErrorMsg) {
		t.Errorf("Unexpected error contract returned. Got: %v, Expected context containing: %s",
			err, expectedErrorMsg)
	}
}

func TestExecuteRequest_SingleRequest_HappyPathWithPayload(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("USERPROFILE", tmpDir)

	expectedResponseBody := "architectural-telemetry-validated"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Method Corruption: Expected POST, but orchestrator sent: %s", r.Method)
		}

		if r.Header.Get("X-Correlation-ID") != "test-sig-123" {
			t.Errorf("Header Drop: Target header missing or malformed")
		}

		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(expectedResponseBody))
	}))
	defer server.Close()

	cfg := &config.RequestConfiguration{
		Method:        "POST",
		URL:           server.URL,
		TotalRequests: 1, // Enforces single-request code path execution
		Concurrency:   1,
		Data:          "client-payload-bytes",
		Verbose:       true, // Forces execution of printing loops to maximize coverage
		Headers: http.Header{
			"X-Correlation-ID": []string{"test-sig-123"},
		},
	}

	mockCmd := &cobra.Command{}
	stdoutBuf := new(bytes.Buffer)
	stderrBuf := new(bytes.Buffer)
	mockCmd.SetOut(stdoutBuf)
	mockCmd.SetErr(stderrBuf)

	err := cmd.ExecuteRequest(mockCmd, cfg)
	if err != nil {
		t.Fatalf("Orchestrator Failure: Single request path returned unexpected error: %v", err)
	}

	capturedOutput := stdoutBuf.String()
	if capturedOutput != expectedResponseBody {
		t.Errorf("Data Plane Breach: Output stream data mismatch.\nGot: %s\nExpected: %s",
			capturedOutput, expectedResponseBody)
	}

	capturedDiagnostics := stderrBuf.String()
	if !strings.Contains(capturedDiagnostics, "Connection established.") {
		t.Errorf("Presentation Omission: Verbose network connection data missing from stderr logs")
	}

	if cfg.Metrics.BytesTransmitted != int64(len("client-payload-bytes")) {
		t.Errorf("Telemetry Skew: Tx counter registered %d, expected %d",
			cfg.Metrics.BytesTransmitted, len("client-payload-bytes"))
	}
	if cfg.Metrics.BytesReceived != int64(len(expectedResponseBody)) {
		t.Errorf("Telemetry Skew: Rx counter registered %d, expected %d",
			cfg.Metrics.BytesReceived, len(expectedResponseBody))
	}
	if cfg.Metrics.TotalDuration <= 0 {
		t.Error("Clock Fault: System failed to record positive monotonic wall-clock elapsed execution metrics")
	}
}

func TestExecuteRequest_SingleRequest_NetworkUnreachable_FailsGracefully(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("USERPROFILE", tmpDir)

	cfg := &config.RequestConfiguration{
		Method:        "GET",
		URL:           "http://completely-unresolvable-domain-local.dev",
		TotalRequests: 1,
		Concurrency:   1,
	}

	mockCmd := &cobra.Command{}
	stdoutBuf := new(bytes.Buffer)
	stderrBuf := new(bytes.Buffer)
	mockCmd.SetOut(stdoutBuf)
	mockCmd.SetErr(stderrBuf)

	err := cmd.ExecuteRequest(mockCmd, cfg)
	if err == nil {
		t.Fatal("Defensive Guard Failure: Orchestrator passed silently during a total transport layer dropout")
	}
}
