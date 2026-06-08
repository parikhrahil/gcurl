package cmd

import (
	"fmt"
	"text/tabwriter"
	"time"

	"github.com/parikhrahil/gcurl/pkg/audit"
	"github.com/spf13/cobra"
)

func NewHistoryCommand() *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "history",
		Short: "Display historical audit trails and transport metrics captured by gcurl",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			wsMgr, err := audit.BootstrapWorkspace()
			if err != nil || wsMgr.Disabled {
				return fmt.Errorf("unable to access historical workspace path: %v", err)
			}

			repo, err := audit.NewHistoryRepository(wsMgr.DbPath)
			if err != nil {
				return fmt.Errorf("failed to establish connection to historical data layer: %w", err)
			}
			defer repo.Close()

			logs, err := repo.FetchHistory(limit)
			if err != nil {
				return err
			}

			if len(logs) == 0 {
				fmt.Println("No audit records found in the localized system ledger.")
				return nil
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 3, ' ', 0)
			fmt.Fprintln(w, "ID\tMETHOD\tTARGET URL\tSTATUS\tSENT\tRCVD\tLATENCY\tEXECUTED (UTC)")
			fmt.Fprintln(w, "--\t------\t----------\t------\t----\t----\t------\t--------------")

			for _, log := range logs {
				statusStr := fmt.Sprintf("%d", log.StatusCode)
				if log.StatusCode == 0 {
					statusStr = "BATCH/ERR" // Identifies multi-request benchmarking or network errors
				}

				fmt.Fprintf(
					w, "%d\t%s\t%s\t%s\t%d B\t%d B\t%v\t%s\n",
					log.ID,
					log.Method,
					log.URL,
					statusStr,
					log.BytesTransmitted,
					log.BytesReceived,
					log.Duration.Round(time.Millisecond),
					log.ExecutedAt.Format("2006-01-02 15:04:05"),
				)
			}
			w.Flush()
			return nil
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "l", 20, "Limit the number of rendered history entries")

	return cmd
}
