// SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

package registry

import "github.com/OpenCHAMI/legendary-funicular/query/internal/report"

// //////////////////////////////////////////////////////////////////////////////
// REPORT: FIND PARSE ERRORS
// //////////////////////////////////////////////////////////////////////////////

type ReportFindParseErrors struct{}

func (r ReportFindParseErrors) Name() string { return "find-all-parse-errors" }
func (r ReportFindParseErrors) Description() string {
	return "List all records where parsing failed (parse_error is non-null) to help identify malformed log lines and ingestion issues during the collection phase."
}
func (r ReportFindParseErrors) ParamSpecs() []report.ParamSpec { return make([]report.ParamSpec, 0) }
func (r ReportFindParseErrors) SampleParams() map[string]any   { return make(map[string]any) }
func (r ReportFindParseErrors) BuildQueryString(_ map[string]any) (string, error) {
	return "SELECT * FROM SOURCES WHERE parse_error is not null", nil
}

// //////////////////////////////////////////////////////////////////////////////
// REPORT: FIND SERVICE ERRORS
// //////////////////////////////////////////////////////////////////////////////

type ReportFindServiceErrors struct{}

func (r ReportFindServiceErrors) Name() string { return "find-all-service-errors" }
func (r ReportFindServiceErrors) Description() string {
	return "Query all log records where the service reported an error-level condition (level = 'error' or 'err') to assist in failure triage and incident investigation."

}
func (r ReportFindServiceErrors) ParamSpecs() []report.ParamSpec { return make([]report.ParamSpec, 0) }
func (r ReportFindServiceErrors) SampleParams() map[string]any   { return make(map[string]any) }
func (r ReportFindServiceErrors) BuildQueryString(_ map[string]any) (string, error) {
	return `
SELECT *
FROM SOURCES
WHERE level = 'error' OR level = 'err'
ORDER BY ts DESC`, nil
}
