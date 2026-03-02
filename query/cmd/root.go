// Package cmd
package cmd

import (
	"os"

	"github.com/seantronsen/openchami-logq/query/cmd/dump"
	"github.com/seantronsen/openchami-logq/query/cmd/inspect"
	"github.com/seantronsen/openchami-logq/query/cmd/report"
	"github.com/seantronsen/openchami-logq/query/cmd/sql"
	"github.com/seantronsen/openchami-logq/query/cmd/version"
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:   "openchami-logq",
		Short: "Query OpenCHAMI log lake data",
		Long: `openchami-logq is a CLI for querying OpenCHAMI log lake data stored in S3 using
DuckDB. It supports ad-hoc SQL queries, predefined reports, and dataset
inspection using serverless technologies.`,
	}
	return rootCmd
}

func Execute() {
	rootCmd := NewRootCmd()
	rootCmd.AddCommand(
		dump.NewCmd(),
		inspect.NewCmd(),
		report.NewCmd(),
		sql.NewCmd(),
		version.NewCmd(),
	)

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
