// Package report
package report

import (
	"github.com/seantronsen/openchami-logq/query/cmd/report/describe"
	"github.com/seantronsen/openchami-logq/query/cmd/report/list"
	"github.com/seantronsen/openchami-logq/query/cmd/report/run"
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "report",
		Short: "Run predefined reports",
		Long: `Execute predefined reports for common OpenCHAMI analysis
and troubleshooting tasks. Each report encapsulates all raw query logic to expose a
simplified interface for querying logs and events.`,
	}
	cmd.AddCommand(
		describe.NewCmd(),
		list.NewCmd(),
		run.NewCmd(),
	)

	return cmd
}
