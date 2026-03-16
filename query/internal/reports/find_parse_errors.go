// SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

package reports

import (
	"github.com/seantronsen/openchami-logq/query/internal/render"
	"github.com/seantronsen/openchami-logq/query/internal/sql"
)

type ReportFindParseErrors struct{}

func (r ReportFindParseErrors) Name() string { return "find-all-parse-errors" }
func (r ReportFindParseErrors) Description() string {
	return "List all records where parsing failed (parse_error is non-null) to help identify malformed log lines and ingestion issues during the collection phase."
}
func (r ReportFindParseErrors) BuildQueryString() string {
	return "SELECT * FROM SOURCES WHERE parse_error is not null"
}

func (r ReportFindParseErrors) QueryRecords(engine *sql.Engine, sources []string, enc render.Encoder) error {
	// todo: this likely will have a lot... of repeated logic.
	// if we do a V2, we should abstract this out further.
	querystr := r.BuildQueryString()
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

	if err := rows.Err(); err != nil {
		return err
	}

	return nil
}
