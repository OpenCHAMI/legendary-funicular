// SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

// Package registry
package registry

import (
	"fmt"
	"slices"

	"github.com/OpenCHAMI/legendary-funicular/query/internal/report"
)

func New() []report.Report {
	return []report.Report{
		&ReportFindParseErrors{},
		&ReportFindServiceErrors{},
	}
}

func Get(name string) (report.Report, error) {
	registry := New()
	criterion := func(r report.Report) bool {
		return r.Name() == name
	}

	idx := slices.IndexFunc(registry, criterion)
	if idx < 0 {
		return nil, fmt.Errorf("unknown report: %s", name)
	}

	return registry[idx], nil
}
