// SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

// Package sql
package sql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	_ "github.com/duckdb/duckdb-go/v2"
	"github.com/OpenCHAMI/legendary-funicular/query/internal/config"
)

const QueryPlaceholderSources = "SOURCES"

func buildQuerySecret(cfg *config.Config) string {
	return fmt.Sprintf(`
CREATE OR REPLACE SECRET local_s3 (
  TYPE s3,
  PROVIDER config,
  KEY_ID '%s',
  SECRET '%s',
  REGION '%s',
  ENDPOINT '%s',
  USE_SSL %t,
  URL_STYLE 'path'
);`,
		*cfg.S3KeyAccess,
		*cfg.S3KeySecret,
		*cfg.S3Region,
		*cfg.S3Endpoint,
		cfg.S3SSL,
	)
}

type Engine struct {
	Pool *sql.DB
	cfg  *config.Config
}

func New(cfg *config.Config, ctx context.Context) (*Engine, error) {
	var db *sql.DB
	var err error
	query := buildQuerySecret(cfg)

	db, err = sql.Open("duckdb", "")
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(query)
	if err != nil {
		return nil, err
	}

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}
	return &Engine{Pool: db, cfg: cfg}, nil
}

func (e *Engine) Close() error {
	return e.Pool.Close()
}

// todo: after prototype completion, gut this entire routine into parts
// and spread the pieces to more appropriate places. the entire routine
// is now too domain specific. internal libs should have no knowledge
// of sources, streams, or really any concrete type info. it goes against
// the spirit of the original design of this subproject (favor simplicity
// and extendability). also, given the nature of this routine, it would
// fit in better in the cmd/query/query.go package.
//
// also some of the code below is just outright prototype quality garbage...
//
// todo: another candidate for a grammar/formal parser
// might be able to find an existing duckdb grammar online to save time...
func (e *Engine) prepareQuerystr(querystr string, sources []string) (string, error) {
	// remove trailing ; if it exists to prep for potential query nesting
	querystr = strings.TrimSpace(querystr)
	if len(querystr) > 0 && querystr[len(querystr)-1] == ';' {
		querystr = querystr[:len(querystr)-1]
	}

	// todo: if contains * or ,?\s*data
	s := strings.ToLower(querystr)
	pt := strings.Index(s, "from")
	if pt < 0 {
		return querystr, errors.New("invalid SQL query")
	}

	// although this mess allows us to ensure no double json encoding occurs on the target
	// unstructured data column (and implicitly enable pretty-print for it), it doesn't,
	// readily enable querying or specifying clauses w.r.t. the internal fields.
	//
	// we may be approaching the limit of hiding internal behaviors w.r.t. the
	// unstructured data and duckdb. users using the `sql` subcommand may need to learn
	// elements of the duckdb sql dialect if those behaviors are needed. for example:
	// `go run . -s events sql "select * from SOURCES where json_extract_string(cloudevent, '$.specversion') == '1.0'" | jq`

	// if strings.Contains(s, "*") {
	// 	querystr = strings.Replace(querystr, "*", "* exclude(data), json(data) as data", 1)
	// } else if strings.Contains(s, "data") {
	// 	querystr = strings.Replace(querystr, "data", "json(data) as data", 1)
	// }

	expansionField := "data" // default to syslog schema expectation
	for _, source := range sources {
		if strings.Contains(strings.ToLower(source), "events") {
			expansionField = "cloudevent"
		}
	}

	// todo: fix mess, refactor/revise
	s = s[:pt]
	if strings.Contains(s, "*") {
		querystr = strings.Replace(querystr, "*", fmt.Sprintf("* exclude(%s), json(%s) as %s", expansionField, expansionField, expansionField), 1)
	} else if strings.Contains(s, expansionField) {
		querystr = strings.Replace(querystr, expansionField, fmt.Sprintf("json(%s) as %s", expansionField, expansionField), 1)
	}

	if strings.Contains(querystr, QueryPlaceholderSources) {
		if sources == nil {
			return querystr, fmt.Errorf("SOURCES present but sources slice is nil")
		} else if len(sources) == 0 {
			return querystr, fmt.Errorf("SOURCES present but sources slice is empty")
		}
		for i, s := range sources {
			if s == "" {
				return querystr, fmt.Errorf("sources[%d] is an empty string", i)
			}
		}

		var replacement string
		if len(sources) == 1 {
			replacement = sources[0]
		} else {
			var subqueries []string
			for _, s := range sources {
				subqueries = append(subqueries, fmt.Sprintf("SELECT * FROM %s", s))
			}
			replacement = strings.Join(subqueries, " UNION ALL BY NAME ")
			replacement = fmt.Sprintf("(%s)", replacement)
		}
		querystr = strings.Replace(querystr, QueryPlaceholderSources, replacement, 1)
	}

	return querystr, nil
}

// Query executes a SQL query after resolving sources via prepareQuerystr.
// Behavior is controlled by requireJSON, which determines how results are
// materialized and consumed by the caller.
//
// If requireJSON is false, the query is executed as-is and returns native
// column types. This is the most efficient path and should be preferred
// when the schema is known and flat. If the schema is unstructured and
// nested, this pathway requires the caller to explicitly handle
// DuckDB-specific collection/structure types (e.g., MAP, STRUCT) during
// scanning.
//
// If requireJSON is true, the query is wrapped as:
//
//	SELECT to_json(t) FROM (<query>) t
//
// This forces each row into a single JSON object column, enabling
// schema-agnostic consumption (e.g., scan into map[string]any)
// and avoiding manual normalization of DuckDB composite types.
//
// Performance:
//   - The JSON path incurs additional overhead (~30% in practice) due to
//     per-row serialization in DuckDB and subsequent JSON decoding during
//     scanning in Go.
//   - The direct path avoids this cost and should be used in performance-
//     sensitive or high-throughput scenarios.
//
// Use requireJSON for arbitrary/user-defined queries or heterogeneous inputs
// where schema cannot be assumed. Otherwise, prefer the direct mode.
//
// Returns a *sql.Rows iterator over the result set.
func (e *Engine) Query(querystr string, sources []string, requireJSON bool) (*sql.Rows, error) {
	querystr, err := e.prepareQuerystr(querystr, sources)
	if err != nil {
		return nil, err
	}

	if requireJSON {
		querystr = fmt.Sprintf("SELECT to_json(t) FROM (%s) t", querystr)
	}

	rows, err := e.Pool.Query(querystr)
	if err != nil {
		return nil, err
	}
	return rows, err
}

// ScanRowDirectToMap scans the current row into a map by performing a
// column-wise scan using database/sql primitives.
//
// Intended for use with direct query execution where schema is known or
// controlled and performance is critical.
func ScanRowDirectToMap(rows *sql.Rows) (map[string]any, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	values := make([]any, len(cols))
	ptrs := make([]any, len(cols))
	for i := range values {
		ptrs[i] = &values[i]
	}

	if err := rows.Scan(ptrs...); err != nil {
		return nil, err
	}

	row := make(map[string]any, len(cols))
	for i, c := range cols {
		row[c] = values[i]
	}

	return row, nil
}

// ScanRowJSONToMap scans the current row into a map assuming the result set
// contains a single JSON column (e.g., produced via `SELECT to_json(t)`).
//
// Intended for use with flexible query execution where schema is unknown,
// heterogeneous, or not worth handling explicitly.
func ScanRowJSONToMap(rows *sql.Rows) (map[string]any, error) {
	m := map[string]any{}
	err := rows.Scan(&m)
	return m, err
}
