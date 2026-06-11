// SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

package cloudevent

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParse_ValidJSON tests parsing of valid CloudEvent JSON
func TestParse_ValidJSON(t *testing.T) {
	validJSON := `{
		"ts": "2026-06-10T12:00:00Z",
		"transport_method": "http",
		"cloudevent_id": "evt-123",
		"cloudevent_source": "/api/events",
		"cloudevent_type": "com.openchami.event",
		"cloudevent_specversion": "1.0"
	}`

	record, err := Parse([]byte(validJSON))
	require.NoError(t, err)
	assert.NotNil(t, record.Timestamp)
	assert.Equal(t, "2026-06-10T12:00:00Z", *record.Timestamp)
	assert.NotNil(t, record.TransportMethod)
	assert.Equal(t, "http", *record.TransportMethod)
	assert.NotNil(t, record.EventID)
	assert.Equal(t, "evt-123", *record.EventID)
	assert.NotNil(t, record.EventSource)
	assert.Equal(t, "/api/events", *record.EventSource)
	assert.NotNil(t, record.EventType)
	assert.Equal(t, "com.openchami.event", *record.EventType)
	assert.NotNil(t, record.PayloadJSON)
}

// TestParse_MinimalJSON tests parsing with minimal required fields
func TestParse_MinimalJSON(t *testing.T) {
	minimalJSON := `{"ts": "2026-06-10T12:00:00Z"}`

	record, err := Parse([]byte(minimalJSON))
	require.NoError(t, err)
	assert.NotNil(t, record.Timestamp)
	assert.Nil(t, record.EventID)
	assert.Nil(t, record.EventSource)
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

// TestParse_CloudEventSpec tests CloudEvent specification fields
func TestParse_CloudEventSpec(t *testing.T) {
	json := `{
		"ts": "2026-06-10T12:00:00Z",
		"cloudevent": "full-event-json",
		"cloudevent_id": "a234-1234-1234",
		"cloudevent_source": "https://github.com/OpenCHAMI",
		"cloudevent_type": "com.openchami.state.change",
		"cloudevent_specversion": "1.0",
		"cloudevent_binding": "http-binary"
	}`

	record, err := Parse([]byte(json))
	require.NoError(t, err)
	assert.NotNil(t, record.Event)
	assert.Equal(t, "full-event-json", *record.Event)
	assert.NotNil(t, record.EventID)
	assert.Equal(t, "a234-1234-1234", *record.EventID)
	assert.NotNil(t, record.EventSource)
	assert.Equal(t, "https://github.com/OpenCHAMI", *record.EventSource)
	assert.NotNil(t, record.EventType)
	assert.Equal(t, "com.openchami.state.change", *record.EventType)
	assert.NotNil(t, record.EventSpecversion)
	assert.Equal(t, "1.0", *record.EventSpecversion)
	assert.NotNil(t, record.EventBinding)
	assert.Equal(t, "http-binary", *record.EventBinding)
}

// TestParse_TransportFields tests transport metadata fields
func TestParse_TransportFields(t *testing.T) {
	json := `{
		"ts": "2026-06-10T12:00:00Z",
		"transport_method": "kafka",
		"transport_metadata": "topic=events,partition=0"
	}`

	record, err := Parse([]byte(json))
	require.NoError(t, err)
	assert.NotNil(t, record.TransportMethod)
	assert.Equal(t, "kafka", *record.TransportMethod)
	assert.NotNil(t, record.TransportMetadata)
	assert.Equal(t, "topic=events,partition=0", *record.TransportMetadata)
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
		name  string
		value string
	}{
		{"newlines", "Line 1\\nLine 2\\nLine 3"},
		{"tabs", "Column1\\tColumn2\\tColumn3"},
		{"quotes", `He said \"hello\"`},
		{"unicode", "Hello 世界 🌍"},
		{"null bytes", "Before\\u0000After"},
		{"backslashes", "C:\\\\Users\\\\test"},
		{"urls", "https://github.com/OpenCHAMI/legendary-funicular"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			json := `{"ts": "2026-06-10T12:00:00Z", "cloudevent_source": "` + tt.value + `"}`

			record, err := Parse([]byte(json))
			require.NoError(t, err)
			assert.NotNil(t, record.EventSource)
		})
	}
}

// TestParse_LargeEvent tests handling of very large events
func TestParse_LargeEvent(t *testing.T) {
	// Create a 1MB event payload
	largeEvent := make([]byte, 1024*1024)
	for i := range largeEvent {
		largeEvent[i] = 'A'
	}

	json := `{"ts": "2026-06-10T12:00:00Z", "cloudevent": "` + string(largeEvent) + `"}`

	record, err := Parse([]byte(json))
	require.NoError(t, err)
	assert.NotNil(t, record.Event)
	assert.Len(t, *record.Event, 1024*1024)
}

// TestParse_EmptyFields tests handling of empty field values
func TestParse_EmptyFields(t *testing.T) {
	json := `{
		"ts": "",
		"transport_method": "",
		"cloudevent_id": "",
		"cloudevent_source": "",
		"cloudevent_type": ""
	}`

	record, err := Parse([]byte(json))
	require.NoError(t, err)

	// Empty strings should still be set (as empty string pointers)
	assert.NotNil(t, record.Timestamp)
	assert.Equal(t, "", *record.Timestamp)
	assert.NotNil(t, record.TransportMethod)
	assert.Equal(t, "", *record.TransportMethod)
}

// TestParse_NullFields tests handling of null field values
func TestParse_NullFields(t *testing.T) {
	json := `{
		"ts": null,
		"transport_method": null,
		"cloudevent_id": "evt-123",
		"cloudevent_source": null,
		"cloudevent_type": null
	}`

	record, err := Parse([]byte(json))
	require.NoError(t, err)

	// Null values should result in nil pointers
	assert.Nil(t, record.Timestamp)
	assert.Nil(t, record.TransportMethod)
	assert.NotNil(t, record.EventID) // This one is set
	assert.Nil(t, record.EventSource)
	assert.Nil(t, record.EventType)
}

// TestParse_ExtraFields tests that extra fields don't cause errors
func TestParse_ExtraFields(t *testing.T) {
	json := `{
		"ts": "2026-06-10T12:00:00Z",
		"cloudevent_id": "evt-123",
		"unknown_field_1": "value1",
		"unknown_field_2": 12345,
		"unknown_field_3": true,
		"nested_unknown": {"key": "value"}
	}`

	record, err := Parse([]byte(json))
	require.NoError(t, err)
	assert.NotNil(t, record.Timestamp)
	assert.NotNil(t, record.EventID)
}

// TestParse_PayloadJSONPreservation tests that original JSON is preserved
func TestParse_PayloadJSONPreservation(t *testing.T) {
	originalJSON := `{"ts":"2026-06-10T12:00:00Z","cloudevent_id":"evt-123"}`

	record, err := Parse([]byte(originalJSON))
	require.NoError(t, err)
	require.NotNil(t, record.PayloadJSON)
	assert.Equal(t, originalJSON, *record.PayloadJSON)
}

// TestParse_RawField tests the Raw field for datadog vector compatibility
func TestParse_RawField(t *testing.T) {
	json := `{
		"ts": "2026-06-10T12:00:00Z",
		"raw": "original-raw-event-data",
		"cloudevent_id": "evt-123"
	}`

	record, err := Parse([]byte(json))
	require.NoError(t, err)
	assert.NotNil(t, record.Raw)
	assert.Equal(t, "original-raw-event-data", *record.Raw)
}

// TestParse_ParseError tests the ParseError field
func TestParse_ParseError(t *testing.T) {
	json := `{
		"ts": "2026-06-10T12:00:00Z",
		"parse_error": "Failed to decode base64 data",
		"cloudevent_id": "evt-123"
	}`

	record, err := Parse([]byte(json))
	require.NoError(t, err)
	assert.NotNil(t, record.ParseError)
	assert.Equal(t, "Failed to decode base64 data", *record.ParseError)
}

// TestParse_ConcurrentParsing tests thread-safety of Parse
func TestParse_ConcurrentParsing(t *testing.T) {
	json := `{"ts": "2026-06-10T12:00:00Z", "cloudevent_id": "evt-123"}`

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

// BenchmarkParse_Small benchmarks parsing of small CloudEvents
func BenchmarkParse_Small(b *testing.B) {
	json := []byte(`{"ts": "2026-06-10T12:00:00Z", "cloudevent_id": "evt-123", "cloudevent_type": "test"}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Parse(json)
	}
}

// BenchmarkParse_Medium benchmarks parsing of medium CloudEvents
func BenchmarkParse_Medium(b *testing.B) {
	event := make([]byte, 1024) // 1KB event
	for i := range event {
		event[i] = 'A'
	}
	json := []byte(`{"ts": "2026-06-10T12:00:00Z", "cloudevent_id": "evt-123", "cloudevent": "` + string(event) + `"}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Parse(json)
	}
}

// BenchmarkParse_Large benchmarks parsing of large CloudEvents
func BenchmarkParse_Large(b *testing.B) {
	event := make([]byte, 100*1024) // 100KB event
	for i := range event {
		event[i] = 'A'
	}
	json := []byte(`{"ts": "2026-06-10T12:00:00Z", "cloudevent_id": "evt-123", "cloudevent": "` + string(event) + `"}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Parse(json)
	}
}

// FuzzParse fuzzes the Parse function with random inputs
func FuzzParse(f *testing.F) {
	// Seed corpus with valid examples
	f.Add([]byte(`{"ts": "2026-06-10T12:00:00Z"}`))
	f.Add([]byte(`{"ts": "2026-06-10T12:00:00Z", "cloudevent_id": "evt-123"}`))
	f.Add([]byte(`{"ts": "2026-06-10T12:00:00Z", "cloudevent_type": "test.event"}`))
	f.Add([]byte(`{"cloudevent_source": "https://example.com"}`))
	f.Add([]byte(`{}`))
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
