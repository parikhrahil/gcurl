package version_test

import (
	"strings"
	"testing"

	"github.com/parikhrahil/gcurl/pkg/version"
)

func TestGetInfo_WithInjectedBuildMetadata(t *testing.T) {
	oldVersion := version.Version
	oldCommit := version.GitCommit
	oldBuildTime := version.BuildTime

	defer func() {
		version.Version = oldVersion
		version.GitCommit = oldCommit
		version.BuildTime = oldBuildTime
	}()

	version.Version = "2.1.0"
	version.GitCommit = "9cf8e1b"
	version.BuildTime = "2026-06-09_00:00:00"

	// Step 4: Execute the hydration engine
	meta := version.GetInfo()

	// Step 5: Enforce domain contract invariants
	if meta.Version != "2.1.0" {
		t.Errorf("Contract Violation: Expected version '2.1.0', but resolved: %s", meta.Version)
	}
	if meta.GitCommit != "9cf8e1b" {
		t.Errorf("Telemetry Deviation: Expected commit SHA '9cf8e1b', but resolved: %s", meta.GitCommit)
	}

	// Validate the structural string formatter output
	displayString := meta.String()
	if !strings.Contains(displayString, "gcurl version 2.1.0") || !strings.Contains(displayString, "Git: 9cf8e1b") {
		t.Errorf("Presentation Corruption: Rendered layout fails contract validation. Got:\n%s", displayString)
	}
}

func TestGetInfo_FallbackWhenVariablesAreOmitted(t *testing.T) {
	oldVersion := version.Version
	defer func() { version.Version = oldVersion }()

	// Force an uninitialized state to trigger our runtime reflection fallback path
	version.Version = "unknown"

	meta := version.GetInfo()

	// The system should degrade gracefully to 'unknown' or extract from runtime.BuildInfo
	if meta.Version == "" {
		t.Error("Defensive Guard Broken: Hydration engine returned a blank version string under uninitialized parameters")
	}
}
