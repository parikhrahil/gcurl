package cmd

import (
	"fmt"

	"github.com/parikhrahil/gcurl/pkg/version"
	"github.com/spf13/cobra"
)

func NewVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print extended system telemetry and compilation version information",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.SilenceUsage = true
			info := version.GetInfo()
			fmt.Fprintln(cmd.OutOrStdout(), info.String())
		},
	}
}
