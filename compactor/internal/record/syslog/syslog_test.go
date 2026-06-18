// SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

package syslog

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParse_ValidJSON tests parsing of valid syslog JSON
func TestParse_ValidJSON(t *testing.T) {
	validJSON := `{
		"ts": "2026-06-10T12:00:00Z",
		"service": "test-service",
		"host": "test-host",
		"level": "INFO",
		"msg": "Test message"
	}`

	record, err := Parse([]byte(validJSON))
	require.NoError(t, err)
	assert.NotNil(t, record.Timestamp)
	assert.Equal(t, "2026-06-10T12:00:00Z", *record.Timestamp)
	assert.NotNil(t, record.Service)
	assert.Equal(t, "test-service", *record.Service)
	assert.NotNil(t, record.Host)
	assert.Equal(t, "test-host", *record.Host)
	assert.NotNil(t, record.Level)
	assert.Equal(t, "INFO", *record.Level)
	assert.NotNil(t, record.Message)
	assert.Equal(t, "Test message", *record.Message)
	assert.NotNil(t, record.PayloadJSON)
}

// TestParse_MinimalJSON tests parsing with minimal required fields
func TestParse_MinimalJSON(t *testing.T) {
	minimalJSON := `{"ts": "2026-06-10T12:00:00Z"}`

	record, err := Parse([]byte(minimalJSON))
	require.NoError(t, err)
	assert.NotNil(t, record.Timestamp)
	assert.Nil(t, record.Service)
	assert.Nil(t, record.Host)
}

// TestParse_OpenCHAMIFields tests OpenCHAMI-specific fields
func TestParse_OpenCHAMIFields(t *testing.T) {
	json := `{
		"ts": "2026-06-10T12:00:00Z",
		"xname": "x3000c0s1b0n0",
		"component_id": "comp-123",
		"node_id": "node-456",
		"trace_id": "trace-789",
		"request_id": "req-abc",
		"request_uri": "/api/v1/resource",
		"request_user": "admin"
	}`

	record, err := Parse([]byte(json))
	require.NoError(t, err)
	assert.NotNil(t, record.XName)
	assert.Equal(t, "x3000c0s1b0n0", *record.XName)
	assert.NotNil(t, record.IDComponent)
	assert.Equal(t, "comp-123", *record.IDComponent)
	assert.NotNil(t, record.IDNode)
	assert.Equal(t, "node-456", *record.IDNode)
	assert.NotNil(t, record.IDTrace)
	assert.Equal(t, "trace-789", *record.IDTrace)
}

// TestParse_MalformedJSON tests handling of malformed JSON
func TestParse_MalformedJSON(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"missing closing brace", `{"ts": "2026-06-10T12:00:00Z"`},
		{"invalid json syntax", `{ts: 2026-06-10}`},
		{"trailing comma", `{"ts": "2026-06-10T12:00:00Z",}`},
		{"unquoted keys", `{ts: "2026-06-10T12:00:00Z"}`},
		{"empty string", ``},
		{"just whitespace", `   `},
		// Note: "null" is actually valid JSON, removed from malformed tests
		{"array instead of object", `["ts", "2026-06-10T12:00:00Z"]`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record, err := Parse([]byte(tt.input))

			// Parser should return an error for malformed JSON
			assert.Error(t, err, "Expected error for malformed JSON: %s", tt.name)

			// PayloadJSON should not be set on error
			assert.Nil(t, record.PayloadJSON, "PayloadJSON should be nil on parse error")
		})
	}
}

// TestParse_SpecialCharacters tests handling of special characters
func TestParse_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name    string
		message string
	}{
		{"newlines", "Line 1\\nLine 2\\nLine 3"},
		{"tabs", "Column1\\tColumn2\\tColumn3"},
		{"quotes", `He said \"hello\"`},
		{"unicode", "Hello 世界 🌍"},
		{"null bytes", "Before\\u0000After"},
		{"backslashes", "C:\\\\Users\\\\test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			json := `{"ts": "2026-06-10T12:00:00Z", "msg": "` + tt.message + `"}`

			record, err := Parse([]byte(json))
			require.NoError(t, err)
			assert.NotNil(t, record.Message)
		})
	}
}

// TestParse_LargeMessage tests handling of very large messages
func TestParse_LargeMessage(t *testing.T) {
	// Create a 1MB message
	largeMsg := make([]byte, 1024*1024)
	for i := range largeMsg {
		largeMsg[i] = 'A'
	}

	json := `{"ts": "2026-06-10T12:00:00Z", "msg": "` + string(largeMsg) + `"}`

	record, err := Parse([]byte(json))
	require.NoError(t, err)
	assert.NotNil(t, record.Message)
	assert.Len(t, *record.Message, 1024*1024)
}

// TestParse_EmptyFields tests handling of empty field values
func TestParse_EmptyFields(t *testing.T) {
	json := `{
		"ts": "",
		"service": "",
		"host": "",
		"level": "",
		"msg": ""
	}`

	record, err := Parse([]byte(json))
	require.NoError(t, err)

	// Empty strings should still be set (as empty string pointers)
	assert.NotNil(t, record.Timestamp)
	assert.Equal(t, "", *record.Timestamp)
	assert.NotNil(t, record.Service)
	assert.Equal(t, "", *record.Service)
}

// TestParse_NullFields tests handling of null field values
func TestParse_NullFields(t *testing.T) {
	json := `{
		"ts": null,
		"service": null,
		"host": "test-host",
		"level": null,
		"msg": null
	}`

	record, err := Parse([]byte(json))
	require.NoError(t, err)

	// Null values should result in nil pointers
	assert.Nil(t, record.Timestamp)
	assert.Nil(t, record.Service)
	assert.NotNil(t, record.Host) // This one is set
	assert.Nil(t, record.Level)
	assert.Nil(t, record.Message)
}

// TestParse_ExtraFields tests that extra fields don't cause errors
func TestParse_ExtraFields(t *testing.T) {
	json := `{
		"ts": "2026-06-10T12:00:00Z",
		"service": "test-service",
		"unknown_field_1": "value1",
		"unknown_field_2": 12345,
		"unknown_field_3": true
	}`

	record, err := Parse([]byte(json))
	require.NoError(t, err)
	assert.NotNil(t, record.Timestamp)
	assert.NotNil(t, record.Service)
}

// TestParse_PayloadJSONPreservation tests that original JSON is preserved
func TestParse_PayloadJSONPreservation(t *testing.T) {
	originalJSON := `{"ts":"2026-06-10T12:00:00Z","service":"test"}`

	record, err := Parse([]byte(originalJSON))
	require.NoError(t, err)
	require.NotNil(t, record.PayloadJSON)
	assert.Equal(t, originalJSON, *record.PayloadJSON)
}

// TestParse_ConcurrentParsing tests thread-safety of Parse
func TestParse_ConcurrentParsing(t *testing.T) {
	json := `{"ts": "2026-06-10T12:00:00Z", "service": "test-service"}`

	// Parse the same JSON from 100 goroutines concurrently
	done := make(chan bool, 100)
	for i := 0; i < 100; i++ {
		go func() {
			record, err := Parse([]byte(json))
			assert.NoError(t, err)
			assert.NotNil(t, record.Timestamp)
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 100; i++ {
		<-done
	}
}

// BenchmarkParse_Small benchmarks parsing of small messages
func BenchmarkParse_Small(b *testing.B) {
	json := []byte(`{"ts": "2026-06-10T12:00:00Z", "service": "test", "msg": "Small message"}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Parse(json)
	}
}

// BenchmarkParse_Medium benchmarks parsing of medium messages
func BenchmarkParse_Medium(b *testing.B) {
	msg := make([]byte, 1024) // 1KB message
	for i := range msg {
		msg[i] = 'A'
	}
	json := []byte(`{"ts": "2026-06-10T12:00:00Z", "service": "test", "msg": "` + string(msg) + `"}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Parse(json)
	}
}

// BenchmarkParse_Large benchmarks parsing of large messages
func BenchmarkParse_Large(b *testing.B) {
	msg := make([]byte, 100*1024) // 100KB message
	for i := range msg {
		msg[i] = 'A'
	}
	json := []byte(`{"ts": "2026-06-10T12:00:00Z", "service": "test", "msg": "` + string(msg) + `"}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Parse(json)
	}
}

// FuzzParse fuzzes the Parse function with random inputs
func FuzzParse(f *testing.F) {
	// Seed corpus with valid examples
	f.Add([]byte(`{"ts": "2026-06-10T12:00:00Z"}`))
	f.Add([]byte(`{"ts": "2026-06-10T12:00:00Z", "service": "test"}`))
	f.Add([]byte(`{"ts": "2026-06-10T12:00:00Z", "msg": "Test message"}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`null`))
	f.Add([]byte(``))

	f.Fuzz(func(t *testing.T, data []byte) {
		// Parse should not panic on any input
		record, err := Parse(data)

		if err == nil {
			// If parsing succeeded, PayloadJSON should be set
			assert.NotNil(t, record.PayloadJSON)
		} else {
			// If parsing failed, PayloadJSON should be nil
			assert.Nil(t, record.PayloadJSON)
		}
	})
}
