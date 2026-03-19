// SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

// Package sql
package sql

import (
	"context"
	"database/sql"
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

func (e *Engine) prepareQuerystr(querystr string, sources []string) (string, error) {
	// remove trailing ; if it exists to prep for potential query nesting
	querystr = strings.TrimSpace(querystr)
	if len(querystr) > 0 && querystr[len(querystr)-1] == ';' {
		querystr = querystr[:len(querystr)-1]
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

		sourcesQuoted := make([]string, len(sources))
		for i, s := range sources {
			sourcesQuoted[i] = fmt.Sprintf("'%s'", s)
		}
		replacement := fmt.Sprintf(
			"read_parquet([%s], union_by_name = true )",
			strings.Join(sourcesQuoted, ", "),
		)
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
