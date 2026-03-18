// SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

// Package inspect
package inspect

import (
	cmdConfig "github.com/OpenCHAMI/legendary-funicular/query/cmd/inspect/config"
	cmdDates "github.com/OpenCHAMI/legendary-funicular/query/cmd/inspect/dates"
	cmdSchema "github.com/OpenCHAMI/legendary-funicular/query/cmd/inspect/schema"
	"github.com/OpenCHAMI/legendary-funicular/query/internal/config"
	"github.com/spf13/cobra"
)

func NewCmd(cfg *config.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "inspect",
		Short: "Inspect dataset metadata",
		Long: `Inspect metadata about datasets stored in the log lake, including
available date partitions and schema information.`,
	}

	cmd.AddCommand(
		cmdConfig.NewCmd(cfg),
		cmdDates.NewCmd(cfg),
		cmdSchema.NewCmd(cfg),
	)

	return cmd
}
