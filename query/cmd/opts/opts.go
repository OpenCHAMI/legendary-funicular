// Package opts
package opts

import (
	"io"
	"os"

	"github.com/seantronsen/openchami-logq/query/internal/utils"
	"github.com/spf13/cobra"
)

type Opts struct {
	Format string
	Output io.Writer
}

func FromCobraCmd(cmd *cobra.Command) Opts {
	format, err := cmd.Root().PersistentFlags().GetString("format")
	utils.CheckFatal(err)

	output, err := cmd.Root().PersistentFlags().GetString("output")
	utils.CheckFatal(err)

	var w io.Writer = os.Stdout
	if output != "" {
		f, err := os.Create(output)
		utils.CheckFatal(err)
		w = f
	}

	return Opts{Format: format, Output: w}
}
