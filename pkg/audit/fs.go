package audit

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	DirName = ".gcurl"
	DBName  = "history.db"
)

type WorkspaceManager struct {
	DirPath  string
	DbPath   string
	Disabled bool
}

func BootstrapWorkspace() (*WorkspaceManager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// If the OS environment lacks a home definition, safely degrade rather than panic
		return &WorkspaceManager{Disabled: true},
			fmt.Errorf("unable to resolve user home directory %w", err)
	}

	targetDir := filepath.Join(homeDir, DirName)
	targetDb := filepath.Join(targetDir, DBName)

	mgr := &WorkspaceManager{
		DirPath:  targetDir,
		DbPath:   targetDb,
		Disabled: false,
	}

	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		err := os.MkdirAll(targetDir, 0o700)
		if err != nil {
			// Catch read-only environments (e.g., restricted CI/CD runners) and mark as disabled
			return &WorkspaceManager{Disabled: true}, fmt.Errorf("workspace permission denial on directory creation: %w", err)
		}
	}

	return mgr, nil
}
