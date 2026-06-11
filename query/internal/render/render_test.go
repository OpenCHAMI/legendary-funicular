// SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

package render

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test structs
type TestRecord struct {
	ID      int    `json:"id"`
	Message string `json:"message"`
}

type TestRecordWithSecret struct {
	ID       int     `json:"id"`
	Username string  `json:"username"`
	Password *string `json:"password" secret:"true"`
	APIKey   *string `json:"api_key" secret:"true"`
}

// TestBuildEncoder_NDJSON tests building NDJSON encoder
func TestBuildEncoder_NDJSON(t *testing.T) {
	var buf bytes.Buffer

	enc, err := BuildEncoder(&buf, "ndjson")

	require.NoError(t, err)
	assert.NotNil(t, enc)
}

// TestBuildEncoder_JSON tests building JSON encoder
func TestBuildEncoder_JSON(t *testing.T) {
	var buf bytes.Buffer

	enc, err := BuildEncoder(&buf, "json")

	require.NoError(t, err)
	assert.NotNil(t, enc)
	assert.Equal(t, "[", buf.String()) // Should write opening bracket
}

// TestBuildEncoder_CSV tests CSV encoder (not implemented)
func TestBuildEncoder_CSV(t *testing.T) {
	var buf bytes.Buffer

	enc, err := BuildEncoder(&buf, "csv")

	assert.Error(t, err)
	assert.Nil(t, enc)
}

// TestBuildEncoder_Text tests text encoder (not implemented)
func TestBuildEncoder_Text(t *testing.T) {
	var buf bytes.Buffer

	enc, err := BuildEncoder(&buf, "text")

	assert.Error(t, err)
	assert.Nil(t, enc)
}

// TestBuildEncoder_UnknownFormat tests unknown format error
func TestBuildEncoder_UnknownFormat(t *testing.T) {
	var buf bytes.Buffer

	enc, err := BuildEncoder(&buf, "xml")

	assert.Error(t, err)
	assert.Nil(t, enc)
	assert.Contains(t, err.Error(), "unknown format: xml")
}

// TestRender_Struct_NDJSON tests rendering a single struct as NDJSON
func TestRender_Struct_NDJSON(t *testing.T) {
	var buf bytes.Buffer
	record := TestRecord{ID: 1, Message: "test"}

	err := Render(&buf, "ndjson", record)

	require.NoError(t, err)

	// Should be single line JSON
	output := buf.String()
	assert.Contains(t, output, `"id":1`)
	assert.Contains(t, output, `"message":"test"`)
	assert.True(t, strings.HasSuffix(output, "\n"))
}

// TestRender_Struct_JSON tests rendering a single struct as JSON
func TestRender_Struct_JSON(t *testing.T) {
	var buf bytes.Buffer
	record := TestRecord{ID: 1, Message: "test"}

	err := Render(&buf, "json", record)

	require.NoError(t, err)

	// Should be wrapped in array
	output := buf.String()
	assert.True(t, strings.HasPrefix(output, "["))
	assert.True(t, strings.HasSuffix(strings.TrimSpace(output), "]"))
	assert.Contains(t, output, `"id":1`)
	assert.Contains(t, output, `"message":"test"`)
}

// TestRender_Slice_NDJSON tests rendering a slice as NDJSON
func TestRender_Slice_NDJSON(t *testing.T) {
	var buf bytes.Buffer
	records := []TestRecord{
		{ID: 1, Message: "first"},
		{ID: 2, Message: "second"},
		{ID: 3, Message: "third"},
	}

	err := Render(&buf, "ndjson", records)

	require.NoError(t, err)

	// Should be multiple lines
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	assert.Len(t, lines, 3)

	// Verify each line is valid JSON
	for i, line := range lines {
		var record TestRecord
		err := json.Unmarshal([]byte(line), &record)
		require.NoError(t, err)
		assert.Equal(t, i+1, record.ID)
	}
}

// TestRender_Slice_JSON tests rendering a slice as JSON
func TestRender_Slice_JSON(t *testing.T) {
	var buf bytes.Buffer
	records := []TestRecord{
		{ID: 1, Message: "first"},
		{ID: 2, Message: "second"},
		{ID: 3, Message: "third"},
	}

	err := Render(&buf, "json", records)

	require.NoError(t, err)

	// Should be valid JSON array
	var decoded []TestRecord
	err = json.Unmarshal(buf.Bytes(), &decoded)
	require.NoError(t, err)
	assert.Len(t, decoded, 3)
	assert.Equal(t, 1, decoded[0].ID)
	assert.Equal(t, "first", decoded[0].Message)
}

// TestRender_EmptySlice tests rendering empty slice
func TestRender_EmptySlice(t *testing.T) {
	var buf bytes.Buffer
	records := []TestRecord{}

	err := Render(&buf, "json", records)

	require.NoError(t, err)

	// Should be empty array
	var decoded []TestRecord
	err = json.Unmarshal(buf.Bytes(), &decoded)
	require.NoError(t, err)
	assert.Len(t, decoded, 0)
}

// TestRender_PointerToStruct tests rendering pointer to struct
func TestRender_PointerToStruct(t *testing.T) {
	var buf bytes.Buffer
	record := &TestRecord{ID: 1, Message: "test"}

	err := Render(&buf, "ndjson", record)

	require.NoError(t, err)
	assert.Contains(t, buf.String(), `"id":1`)
}

// TestRender_UnknownFormat tests unknown format error
func TestRender_UnknownFormat(t *testing.T) {
	var buf bytes.Buffer
	record := TestRecord{ID: 1, Message: "test"}

	err := Render(&buf, "yaml", record)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown format: yaml")
}

// TestRender_InvalidType tests rendering invalid type
func TestRender_InvalidType(t *testing.T) {
	var buf bytes.Buffer

	tests := []struct {
		name  string
		value any
	}{
		{"string", "invalid"},
		{"int", 42},
		{"map", map[string]string{"key": "value"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			err := Render(&buf, "json", tt.value)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "expected struct or slice of struct")
		})
	}
}

// TestMaskSecrets_WithSecrets tests secret masking
func TestMaskSecrets_WithSecrets(t *testing.T) {
	password := "super-secret"
	apiKey := "api-key-12345"
	record := &TestRecordWithSecret{
		ID:       1,
		Username: "user",
		Password: &password,
		APIKey:   &apiKey,
	}

	MaskSecrets(record)

	assert.Equal(t, 1, record.ID)
	assert.Equal(t, "user", record.Username)
	assert.Equal(t, "************************", *record.Password)
	assert.Equal(t, "************************", *record.APIKey)
}

// TestMaskSecrets_NilSecrets tests masking with nil secrets
func TestMaskSecrets_NilSecrets(t *testing.T) {
	record := &TestRecordWithSecret{
		ID:       1,
		Username: "user",
		Password: nil,
		APIKey:   nil,
	}

	MaskSecrets(record)

	assert.Equal(t, 1, record.ID)
	assert.Equal(t, "user", record.Username)
	assert.Nil(t, record.Password)
	assert.Nil(t, record.APIKey)
}

// TestMaskSecrets_NonPointer tests masking with non-pointer (no-op)
func TestMaskSecrets_NonPointer(t *testing.T) {
	password := "super-secret"
	record := TestRecordWithSecret{
		ID:       1,
		Username: "user",
		Password: &password,
	}

	// Should not panic, but won't mask (not a pointer)
	MaskSecrets(record)

	// Original should be unchanged
	assert.Equal(t, "super-secret", *record.Password)
}

// TestMaskSecrets_NilPointer tests masking with nil pointer (no-op)
func TestMaskSecrets_NilPointer(t *testing.T) {
	var record *TestRecordWithSecret

	// Should not panic
	MaskSecrets(record)
}

// TestMaskSecrets_NonStruct tests masking non-struct (no-op)
func TestMaskSecrets_NonStruct(t *testing.T) {
	value := 42

	// Should not panic
	MaskSecrets(&value)
}

// TestEncoderJSON_MultipleEncodes tests JSON encoder with multiple records
func TestEncoderJSON_MultipleEncodes(t *testing.T) {
	var buf bytes.Buffer

	enc, err := newEncoderJSON(&buf)
	require.NoError(t, err)

	// Encode multiple records
	err = enc.Encode(TestRecord{ID: 1, Message: "first"})
	require.NoError(t, err)

	err = enc.Encode(TestRecord{ID: 2, Message: "second"})
	require.NoError(t, err)

	err = enc.Encode(TestRecord{ID: 3, Message: "third"})
	require.NoError(t, err)

	err = enc.Close()
	require.NoError(t, err)

	// Should be valid JSON array
	var decoded []TestRecord
	err = json.Unmarshal(buf.Bytes(), &decoded)
	require.NoError(t, err)
	assert.Len(t, decoded, 3)
}

// TestEncoderJSON_EmptyArray tests JSON encoder with no records
func TestEncoderJSON_EmptyArray(t *testing.T) {
	var buf bytes.Buffer

	enc, err := newEncoderJSON(&buf)
	require.NoError(t, err)

	err = enc.Close()
	require.NoError(t, err)

	// Should be empty array
	assert.Equal(t, "[]", buf.String())
}

// TestEncoderNDJSON_MultipleEncodes tests NDJSON encoder with multiple records
func TestEncoderNDJSON_MultipleEncodes(t *testing.T) {
	var buf bytes.Buffer

	enc, err := newEncoderNDJSON(&buf)
	require.NoError(t, err)

	// Encode multiple records
	err = enc.Encode(TestRecord{ID: 1, Message: "first"})
	require.NoError(t, err)

	err = enc.Encode(TestRecord{ID: 2, Message: "second"})
	require.NoError(t, err)

	err = enc.Close()
	require.NoError(t, err)

	// Should be two lines
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	assert.Len(t, lines, 2)
}

// TestMessage_Struct tests Message struct
func TestMessage_Struct(t *testing.T) {
	msg := Message{
		Message: "Operation successful",
		Code:    "SUCCESS",
	}

	data, err := json.Marshal(msg)
	require.NoError(t, err)

	var decoded Message
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "Operation successful", decoded.Message)
	assert.Equal(t, "SUCCESS", decoded.Code)
}

// TestMessage_OmitEmpty tests Message omitempty tags
func TestMessage_OmitEmpty(t *testing.T) {
	msg := Message{
		Message: "test",
	}

	data, err := json.Marshal(msg)
	require.NoError(t, err)

	// Code should be omitted
	assert.Contains(t, string(data), `"message":"test"`)
	assert.NotContains(t, string(data), `"code"`)
}

// BenchmarkRender_NDJSON_SingleStruct benchmarks NDJSON rendering
func BenchmarkRender_NDJSON_SingleStruct(b *testing.B) {
	var buf bytes.Buffer
	record := TestRecord{ID: 1, Message: "benchmark test"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = Render(&buf, "ndjson", record)
	}
}

// BenchmarkRender_JSON_SingleStruct benchmarks JSON rendering
func BenchmarkRender_JSON_SingleStruct(b *testing.B) {
	var buf bytes.Buffer
	record := TestRecord{ID: 1, Message: "benchmark test"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = Render(&buf, "json", record)
	}
}

// BenchmarkRender_NDJSON_Slice benchmarks NDJSON slice rendering
func BenchmarkRender_NDJSON_Slice(b *testing.B) {
	var buf bytes.Buffer
	records := make([]TestRecord, 100)
	for i := range records {
		records[i] = TestRecord{ID: i, Message: "benchmark test"}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = Render(&buf, "ndjson", records)
	}
}

// BenchmarkRender_JSON_Slice benchmarks JSON slice rendering
func BenchmarkRender_JSON_Slice(b *testing.B) {
	var buf bytes.Buffer
	records := make([]TestRecord, 100)
	for i := range records {
		records[i] = TestRecord{ID: i, Message: "benchmark test"}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = Render(&buf, "json", records)
	}
}

// BenchmarkMaskSecrets benchmarks secret masking
func BenchmarkMaskSecrets(b *testing.B) {
	password := "super-secret"
	apiKey := "api-key-12345"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		record := &TestRecordWithSecret{
			ID:       1,
			Username: "user",
			Password: &password,
			APIKey:   &apiKey,
		}
		MaskSecrets(record)
	}
}

// BenchmarkEncoderJSON_Stream benchmarks streaming JSON encoding
func BenchmarkEncoderJSON_Stream(b *testing.B) {
	var buf bytes.Buffer

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		enc, _ := newEncoderJSON(&buf)
		for j := 0; j < 100; j++ {
			_ = enc.Encode(TestRecord{ID: j, Message: "test"})
		}
		_ = enc.Close()
	}
}

// BenchmarkEncoderNDJSON_Stream benchmarks streaming NDJSON encoding
func BenchmarkEncoderNDJSON_Stream(b *testing.B) {
	var buf bytes.Buffer

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		enc, _ := newEncoderNDJSON(&buf)
		for j := 0; j < 100; j++ {
			_ = enc.Encode(TestRecord{ID: j, Message: "test"})
		}
		_ = enc.Close()
	}
}
