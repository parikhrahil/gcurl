package audit_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/parikhrahil/gcurl/pkg/audit"
)

func TestBootstrapWorkspace_WritesAllowed(t *testing.T) {
	tmpDir := t.TempDir()
	envVar := "HOME"
	windowsEnvVar := "USERPROFILE"

	// Mock the home env var for this test
	t.Setenv(envVar, tmpDir)
	t.Setenv(windowsEnvVar, tmpDir)

	targetDir := filepath.Join(tmpDir, ".gcurl")
	dbPath := filepath.Join(targetDir, "history.db")

	wsMgr, err := audit.BootstrapWorkspace()
	if err != nil {
		t.Fatalf("Expected successful workspace bootstrap, but encountered an error: %v", err)
	}

	if wsMgr.Disabled {
		t.Error("Structural Failure: Workspace manager was incorrectly marked as Disabled in a healthy writable zone")
	}

	if wsMgr.DbPath != dbPath {
		t.Errorf("Path Alignment Error: Expected database path to anchor to %s, but got %s",
			dbPath, wsMgr.DbPath)
	}

	info, err := os.Stat(targetDir)
	if os.IsNotExist(err) {
		t.Fatal("Validation Error: The target workspace directory failed to materialize on disk")
	}
	if !info.IsDir() {
		t.Error("Type Error: Materialized workspace target path is not a valid directory wrapper")
	}
}

func TestBootstrapWorkspace_ReadOnlyFileSystem(t *testing.T) {
	// Architectural Guardrail: Windows enforces filesystem permissions via completely distinct access-control
	// lists (ACLs). We restrict this specific low-level bitmask test to Unix/Linux/Darwin subsystems.
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Unix bitmask write-block simulation on Windows target architecture")
	}

	tmpDir := t.TempDir()
	envVar := "HOME"

	t.Setenv(envVar, tmpDir)
	err := os.Chmod(tmpDir, 0o500)
	if err != nil {
		t.Fatalf("Failed to apply secure permission restriction mask: %v", err)
	}
	defer func() {
		_ = os.Chmod(tmpDir, 0o755)
	}()

	wsMgr, err := audit.BootstrapWorkspace()
	if err != nil {
		t.Logf("Informational: Ingress pipeline caught expected write restriction error: %v", err)
	}

	if !wsMgr.Disabled {
		t.Error("Defensive Invariant Failure: Workspace manager remained Enabled despite operating inside a write-blocked storage block")
	}
}
