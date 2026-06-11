// SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

package report

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockReport is a test implementation of Report interface
type MockReport struct {
	name         string
	description  string
	paramSpecs   []ParamSpec
	sampleParams map[string]any
	queryString  string
	buildError   error
}

func (m *MockReport) Name() string                 { return m.name }
func (m *MockReport) Description() string          { return m.description }
func (m *MockReport) ParamSpecs() []ParamSpec      { return m.paramSpecs }
func (m *MockReport) SampleParams() map[string]any { return m.sampleParams }
func (m *MockReport) BuildQueryString(params map[string]any) (string, error) {
	return m.queryString, m.buildError
}

// TestParamSpec tests ParamSpec struct
func TestParamSpec(t *testing.T) {
	spec := ParamSpec{
		Name:        "test_param",
		Description: "A test parameter",
		Required:    true,
		Default:     "default_value",
	}

	assert.Equal(t, "test_param", spec.Name)
	assert.Equal(t, "A test parameter", spec.Description)
	assert.True(t, spec.Required)
	assert.Equal(t, "default_value", spec.Default)
}

// TestParamSpec_OptionalWithDefault tests optional param with default
func TestParamSpec_OptionalWithDefault(t *testing.T) {
	spec := ParamSpec{
		Name:        "limit",
		Description: "Maximum number of results",
		Required:    false,
		Default:     100,
	}

	assert.Equal(t, "limit", spec.Name)
	assert.False(t, spec.Required)
	assert.Equal(t, 100, spec.Default)
}

// TestParamSpec_RequiredNoDefault tests required param without default
func TestParamSpec_RequiredNoDefault(t *testing.T) {
	spec := ParamSpec{
		Name:        "user_id",
		Description: "User identifier",
		Required:    true,
		Default:     nil,
	}

	assert.Equal(t, "user_id", spec.Name)
	assert.True(t, spec.Required)
	assert.Nil(t, spec.Default)
}

// TestBuildSchemaQueryStr tests schema query string generation
func TestBuildSchemaQueryStr(t *testing.T) {
	report := &MockReport{
		name:         "test-report",
		queryString:  "SELECT * FROM test_table",
		sampleParams: map[string]any{"param1": "value1"},
	}

	query, err := BuildSchemaQueryStr(report)

	require.NoError(t, err)
	assert.Contains(t, query, "SELECT")
	assert.Contains(t, query, "column_name")
	assert.Contains(t, query, "column_type")
	assert.Contains(t, query, "is_nullable")
	assert.Contains(t, query, "DESCRIBE")
	assert.Contains(t, query, "SELECT * FROM test_table")
}

// TestBuildSchemaQueryStr_WithError tests error propagation
func TestBuildSchemaQueryStr_WithError(t *testing.T) {
	report := &MockReport{
		name:         "test-report",
		buildError:   assert.AnError,
		sampleParams: map[string]any{},
	}

	_, err := BuildSchemaQueryStr(report)

	assert.Error(t, err)
	assert.Equal(t, assert.AnError, err)
}

// TestBuildSchemaQueryStr_ComplexQuery tests with complex query
func TestBuildSchemaQueryStr_ComplexQuery(t *testing.T) {
	report := &MockReport{
		name: "complex-report",
		queryString: `
			SELECT id, name, created_at
			FROM users
			WHERE status = 'active'
			ORDER BY created_at DESC
		`,
		sampleParams: map[string]any{"status": "active"},
	}

	query, err := BuildSchemaQueryStr(report)

	require.NoError(t, err)
	assert.Contains(t, query, "DESCRIBE")
	assert.Contains(t, query, "SELECT id, name, created_at")
	assert.Contains(t, query, "WHERE status = 'active'")
}

// TestParseRawParams_Empty tests parsing empty params
func TestParseRawParams_Empty(t *testing.T) {
	params, err := ParseRawParams([]string{})

	require.NoError(t, err)
	assert.NotNil(t, params)
	assert.Len(t, params, 0)
}

// TestParseRawParams_Nil tests parsing nil params
func TestParseRawParams_Nil(t *testing.T) {
	params, err := ParseRawParams(nil)

	require.NoError(t, err)
	assert.NotNil(t, params)
	assert.Len(t, params, 0)
}

// TestParseRawParams_NotImplemented tests that non-empty params return not implemented
func TestParseRawParams_NotImplemented(t *testing.T) {
	tests := []struct {
		name   string
		params []string
	}{
		{
			name:   "single param",
			params: []string{"key=value"},
		},
		{
			name:   "multiple params",
			params: []string{"key1=value1", "key2=value2"},
		},
		{
			name:   "complex param",
			params: []string{"filter=status='active'"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params, err := ParseRawParams(tt.params)

			assert.Error(t, err)
			assert.NotNil(t, params)
			// Error should indicate not implemented
			assert.Contains(t, err.Error(), "not implemented")
		})
	}
}

// TestMockReport_Interface tests that MockReport implements Report interface
func TestMockReport_Interface(t *testing.T) {
	var _ Report = (*MockReport)(nil) // Compile-time check

	report := &MockReport{
		name:        "test-report",
		description: "A test report",
		paramSpecs: []ParamSpec{
			{Name: "param1", Description: "First param", Required: true},
		},
		sampleParams: map[string]any{"param1": "value1"},
		queryString:  "SELECT * FROM test",
	}

	assert.Equal(t, "test-report", report.Name())
	assert.Equal(t, "A test report", report.Description())
	assert.Len(t, report.ParamSpecs(), 1)
	assert.Len(t, report.SampleParams(), 1)

	query, err := report.BuildQueryString(nil)
	require.NoError(t, err)
	assert.Equal(t, "SELECT * FROM test", query)
}

// TestBuildSchemaQueryStr_EmptyQuery tests with empty query string
func TestBuildSchemaQueryStr_EmptyQuery(t *testing.T) {
	report := &MockReport{
		name:         "empty-report",
		queryString:  "",
		sampleParams: map[string]any{},
	}

	query, err := BuildSchemaQueryStr(report)

	require.NoError(t, err)
	assert.Contains(t, query, "DESCRIBE")
	// Empty query should still be wrapped
	assert.Contains(t, query, "column_name")
}

// TestBuildSchemaQueryStr_WithSampleParams tests that sample params are used
func TestBuildSchemaQueryStr_WithSampleParams(t *testing.T) {
	report := &MockReport{
		name: "param-test",
		sampleParams: map[string]any{
			"test_param": "test_value",
			"limit":      100,
		},
		queryString: "SELECT * FROM test WHERE param = 'test_value' LIMIT 100",
	}

	query, err := BuildSchemaQueryStr(report)
	require.NoError(t, err)

	// Verify the original query is wrapped in DESCRIBE
	assert.Contains(t, query, "DESCRIBE")
	assert.Contains(t, query, "SELECT * FROM test")
}

// TestParamSpec_VariousTypes tests ParamSpec with various default types
func TestParamSpec_VariousTypes(t *testing.T) {
	tests := []struct {
		name        string
		spec        ParamSpec
		wantDefault any
	}{
		{
			name: "string default",
			spec: ParamSpec{
				Name:    "string_param",
				Default: "default_string",
			},
			wantDefault: "default_string",
		},
		{
			name: "int default",
			spec: ParamSpec{
				Name:    "int_param",
				Default: 42,
			},
			wantDefault: 42,
		},
		{
			name: "bool default",
			spec: ParamSpec{
				Name:    "bool_param",
				Default: true,
			},
			wantDefault: true,
		},
		{
			name: "slice default",
			spec: ParamSpec{
				Name:    "slice_param",
				Default: []string{"a", "b", "c"},
			},
			wantDefault: []string{"a", "b", "c"},
		},
		{
			name: "map default",
			spec: ParamSpec{
				Name:    "map_param",
				Default: map[string]int{"key": 123},
			},
			wantDefault: map[string]int{"key": 123},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantDefault, tt.spec.Default)
		})
	}
}

// BenchmarkBuildSchemaQueryStr benchmarks schema query generation
func BenchmarkBuildSchemaQueryStr(b *testing.B) {
	report := &MockReport{
		name:         "benchmark-report",
		queryString:  "SELECT * FROM test_table WHERE id > 100",
		sampleParams: map[string]any{"limit": 100},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = BuildSchemaQueryStr(report)
	}
}

// BenchmarkParseRawParams_Empty benchmarks parsing empty params
func BenchmarkParseRawParams_Empty(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseRawParams([]string{})
	}
}

// BenchmarkParseRawParams_NotImplemented benchmarks not implemented path
func BenchmarkParseRawParams_NotImplemented(b *testing.B) {
	params := []string{"key=value"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseRawParams(params)
	}
}
