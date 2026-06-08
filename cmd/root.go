/*
Copyright © 2026 Rahil Parikh <rahilparikh11@gmail.com>
*/
package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/parikhrahil/gcurl/pkg/config"
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

			return ExecuteRequest(cfg)
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

	// Sub commands
	rootCmd.AddCommand(NewVersionCommand())

	return rootCmd
}

func ExecuteRequest(cfg *config.RequestConfiguration) error {
	fmt.Printf("[gcurl Engine Ingress Success]\n")
	fmt.Printf("▸ Target Endpoint: %s %s\n", cfg.Method, cfg.URL)
	fmt.Printf("▸ Insecure TLS: %t | Verbose Logging: %t\n", cfg.Insecure, cfg.Verbose)
	if len(cfg.Headers) > 0 {
		fmt.Printf("▸ Captured Metadata Headers: %v\n", cfg.Headers)
	}
	return nil
}
