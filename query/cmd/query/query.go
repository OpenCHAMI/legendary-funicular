// SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

// Package query
package query

import (
	"context"
	dbsql "database/sql"

	"github.com/OpenCHAMI/legendary-funicular/query/cmd/opts"
	"github.com/OpenCHAMI/legendary-funicular/query/internal/config"
	"github.com/OpenCHAMI/legendary-funicular/query/internal/render"
	"github.com/OpenCHAMI/legendary-funicular/query/internal/sql"
)

func ExecUnstructured(
	querystr string,
	ctx context.Context,
	cfg *config.Config,
	options opts.Opts,
) error {
	sources := options.BuildSources(cfg)
	engine, err := sql.New(cfg, ctx)
	if err != nil {
		return err
	}
	defer engine.Close()

	enc, err := render.BuildEncoder(options.Output, options.Format)
	if err != nil {
		return err
	}
	defer enc.Close()

	rows, err := engine.Query(querystr, sources)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		record, err := sql.ScanRowToMap(rows)
		if err != nil {
			return err
		}
		if err := enc.Encode(record); err != nil {
			return err
		}
	}

	return nil
}

type SQLRowScanner[T any] func(rows *dbsql.Rows) (T, error)

func ExecStructured[T any](
	querystr string,
	ctx context.Context,
	cfg *config.Config,
	options opts.Opts,
	scanner SQLRowScanner[T],
) error {

	sources := options.BuildSources(cfg)
	engine, err := sql.New(cfg, ctx)
	if err != nil {
		return err
	}
	defer engine.Close()

	enc, err := render.BuildEncoder(options.Output, options.Format)
	if err != nil {
		return err
	}
	defer enc.Close()

	rows, err := engine.Query(querystr, sources)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		record, err := scanner(rows)
		if err != nil {
			return err
		}
		if err := enc.Encode(record); err != nil {
			return err
		}
	}

	return nil
}

type RecordSchemaInspection struct {
	ColumnName string `json:"column_name"`
	ColumnType string `json:"column_type"`
	IsNullable string `json:"is_nullable"`
}

func ScanRecordSchemaInspection(rows *dbsql.Rows) (RecordSchemaInspection, error) {
	var record RecordSchemaInspection
	err := rows.Scan(
		&record.ColumnName,
		&record.ColumnType,
		&record.IsNullable,
	)
	return record, err
}
