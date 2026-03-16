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
	"github.com/seantronsen/openchami-logq/query/internal/config"
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

func (e *Engine) Query(querystr string, sources []string) (*sql.Rows, error) {
	// todo: sources should include the path to the tmp parquet store once
	// that feature is built. recall this store exists to avoid one-off
	// premature parsing of non-compacted ndjson records (by creating a
	// temporary compaction early). ideally, this will create a re-usable
	// tmp compaction store where new new ndjson logs captured since the
	// last query can be appended as opposed to recreating the entire
	// compaction on each query.
	if strings.Contains(querystr, QueryPlaceholderSources) {
		if sources == nil {
			return nil, fmt.Errorf("SOURCES present but sources slice is nil")
		} else if len(sources) == 0 {
			return nil, fmt.Errorf("SOURCES present but sources slice is empty")
		}
		for i, s := range sources {
			if s == "" {
				return nil, fmt.Errorf("sources[%d] is an empty string", i)
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

	rows, err := e.Pool.Query(querystr)
	if err != nil {
		return nil, err
	}
	return rows, err
}

func (e *Engine) Close() error {
	return e.Pool.Close()
}

func ScanRowToMap(rows *sql.Rows) (map[string]any, error) {
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
