// SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

package reports

import (
	"github.com/seantronsen/openchami-logq/query/internal/render"
	"github.com/seantronsen/openchami-logq/query/internal/sql"
)

type ReportFindServiceErrors struct{}

func (r ReportFindServiceErrors) Name() string { return "find-all-service-errors" }
func (r ReportFindServiceErrors) Description() string {
	return "Query all log records where the service reported an error-level condition (level = 'error' or 'err') to assist in failure triage and incident investigation."

}
func (r ReportFindServiceErrors) BuildQueryString() string {
	return `
SELECT *
FROM SOURCES
WHERE level = 'error' OR level = 'err'
ORDER BY ts DESC`
}

func (r ReportFindServiceErrors) QueryRecords(engine *sql.Engine, sources []string, enc render.Encoder) error {
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
