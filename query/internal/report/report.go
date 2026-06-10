// SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

// Package report
package report

import (
	"fmt"
	"log/slog"

	"github.com/OpenCHAMI/legendary-funicular/query/internal/dev"
)

type ParamSpec struct {
	Name        string
	Description string
	Required    bool
	Default     any
}

type Report interface {
	Name() string
	Description() string
	ParamSpecs() []ParamSpec
	SampleParams() map[string]any
	BuildQueryString(params map[string]any) (string, error)
}

func BuildSchemaQueryStr(r Report) (querystr string, err error) {
	params := r.SampleParams()
	querystr, err = r.BuildQueryString(params)
	if err != nil {
		return querystr, err
	}
	querystr = fmt.Sprintf(`
SELECT
	column_name,
	column_type,
	"null" AS is_nullable
FROM (
	DESCRIBE %s
);`, querystr)
	return querystr, err
}

// ParseRawParams parses a slice of raw string arguments into a map of report params.
// Args are expected in key=value format (e.g., "id=123" "status=active").
//
// The returned map is compatible with Report.BuildQueryStr and ValidateParams.
//
// NOTE: Full key=value parsing is not yet implemented and will require a proper
// grammar. For now, any non-empty args slice returns error NotImplemented.
// TODO: implement key=value parsing. considerations include:
//   - validate each element is proper key=value format, fail if not
//   - can't assume each slice element is a simple pair (e.g., value could be
//     a quoted sentence, a list, or a conditional expression)
//   - may require a grammar/lexer rather than naive strings.Split("=")
//   - likely will require escape characters (e.g., "\=") when special
//     characters occur in the value component.
func ParseRawParams(raw []string) (map[string]any, error) {
	slog.Warn("ParseRawParams still requires a proper grammar definition for safe argument parsing and will fail if provided a non-empty slice")
	params := make(map[string]any)
	if len(raw) != 0 {
		return params, dev.NotImplemented()
	}

	return params, nil
}
