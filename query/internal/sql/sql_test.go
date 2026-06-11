// SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

package sql

import (
	"strings"
	"testing"

	"github.com/OpenCHAMI/legendary-funicular/query/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper to create test config
func testConfig() *config.Config {
	cfg, _ := config.New(
		config.WithDefaults(),
		config.WithS3KeyAccess("test-access-key"),
		config.WithS3KeySecret("test-secret-key"),
	)
	return cfg
}

// TestBuildQuerySecret tests the secret query generation
func TestBuildQuerySecret(t *testing.T) {
	cfg := testConfig()

	query := buildQuerySecret(cfg)

	// Verify query structure
	assert.Contains(t, query, "CREATE OR REPLACE SECRET local_s3")
	assert.Contains(t, query, "TYPE s3")
	assert.Contains(t, query, "PROVIDER config")

	// Verify credentials
	assert.Contains(t, query, "test-access-key")
	assert.Contains(t, query, "test-secret-key")

	// Verify config values
	assert.Contains(t, query, "us-east-1")      // Default region
	assert.Contains(t, query, "localhost:7070") // Default endpoint
	assert.Contains(t, query, "USE_SSL false")  // Default SSL
	assert.Contains(t, query, "URL_STYLE 'path'")
}

// TestBuildQuerySecret_WithSSL tests secret generation with SSL enabled
func TestBuildQuerySecret_WithSSL(t *testing.T) {
	cfg, _ := config.New(
		config.WithS3KeyAccess("access"),
		config.WithS3KeySecret("secret"),
		config.WithS3Region("us-west-2"),
		config.WithS3Endpoint("https://s3.amazonaws.com"),
		config.WithS3SSL(true),
		config.WithS3BucketNDJSON("bucket1"),
		config.WithS3BucketParquet("bucket2"),
	)

	query := buildQuerySecret(cfg)

	assert.Contains(t, query, "USE_SSL true")
	assert.Contains(t, query, "us-west-2")
	assert.Contains(t, query, "https://s3.amazonaws.com")
}

// TestQueryPlaceholderSources tests the constant
func TestQueryPlaceholderSources(t *testing.T) {
	assert.Equal(t, "SOURCES", QueryPlaceholderSources)
}

// TestPrepareQuerystr_SimpleSELECT tests basic SELECT query
func TestPrepareQuerystr_SimpleSELECT(t *testing.T) {
	e := &Engine{cfg: testConfig()}
	sources := []string{"read_parquet('s3://bucket/logs/*.parquet')"}

	query := "SELECT * FROM SOURCES"
	result, err := e.prepareQuerystr(query, sources)

	require.NoError(t, err)
	assert.Contains(t, result, "read_parquet")
	assert.Contains(t, result, "* exclude(data), json(data) as data")
}

// TestPrepareQuerystr_TrailingSemicolon tests semicolon removal
func TestPrepareQuerystr_TrailingSemicolon(t *testing.T) {
	e := &Engine{cfg: testConfig()}
	sources := []string{"source1"}

	query := "SELECT * FROM SOURCES;"
	result, err := e.prepareQuerystr(query, sources)

	require.NoError(t, err)
	assert.NotContains(t, result, ";")
}

// TestPrepareQuerystr_NoFROM tests missing FROM clause
func TestPrepareQuerystr_NoFROM(t *testing.T) {
	e := &Engine{cfg: testConfig()}
	sources := []string{"source1"}

	query := "SELECT *"
	_, err := e.prepareQuerystr(query, sources)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid SQL query")
}

// TestPrepareQuerystr_SingleSource tests single source replacement
func TestPrepareQuerystr_SingleSource(t *testing.T) {
	e := &Engine{cfg: testConfig()}
	sources := []string{"read_parquet('s3://bucket/data.parquet')"}

	query := "SELECT * FROM SOURCES"
	result, err := e.prepareQuerystr(query, sources)

	require.NoError(t, err)
	assert.Contains(t, result, "read_parquet('s3://bucket/data.parquet')")
	assert.NotContains(t, result, "SOURCES")
}

// TestPrepareQuerystr_MultipleSources tests UNION ALL generation
func TestPrepareQuerystr_MultipleSources(t *testing.T) {
	e := &Engine{cfg: testConfig()}
	sources := []string{
		"read_parquet('s3://bucket/logs/*.parquet')",
		"read_json('s3://bucket/logs/*.ndjson.zst')",
	}

	query := "SELECT * FROM SOURCES"
	result, err := e.prepareQuerystr(query, sources)

	require.NoError(t, err)
	assert.Contains(t, result, "UNION ALL BY NAME")
	assert.Contains(t, result, "read_parquet")
	assert.Contains(t, result, "read_json")
	assert.Contains(t, result, "(SELECT * FROM")
}

// TestPrepareQuerystr_NilSources tests nil sources error
func TestPrepareQuerystr_NilSources(t *testing.T) {
	e := &Engine{cfg: testConfig()}

	query := "SELECT * FROM SOURCES"
	_, err := e.prepareQuerystr(query, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sources slice is nil")
}

// TestPrepareQuerystr_EmptySources tests empty sources error
func TestPrepareQuerystr_EmptySources(t *testing.T) {
	e := &Engine{cfg: testConfig()}

	query := "SELECT * FROM SOURCES"
	_, err := e.prepareQuerystr(query, []string{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sources slice is empty")
}

// TestPrepareQuerystr_EmptySourceString tests empty source string error
func TestPrepareQuerystr_EmptySourceString(t *testing.T) {
	e := &Engine{cfg: testConfig()}

	query := "SELECT * FROM SOURCES"
	_, err := e.prepareQuerystr(query, []string{"source1", "", "source3"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sources[1] is an empty string")
}

// TestPrepareQuerystr_DataFieldExpansion tests data field JSON expansion
func TestPrepareQuerystr_DataFieldExpansion(t *testing.T) {
	e := &Engine{cfg: testConfig()}
	sources := []string{"read_parquet('s3://bucket/logs/*.parquet')"}

	tests := []struct {
		name     string
		query    string
		contains string
	}{
		{
			name:     "SELECT * expands data field",
			query:    "SELECT * FROM SOURCES",
			contains: "* exclude(data), json(data) as data",
		},
		{
			name:     "SELECT data expands to json(data)",
			query:    "SELECT data FROM SOURCES",
			contains: "json(data) as data",
		},
		{
			name:     "SELECT without data or * no expansion",
			query:    "SELECT ts, host FROM SOURCES",
			contains: "SELECT ts, host FROM",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := e.prepareQuerystr(tt.query, sources)
			require.NoError(t, err)
			assert.Contains(t, result, tt.contains)
		})
	}
}

// TestPrepareQuerystr_CloudEventExpansion tests cloudevent field expansion
func TestPrepareQuerystr_CloudEventExpansion(t *testing.T) {
	e := &Engine{cfg: testConfig()}
	sources := []string{"read_parquet('s3://bucket/events/*.parquet')"}

	query := "SELECT * FROM SOURCES"
	result, err := e.prepareQuerystr(query, sources)

	require.NoError(t, err)
	assert.Contains(t, result, "* exclude(cloudevent), json(cloudevent) as cloudevent")
	assert.NotContains(t, result, "data")
}

// TestPrepareQuerystr_CaseInsensitive tests case insensitivity
func TestPrepareQuerystr_CaseInsensitive(t *testing.T) {
	e := &Engine{cfg: testConfig()}
	sources := []string{"source1"}

	tests := []struct {
		name  string
		query string
	}{
		{"lowercase from", "select * from SOURCES"},
		{"uppercase FROM", "SELECT * FROM SOURCES"},
		{"mixed case From", "Select * From SOURCES"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := e.prepareQuerystr(tt.query, sources)
			assert.NoError(t, err)
		})
	}
}

// TestPrepareQuerystr_ComplexQuery tests more complex SQL
func TestPrepareQuerystr_ComplexQuery(t *testing.T) {
	e := &Engine{cfg: testConfig()}
	sources := []string{
		"read_parquet('s3://bucket/logs/*.parquet')",
		"read_json('s3://bucket/logs/*.ndjson.zst')",
	}

	query := "SELECT * FROM SOURCES WHERE level = 'ERROR' ORDER BY ts DESC LIMIT 100"
	result, err := e.prepareQuerystr(query, sources)

	require.NoError(t, err)
	assert.Contains(t, result, "WHERE level = 'ERROR'")
	assert.Contains(t, result, "ORDER BY ts DESC")
	assert.Contains(t, result, "LIMIT 100")
	assert.Contains(t, result, "UNION ALL BY NAME")
}

// TestPrepareQuerystr_NoSOURCESPlaceholder tests query without SOURCES
func TestPrepareQuerystr_NoSOURCESPlaceholder(t *testing.T) {
	e := &Engine{cfg: testConfig()}

	query := "SELECT * FROM my_table"
	result, err := e.prepareQuerystr(query, nil)

	require.NoError(t, err)
	assert.Contains(t, result, "FROM my_table")
	assert.Contains(t, result, "* exclude(data), json(data) as data")
}

// TestPrepareQuerystr_WhitespaceHandling tests whitespace handling
func TestPrepareQuerystr_WhitespaceHandling(t *testing.T) {
	e := &Engine{cfg: testConfig()}
	sources := []string{"source1"}

	tests := []struct {
		name  string
		query string
	}{
		{"leading whitespace", "  SELECT * FROM SOURCES"},
		{"trailing whitespace", "SELECT * FROM SOURCES  "},
		{"leading and trailing", "  SELECT * FROM SOURCES  "},
		{"semicolon with whitespace", "SELECT * FROM SOURCES  ;  "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := e.prepareQuerystr(tt.query, sources)
			require.NoError(t, err)
			assert.NotContains(t, result, "SOURCES")
		})
	}
}

// TestPrepareQuerystr_SpecialCharacters tests queries with special chars
func TestPrepareQuerystr_SpecialCharacters(t *testing.T) {
	e := &Engine{cfg: testConfig()}
	sources := []string{"source1"}

	tests := []struct {
		name  string
		query string
	}{
		{"single quotes", "SELECT * FROM SOURCES WHERE msg = 'test'"},
		{"double quotes", `SELECT * FROM SOURCES WHERE "column" = 'value'`},
		{"parentheses", "SELECT * FROM SOURCES WHERE (level = 'ERROR' OR level = 'WARN')"},
		{"wildcards in WHERE", "SELECT * FROM SOURCES WHERE msg LIKE '%error%'"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := e.prepareQuerystr(tt.query, sources)
			require.NoError(t, err)
			assert.NotContains(t, result, "SOURCES")
		})
	}
}

// TestScanRowDirectToMap_EmptyColumns tests empty column set
func TestScanRowDirectToMap_EmptyColumns(t *testing.T) {
	// Note: This would require mocking sql.Rows which is complex
	// Documenting the expected behavior instead
	t.Skip("Requires sql.Rows mock - behavior: should return empty map")
}

// TestScanRowJSONToMap_EmptyJSON tests empty JSON
func TestScanRowJSONToMap_EmptyJSON(t *testing.T) {
	// Note: This would require mocking sql.Rows which is complex
	// Documenting the expected behavior instead
	t.Skip("Requires sql.Rows mock - behavior: should return empty map or error")
}

// BenchmarkBuildQuerySecret benchmarks secret query generation
func BenchmarkBuildQuerySecret(b *testing.B) {
	cfg := testConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = buildQuerySecret(cfg)
	}
}

// BenchmarkPrepareQuerystr_SingleSource benchmarks single source query prep
func BenchmarkPrepareQuerystr_SingleSource(b *testing.B) {
	e := &Engine{cfg: testConfig()}
	sources := []string{"read_parquet('s3://bucket/data.parquet')"}
	query := "SELECT * FROM SOURCES"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = e.prepareQuerystr(query, sources)
	}
}

// BenchmarkPrepareQuerystr_MultipleSources benchmarks multiple source query prep
func BenchmarkPrepareQuerystr_MultipleSources(b *testing.B) {
	e := &Engine{cfg: testConfig()}
	sources := []string{
		"read_parquet('s3://bucket/logs/*.parquet')",
		"read_json('s3://bucket/logs/*.ndjson.zst')",
	}
	query := "SELECT * FROM SOURCES WHERE level = 'ERROR'"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = e.prepareQuerystr(query, sources)
	}
}

// BenchmarkPrepareQuerystr_ComplexQuery benchmarks complex query prep
func BenchmarkPrepareQuerystr_ComplexQuery(b *testing.B) {
	e := &Engine{cfg: testConfig()}
	sources := []string{
		"read_parquet('s3://bucket1/logs/*.parquet')",
		"read_parquet('s3://bucket2/logs/*.parquet')",
		"read_json('s3://bucket1/logs/*.ndjson.zst')",
		"read_json('s3://bucket2/logs/*.ndjson.zst')",
	}
	query := "SELECT * FROM SOURCES WHERE level IN ('ERROR', 'WARN') AND ts > '2026-01-01' ORDER BY ts DESC LIMIT 1000"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = e.prepareQuerystr(query, sources)
	}
}

// TestPrepareQuerystr_RealWorldQueries tests realistic query scenarios
func TestPrepareQuerystr_RealWorldQueries(t *testing.T) {
	e := &Engine{cfg: testConfig()}

	tests := []struct {
		name    string
		query   string
		sources []string
		wantErr bool
	}{
		{
			name:  "Count by level",
			query: "SELECT level, COUNT(*) as count FROM SOURCES GROUP BY level",
			sources: []string{
				"read_parquet('s3://bucket/logs/*.parquet')",
			},
			wantErr: false,
		},
		{
			name:  "Recent errors",
			query: "SELECT * FROM SOURCES WHERE level = 'ERROR' ORDER BY ts DESC LIMIT 10",
			sources: []string{
				"read_parquet('s3://bucket/logs/*.parquet')",
			},
			wantErr: false,
		},
		{
			name:  "Time range query",
			query: "SELECT * FROM SOURCES WHERE ts BETWEEN '2026-01-01' AND '2026-01-31'",
			sources: []string{
				"read_parquet('s3://bucket/logs/*.parquet')",
			},
			wantErr: false,
		},
		{
			name:  "Aggregation query",
			query: "SELECT host, COUNT(*) as count, AVG(length(msg)) as avg_msg_len FROM SOURCES GROUP BY host",
			sources: []string{
				"read_parquet('s3://bucket/logs/*.parquet')",
			},
			wantErr: false,
		},
		{
			name:  "Join-like UNION",
			query: "SELECT * FROM SOURCES WHERE service = 'api'",
			sources: []string{
				"read_parquet('s3://bucket/logs/*.parquet')",
				"read_json('s3://bucket/logs/*.ndjson.zst')",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := e.prepareQuerystr(tt.query, tt.sources)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, result)
				assert.NotContains(t, result, "SOURCES")

				// Verify query structure is preserved
				lowerResult := strings.ToLower(result)
				assert.Contains(t, lowerResult, "from")

				// If multiple sources, should have UNION
				if len(tt.sources) > 1 {
					assert.Contains(t, result, "UNION ALL BY NAME")
				}
			}
		})
	}
}
