# ====================================================================================
# Configuration & Telemetry Variable Extraction
# ====================================================================================
MODULE       := github.com/parikhrahil/gcurl
BINARY_NAME  := gcurl
OUT_DIR      := bin

# Build metadata dynamically fetched from local git tree state
VERSION      ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "v1.1.0-dev")
GIT_COMMIT   := $(shell git rev-parse --short HEAD 2>/dev/null || echo "sha-not-set")
BUILD_TIME   := $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')

# Linker flags targeting variable injection maps within pkg/version
LDFLAGS      := -w -s \
                -X $(MODULE)/pkg/version.Version=$(VERSION) \
                -X $(MODULE)/pkg/version.GitCommit=$(GIT_COMMIT) \
                -X $(MODULE)/pkg/version.BuildTime=$(BUILD_TIME)

# Target packages filtering out entrypoint paths for coverage accuracy
TARGET_PKGS  := ./pkg/... ./cmd/...

.PHONY: help clean test cover build install cross-compile

# ====================================================================================
# Control Plane Targets
# ====================================================================================

help: ## Display a self-documenting catalog of available automation pipelines
	@echo "gcurl Automation Control Plane ($(VERSION))"
	@echo "========================================================================"
	@@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-18s\033[0m %s\n", $$1, $$2}'

clean: ## Purge local build artifacts, temporary directories, and test coverage maps
	@echo "--> Purging operational outputs and test ledgers..."
	@rm -rf $(OUT_DIR) coverage.out
	@go clean -testcache

test: ## Execute full offline testing matrix across filtered core packages verbose
	@echo "--> Running isolated structural test suite..."
	@go test -v -race $(TARGET_PKGS)

cover: ## Execute statement coverage counters and launch interactive HTML review map
	@echo "--> Evaluating statement coverage matrix paths..."
	@go test -covermode=count -coverprofile=coverage.out $(TARGET_PKGS)
	@go tool cover -func=coverage.out

build: clean ## Compile the native binary local workspace artifact with static ldflag injections
	@echo "--> Compiling local static binary: bin/$(BINARY_NAME) [$(VERSION)]..."
	@mkdir -p $(OUT_DIR)
	@CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/$(BINARY_NAME) main.go
	@echo "--> Complete. Binary compiled successfully."

install: ## Compile and provision binary into local path environment ($GOPATH/bin)
	@echo "--> Installing binary into local path profile storage..."
	@CGO_ENABLED=0 go install -ldflags "$(LDFLAGS)" ./...
	@echo "--> Complete. Execute 'gcurl version' to verify path linkage."

cross-compile: clean ## Matrix compile immutable binary payloads for targets (Linux/Darwin/Windows)
	@echo "--> Cross-compiling architectural matrix targets..."
	@# Linux x86_64 Core Target
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/$(BINARY_NAME)-linux-amd64 main.go
	@# Darwin (macOS Apple Silicon M-Series) Target
	@GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/$(BINARY_NAME)-darwin-arm64 main.go
	@# Windows x86_64 Shell Target
	@GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/$(BINARY_NAME)-windows-amd64.exe main.go
	@echo "--> Architectural cross-compilation pipeline matrix drained successfully."
