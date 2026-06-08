package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/parikhrahil/gcurl/pkg/audit"
	"github.com/parikhrahil/gcurl/pkg/config"
	"github.com/parikhrahil/gcurl/pkg/report"
	"github.com/parikhrahil/gcurl/pkg/transport"
	"github.com/spf13/cobra"
)

var (
	customMethod  string
	customHeaders []string
	rawData       string
	insecureTLS   bool
	verboseMode   bool
)

func NewRootCommand() *cobra.Command {
	cfg := config.NewDefaultConfig()

	rootCmd := &cobra.Command{
		Use:   "gcurl [url]",
		Short: "gcurl is an enterprise-grade, concurrent cURL replica engineered in Go.",
		Long: `
			A highly scannable, open-source command line tool mapped directly to POSIX 
			standard network specifications.
		`,
		Args: cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			cfg.URL = args[0]
			if !strings.HasPrefix(cfg.URL, "http://") && !strings.HasPrefix(cfg.URL, "https://") {
				return errors.New("invalid protocol: target URL must begin with 'http://' or 'https://'")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			if customMethod != "" {
				cfg.Method = strings.ToUpper(customMethod)
			}

			cfg.Insecure = insecureTLS
			cfg.Verbose = verboseMode
			cfg.Data = rawData

			for _, h := range customHeaders {
				parts := strings.SplitN(h, ":", 2)
				if len(parts) != 2 {
					return fmt.Errorf("malformed header input: %s. Expected 'Key: Value'", h)
				}
				cfg.Headers.Add(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
			}

			if cfg.Concurrency > cfg.TotalRequests {
				fmt.Fprintf(cmd.ErrOrStderr(), "Auto-adjusting concurrency pool to %d", cfg.TotalRequests)
				cfg.Concurrency = cfg.TotalRequests // prevent over provisioning of resources
			}

			return ExecuteRequest(cmd, cfg)
		},
	}

	// Flags
	rootCmd.Flags().StringVarP(&customMethod, "request", "X", "",
		"Specify request command to use (e.g., GET, POST, PUT)")
	rootCmd.Flags().StringSliceVarP(&customHeaders, "header", "H", []string{},
		"Pass custom header to server (e.g., 'Content-Type: application/json')")
	rootCmd.Flags().StringVarP(&rawData, "data", "d", "", "HTTP POST data payload string")
	rootCmd.Flags().BoolVarP(&insecureTLS, "insecure", "k", false,
		"Allow insecure server connections when using SSL/TLS")
	rootCmd.Flags().BoolVarP(&verboseMode, "verbose", "v", false,
		"Make the operation talkative (expose connection trace details)")
	rootCmd.Flags().IntVarP(&cfg.Concurrency, "concurrency", "c", 1,
		"Number of concurrent worker goroutines")
	rootCmd.Flags().IntVarP(&cfg.TotalRequests, "requests", "n", 1,
		"Total number of HTTP requests to execute")

	// Sub commands
	rootCmd.AddCommand(NewVersionCommand())
	rootCmd.AddCommand(NewHistoryCommand())

	return rootCmd
}

func ExecuteRequest(cmd *cobra.Command, cfg *config.RequestConfiguration) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if cfg.TotalRequests > 1 || cfg.Concurrency > 1 {
		if cfg.Verbose {
			fmt.Fprintf(cmd.ErrOrStderr(), "* Shifting into Benchmarking Mode."+
				"Worker count: %d. Total target load: %d\n", cfg.Concurrency, cfg.TotalRequests)
		}

		engine := transport.NewParallelEngine(cfg)

		wallclockStart := time.Now()
		results, globalMetrics := engine.Execute(ctx)
		totalWallclockTime := time.Since(wallclockStart)

		cfg.Metrics.BytesTransmitted = globalMetrics.BytesTransmitted
		cfg.Metrics.BytesReceived = globalMetrics.BytesReceived
		cfg.Metrics.TotalDuration = totalWallclockTime

		wsMgr, err := audit.BootstrapWorkspace()
		if err == nil && !wsMgr.Disabled {
			if repo, dbErr := audit.NewHistoryRepository(wsMgr.DbPath); dbErr == nil {
				_ = repo.WriteAuditTrail(cfg) // Execution failure safely guarded internally
				repo.Close()
			}
		}
		report.RenderSummary(results, totalWallclockTime, globalMetrics)
		return nil
	}

	var repo *audit.HistoryRepository
	dbEnabled := false

	wsMgr, err := audit.BootstrapWorkspace()
	if err == nil && !wsMgr.Disabled {
		historyRepo, dbErr := audit.NewHistoryRepository(wsMgr.DbPath)
		if dbErr == nil {
			repo = historyRepo
			dbEnabled = true
			defer repo.Close()
		} else if cfg.Verbose {
			fmt.Fprintf(cmd.ErrOrStderr(), "* Telemetry warning: storage initialization bypassed: %v\n", dbErr)
		}
	}

	var bodyReader io.Reader
	if cfg.Data != "" {
		bodyReader = audit.NewAuditReader(strings.NewReader(cfg.Data), &cfg.Metrics.BytesTransmitted)
	}

	req, err := http.NewRequestWithContext(ctx, cfg.Method, cfg.URL, bodyReader)
	if err != nil {
		return err
	}

	for k, v := range cfg.Headers {
		for _, val := range v {
			req.Header.Add(k, val)
		}
	}

	client := transport.NewHTTPClient(cfg)

	startTime := time.Now()
	res, err := client.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if cfg.Verbose {
		fmt.Fprintf(cmd.ErrOrStderr(), "* Connection established. Status Protocol: %s\n", res.Proto)
		for k, v := range res.Header {
			fmt.Fprintf(cmd.ErrOrStderr(), "< %s: %s\n", k, strings.Join(v, ", "))
		}
	}

	trackedRespBody := audit.NewAuditReader(res.Body, &cfg.Metrics.BytesReceived)
	_, err = io.Copy(cmd.OutOrStdout(), trackedRespBody)
	if err != nil {
		return fmt.Errorf("failed to flush streaming response network buffer: %w", err)
	}

	cfg.Metrics.TotalDuration = time.Since(startTime)
	if dbEnabled && repo != nil {
		writeErr := repo.WriteAuditTrail(cfg)
		if writeErr != nil && cfg.Verbose {
			fmt.Fprintf(cmd.ErrOrStderr(), "* Telemetry warning: audit write failure: %v\n", writeErr)
		}
	}
	return nil
}
