package cmd_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/parikhrahil/gcurl/cmd"
	"github.com/parikhrahil/gcurl/pkg/audit"
	"github.com/parikhrahil/gcurl/pkg/config"
)

func TestNewHistoryCommand_ExecutionAndTabularPresentation(t *testing.T) {
	tmpDir := t.TempDir()

	t.Setenv("HOME", tmpDir)        // Directs Linux/Unix/Darwin profiles
	t.Setenv("USERPROFILE", tmpDir) // Directs Windows profiles

	wsMgr, err := audit.BootstrapWorkspace() // Ensure this helper exists from previous steps
	if err != nil {
		t.Fatalf("Failed to initialize test workspace: %v", err)
	}

	repo, err := audit.NewHistoryRepository(wsMgr.DbPath)
	if err != nil {
		t.Fatalf("Failed to spin up test storage channel: %v", err)
	}

	mockRecord := &config.RequestConfiguration{
		Method: "GET",
		URL:    "https://example.com",
		Metrics: config.AuditMetrics{
			BytesTransmitted: 250,
			BytesReceived:    1024,
			TotalDuration:    45 * time.Millisecond,
		},
	}
	if err := repo.WriteAuditTrail(mockRecord); err != nil {
		t.Fatalf("Failed to pre-populate telemetry ledger table: %v", err)
	}
	repo.Close() // Flush locks so the command pointer can safely read it

	historyCmd := cmd.NewHistoryCommand()
	stdoutBuf := new(bytes.Buffer)
	historyCmd.SetOut(stdoutBuf)

	// Set optional limit flag parameter explicitly
	historyCmd.SetArgs([]string{"--limit", "5"})

	if err := historyCmd.Execute(); err != nil {
		t.Fatalf("Presentation Contract Error: Command failed to execute: %v", err)
	}

	output := stdoutBuf.String()

	headers := []string{"ID", "METHOD", "TARGET URL", "STATUS", "SENT", "RCVD", "LATENCY", "EXECUTED (UTC)"}
	for _, header := range headers {
		if !strings.Contains(output, header) {
			t.Errorf("Presentation Layout Error: Expected output to contain column header [%s]", header)
		}
	}

	expectedDataTokens := []string{"GET", "https://example.com", "1024 B", "45ms"}
	for _, token := range expectedDataTokens {
		if !strings.Contains(output, token) {
			t.Errorf("Data Visibility Error: Expected layout stream to print mapped token [%s]. Output:\n%s", token, output)
		}
	}
}
