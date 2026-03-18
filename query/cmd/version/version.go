// SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT
//
// SPDX-FileNotice: This file is derived from:
// https://github.com/OpenCHAMI/ochami/blob/2247677c4e2a944e230872b5024100078dd8efec/cmd/version/version.go
// Changes: adjusted import paths for local module usage; minor adjustments.

// Package version
package version

import (
	"fmt"
	"runtime"

	"github.com/OpenCHAMI/legendary-funicular/query/internal/version"
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "version",
		Short: "Print detailed version to stdout and exit",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Version:    %s\n", version.Version)
			fmt.Printf("Tag:        %s\n", version.Tag)
			fmt.Printf("Branch:     %s\n", version.Branch)
			fmt.Printf("Commit:     %s\n", version.Commit)
			fmt.Printf("Git State:  %s\n", version.GitState)
			fmt.Printf("Date:       %s\n", version.Date)
			fmt.Printf("Go:         %s\n", version.GoVersion)
			fmt.Printf("Compiler:   %s\n", runtime.Compiler)
			fmt.Printf("Build Host: %s\n", version.BuildHost)
			fmt.Printf("Build User: %s\n", version.BuildUser)
		},
	}
	return cmd
}
