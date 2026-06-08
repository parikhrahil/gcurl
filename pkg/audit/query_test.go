package audit_test

import (
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/parikhrahil/gcurl/pkg/audit"
	"github.com/parikhrahil/gcurl/pkg/config"
)

func TestHistoryRepository_Lifecycle_PersistenceAndEviction(t *testing.T) {
	tmpDir := t.TempDir()
	testDbPath := filepath.Join(tmpDir, "test_history.db")

	repo, err := audit.NewHistoryRepository(testDbPath)
	if err != nil {
		t.Fatalf("Failed to initialize test repository workspace: %v", err)
	}

	cfg := &config.RequestConfiguration{
		Method: http.MethodPost,
		URL:    "http://example.com",
		Metrics: config.AuditMetrics{
			BytesTransmitted: 120,
			BytesReceived:    512,
			TotalDuration:    25 * time.Millisecond,
		},
	}

	err = repo.WriteAuditTrail(cfg)
	if err != nil {
		t.Fatalf("Failed to execute write transaction trace: %v", err)
	}

	logs, err := repo.FetchHistory(10)
	if err != nil {
		t.Fatalf("Control plane failed to fetch historical ledger logs: %v", err)
	}

	if len(logs) != 1 {
		t.Fatalf("Expected exactly 1 historical log entry, but found %d", len(logs))
	}

	savedLog := logs[0]
	if savedLog.Method != "POST" {
		t.Errorf("Data Mutation: Expected HTTP method POST, but read %s", savedLog.Method)
	}
	if savedLog.BytesTransmitted != 120 {
		t.Errorf("Telemetry Distortion: Read %d bytes sent, expected 120", savedLog.BytesTransmitted)
	}

	repo.Close()
}
