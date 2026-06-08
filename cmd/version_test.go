package cmd_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/parikhrahil/gcurl/cmd"
	"github.com/parikhrahil/gcurl/pkg/version"
)

func TestVersionCommand_ExecutionContract(t *testing.T) {
	versionCmd := cmd.NewVersionCommand()

	stdoutBuf := new(bytes.Buffer)
	versionCmd.SetOut(stdoutBuf)

	oldVersion := version.Version
	defer func() { version.Version = oldVersion }()
	version.Version = "1.0.0-beta"

	err := versionCmd.Execute()
	if err != nil {
		t.Fatalf("Execution Barrier: Command failed to execute cleanly: %v", err)
	}

	output := stdoutBuf.String()
	expectedToken := "gcurl version 1.0.0-beta"

	if !strings.Contains(output, expectedToken) {
		t.Errorf("Interface Contract Breach: Expected output stream to contain target token [%s], but captured:\n%s",
			expectedToken, output)
	}
}
