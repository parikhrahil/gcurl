package version

import (
	"fmt"
	"runtime"
	"runtime/debug"
)

// Global constants populated at compilation time via: go build -ldflags "-X ..."
var (
	Version   = "unknown"
	GitCommit = "sha-not-set"
	BuildTime = "time-not-set"
)

// BuildMetadata encapsulates the total runtime and compilation telemetry matrix.
type BuildMetadata struct {
	Version   string `json:"version"`
	GitCommit string `json:"git_commit"`
	BuildTime string `json:"build_time"`
	GoVersion string `json:"go_version"`
	OS        string `json:"os"`
	Arch      string `json:"architecture"`
}

// GetInfo computes and returns a hydrated metadata summary model.
func GetInfo() BuildMetadata {
	meta := BuildMetadata{
		Version:   Version,
		GitCommit: GitCommit,
		BuildTime: BuildTime,
		GoVersion: runtime.Version(),
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	}

	// Fallback to runtime reflection if compile-time flags were omitted
	if meta.Version == "unknown" {
		if info, ok := debug.ReadBuildInfo(); ok {
			if info.Main.Version != "" && info.Main.Version != "(devel)" {
				meta.Version = info.Main.Version
			}
		}
	}

	return meta
}

// String normalizes our layout format for standard terminal views.
func (b BuildMetadata) String() string {
	return fmt.Sprintf("gcurl version %s (Git: %s, Compiled: %s, Runtime: %s on %s/%s)",
		b.Version, b.GitCommit, b.BuildTime, b.GoVersion, b.OS, b.Arch)
}
