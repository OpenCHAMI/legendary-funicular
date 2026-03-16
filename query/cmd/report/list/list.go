// SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

// Package list
package list

import (
	"slices"
	"strings"

	"github.com/seantronsen/openchami-logq/query/cmd/opts"
	"github.com/seantronsen/openchami-logq/query/internal/render"
	"github.com/seantronsen/openchami-logq/query/internal/reports"
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "list",
		Short: "List available reports",
		Long:  "List all available predefined reports along with a brief description to aid discovery.",
		RunE: func(cmd *cobra.Command, args []string) error {
			options := opts.FromCobraCmd(cmd)
			return run(options)
		},
	}

	return cmd
}

type Record struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func run(options opts.Opts) error {

	items := append([]reports.Report(nil), reports.Registry...)
	slices.SortFunc(items, func(a, b reports.Report) int {
		return strings.Compare(a.Name(), b.Name())
	})

	enc, err := render.BuildEncoder(options.Output, options.Format)
	if err != nil {
		return err
	}
	defer enc.Close()

	for _, r := range items {
		rec := Record{Name: r.Name(), Description: r.Description()}
		if err := enc.Encode(rec); err != nil {
			return err
		}
	}
	return nil

}
