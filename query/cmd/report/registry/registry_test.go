// SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

package registry

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNew tests report registry creation
func TestNew(t *testing.T) {
	registry := New()

	assert.NotNil(t, registry)
	assert.Len(t, registry, 2, "Should have 2 reports registered")

	// Verify both reports are present
	names := make([]string, len(registry))
	for i, r := range registry {
		names[i] = r.Name()
	}

	assert.Contains(t, names, "find-all-parse-errors")
	assert.Contains(t, names, "find-all-service-errors")
}

// TestGet_ValidReport tests retrieving a valid report
func TestGet_ValidReport(t *testing.T) {
	tests := []struct {
		name         string
		reportName   string
		wantName     string
		wantQueryStr string
	}{
		{
			name:         "find parse errors",
			reportName:   "find-all-parse-errors",
			wantName:     "find-all-parse-errors",
			wantQueryStr: "SELECT * FROM SOURCES WHERE parse_error is not null",
		},
		{
			name:         "find service errors",
			reportName:   "find-all-service-errors",
			wantName:     "find-all-service-errors",
			wantQueryStr: "level = 'error' OR level = 'err'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report, err := Get(tt.reportName)

			require.NoError(t, err)
			assert.NotNil(t, report)
			assert.Equal(t, tt.wantName, report.Name())

			// Verify query string
			query, err := report.BuildQueryString(nil)
			require.NoError(t, err)
			assert.Contains(t, query, tt.wantQueryStr)
		})
	}
}

// TestGet_InvalidReport tests retrieving an invalid report
func TestGet_InvalidReport(t *testing.T) {
	report, err := Get("non-existent-report")

	assert.Error(t, err)
	assert.Nil(t, report)
	assert.Contains(t, err.Error(), "unknown report: non-existent-report")
}

// TestGet_EmptyName tests retrieving with empty name
func TestGet_EmptyName(t *testing.T) {
	report, err := Get("")

	assert.Error(t, err)
	assert.Nil(t, report)
	assert.Contains(t, err.Error(), "unknown report")
}

// TestReportFindParseErrors tests parse errors report
func TestReportFindParseErrors(t *testing.T) {
	report := &ReportFindParseErrors{}

	// Test Name
	assert.Equal(t, "find-all-parse-errors", report.Name())

	// Test Description
	desc := report.Description()
	assert.NotEmpty(t, desc)
	assert.Contains(t, desc, "parse_error")
	assert.Contains(t, desc, "malformed")

	// Test ParamSpecs (should be empty)
	params := report.ParamSpecs()
	assert.NotNil(t, params)
	assert.Len(t, params, 0)

	// Test SampleParams (should be empty)
	samples := report.SampleParams()
	assert.NotNil(t, samples)
	assert.Len(t, samples, 0)

	// Test BuildQueryString
	query, err := report.BuildQueryString(nil)
	require.NoError(t, err)
	assert.Equal(t, "SELECT * FROM SOURCES WHERE parse_error is not null", query)

	// Should work with any params (ignored)
	query, err = report.BuildQueryString(map[string]any{"ignored": "value"})
	require.NoError(t, err)
	assert.Equal(t, "SELECT * FROM SOURCES WHERE parse_error is not null", query)
}

// TestReportFindServiceErrors tests service errors report
func TestReportFindServiceErrors(t *testing.T) {
	report := &ReportFindServiceErrors{}

	// Test Name
	assert.Equal(t, "find-all-service-errors", report.Name())

	// Test Description
	desc := report.Description()
	assert.NotEmpty(t, desc)
	assert.Contains(t, desc, "error-level")
	assert.Contains(t, desc, "level = 'error'")

	// Test ParamSpecs (should be empty)
	params := report.ParamSpecs()
	assert.NotNil(t, params)
	assert.Len(t, params, 0)

	// Test SampleParams (should be empty)
	samples := report.SampleParams()
	assert.NotNil(t, samples)
	assert.Len(t, samples, 0)

	// Test BuildQueryString
	query, err := report.BuildQueryString(nil)
	require.NoError(t, err)
	assert.Contains(t, query, "SELECT *")
	assert.Contains(t, query, "FROM SOURCES")
	assert.Contains(t, query, "WHERE level = 'error' OR level = 'err'")
	assert.Contains(t, query, "ORDER BY ts DESC")

	// Should work with any params (ignored)
	query, err = report.BuildQueryString(map[string]any{"ignored": "value"})
	require.NoError(t, err)
	assert.Contains(t, query, "level = 'error'")
}

// TestReportFindParseErrors_QueryStructure tests parse errors query structure
func TestReportFindParseErrors_QueryStructure(t *testing.T) {
	report := &ReportFindParseErrors{}
	query, err := report.BuildQueryString(nil)

	require.NoError(t, err)

	// Verify query components
	assert.Contains(t, query, "SELECT *")
	assert.Contains(t, query, "FROM SOURCES")
	assert.Contains(t, query, "WHERE parse_error is not null")

	// Should not have ORDER BY (let caller decide)
	assert.NotContains(t, query, "ORDER BY")
}

// TestReportFindServiceErrors_QueryStructure tests service errors query structure
func TestReportFindServiceErrors_QueryStructure(t *testing.T) {
	report := &ReportFindServiceErrors{}
	query, err := report.BuildQueryString(nil)

	require.NoError(t, err)

	// Verify query components
	assert.Contains(t, query, "SELECT *")
	assert.Contains(t, query, "FROM SOURCES")
	assert.Contains(t, query, "WHERE level = 'error' OR level = 'err'")
	assert.Contains(t, query, "ORDER BY ts DESC")
}

// TestRegistry_AllReportsHaveUniqueName tests that all reports have unique names
func TestRegistry_AllReportsHaveUniqueName(t *testing.T) {
	registry := New()

	names := make(map[string]bool)
	for _, report := range registry {
		name := report.Name()
		assert.False(t, names[name], "Duplicate report name: %s", name)
		names[name] = true
	}
}

// TestRegistry_AllReportsHaveDescription tests that all reports have descriptions
func TestRegistry_AllReportsHaveDescription(t *testing.T) {
	registry := New()

	for _, report := range registry {
		desc := report.Description()
		assert.NotEmpty(t, desc, "Report %s has empty description", report.Name())
		assert.Greater(t, len(desc), 20, "Report %s description too short", report.Name())
	}
}

// TestRegistry_AllReportsCanBuildQuery tests that all reports can build queries
func TestRegistry_AllReportsCanBuildQuery(t *testing.T) {
	registry := New()

	for _, report := range registry {
		t.Run(report.Name(), func(t *testing.T) {
			query, err := report.BuildQueryString(nil)
			require.NoError(t, err, "Report %s failed to build query", report.Name())
			assert.NotEmpty(t, query, "Report %s produced empty query", report.Name())
			assert.Contains(t, query, "SELECT", "Report %s query missing SELECT", report.Name())
			assert.Contains(t, query, "FROM", "Report %s query missing FROM", report.Name())
		})
	}
}

// TestRegistry_AllReportsReturnNonNilParams tests param specs are not nil
func TestRegistry_AllReportsReturnNonNilParams(t *testing.T) {
	registry := New()

	for _, report := range registry {
		t.Run(report.Name(), func(t *testing.T) {
			params := report.ParamSpecs()
			assert.NotNil(t, params, "Report %s returned nil ParamSpecs", report.Name())

			samples := report.SampleParams()
			assert.NotNil(t, samples, "Report %s returned nil SampleParams", report.Name())
		})
	}
}

// TestGet_CaseSensitive tests that report names are case-sensitive
func TestGet_CaseSensitive(t *testing.T) {
	tests := []struct {
		name       string
		reportName string
		wantErr    bool
	}{
		{
			name:       "correct case",
			reportName: "find-all-parse-errors",
			wantErr:    false,
		},
		{
			name:       "uppercase",
			reportName: "FIND-ALL-PARSE-ERRORS",
			wantErr:    true,
		},
		{
			name:       "mixed case",
			reportName: "Find-All-Parse-Errors",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report, err := Get(tt.reportName)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, report)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, report)
			}
		})
	}
}

// BenchmarkNew benchmarks registry creation
func BenchmarkNew(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = New()
	}
}

// BenchmarkGet benchmarks report retrieval
func BenchmarkGet(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Get("find-all-parse-errors")
	}
}

// BenchmarkBuildQueryString benchmarks query building
func BenchmarkBuildQueryString(b *testing.B) {
	report := &ReportFindParseErrors{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = report.BuildQueryString(nil)
	}
}
