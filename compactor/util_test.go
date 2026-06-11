// SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

package main

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCheckErr_NoError tests checkErr with nil error
func TestCheckErr_NoError(t *testing.T) {
	// Should not exit or log anything
	checkErr(nil, true)
	checkErr(nil, false)
	// If we get here, test passed
}

// TestCheckErr_WithError_NoExit tests checkErr logs but doesn't exit
func TestCheckErr_WithError_NoExit(t *testing.T) {
	err := errors.New("test error")

	// With shouldExit=false, this should log a warning but not exit
	checkErr(err, false)
	// If we get here, test passed (didn't exit)
}

// TestCheckErr_WithError_ShouldExit tests checkErr would exit
func TestCheckErr_WithError_ShouldExit(t *testing.T) {
	t.Skip("Cannot test os.Exit() behavior without subprocess or mock")

	// If we could test this, it would:
	// 1. Call checkErr with shouldExit=true
	// 2. Verify it logs an error via slog
	// 3. Verify it calls os.Exit(1)
}

// TestGetRequiredEnv_ExistingVar tests getRequiredEnv with set variable
func TestGetRequiredEnv_ExistingVar(t *testing.T) {
	_ = os.Setenv("COMPACTOR_TEST_VAR", "compactor-value")
	defer func() { _ = os.Unsetenv("COMPACTOR_TEST_VAR") }()

	value := getRequiredEnv("COMPACTOR_TEST_VAR")
	assert.Equal(t, "compactor-value", value)
}

// TestGetRequiredEnv_MissingVar tests getRequiredEnv would exit
func TestGetRequiredEnv_MissingVar(t *testing.T) {
	t.Skip("Cannot test os.Exit() behavior without subprocess or mock")

	// If we could test this, it would:
	// 1. Call getRequiredEnv with missing var
	// 2. Verify it logs an error via slog
	// 3. Verify it calls os.Exit(1)
}

// TestGetRequiredEnv_EmptyVar tests getRequiredEnv treats empty as missing
func TestGetRequiredEnv_EmptyVar(t *testing.T) {
	t.Skip("Cannot test os.Exit() behavior without subprocess or mock")

	// Empty value should be treated as missing and cause exit
}

// TestGetRequiredEnv_MultipleVars tests getRequiredEnv with various values
func TestGetRequiredEnv_MultipleVars(t *testing.T) {
	tests := []struct {
		name     string
		varName  string
		varValue string
	}{
		{"s3 access key", "TEST_S3_ACCESS", "AKIAIOSFODNN7EXAMPLE"},
		{"s3 secret key", "TEST_S3_SECRET", "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"},
		{"s3 region", "TEST_S3_REGION", "us-west-2"},
		{"s3 endpoint", "TEST_S3_ENDPOINT", "http://versitygw:7070"},
		{"bucket name", "TEST_S3_BUCKET", "openchami-logs-raw"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = os.Setenv(tt.varName, tt.varValue)
			defer func() { _ = os.Unsetenv(tt.varName) }()

			value := getRequiredEnv(tt.varName)
			assert.Equal(t, tt.varValue, value)
		})
	}
}

// TestBuildObjectKey tests buildObjectKey generates valid keys
func TestBuildObjectKey(t *testing.T) {
	prefix := "logs"

	key := buildObjectKey(prefix)

	// Verify format: prefix/date=YYYY-MM-DD/uuid.parquet
	assert.Contains(t, key, prefix+"/date=")
	assert.Contains(t, key, ".parquet")
	assert.Regexp(t, `^logs/date=\d{4}-\d{2}-\d{2}/[a-f0-9-]+\.parquet$`, key)
}

// TestBuildObjectKey_DifferentPrefixes tests various prefixes
func TestBuildObjectKey_DifferentPrefixes(t *testing.T) {
	prefixes := []string{"logs", "events", "metrics", "traces"}

	for _, prefix := range prefixes {
		t.Run(prefix, func(t *testing.T) {
			key := buildObjectKey(prefix)
			assert.Contains(t, key, prefix+"/date=")
			assert.Contains(t, key, ".parquet")
		})
	}
}

// TestBuildObjectKey_UniqueKeys tests that keys are unique
func TestBuildObjectKey_UniqueKeys(t *testing.T) {
	keys := make(map[string]bool)

	// Generate 100 keys and verify they're all unique
	for i := 0; i < 100; i++ {
		key := buildObjectKey("logs")
		assert.False(t, keys[key], "Generated duplicate key: %s", key)
		keys[key] = true
	}

	assert.Len(t, keys, 100, "Should have 100 unique keys")
}

// BenchmarkGetRequiredEnv benchmarks getRequiredEnv
func BenchmarkGetRequiredEnv(b *testing.B) {
	_ = os.Setenv("BENCH_REQUIRED_VAR", "bench-value")
	defer func() { _ = os.Unsetenv("BENCH_REQUIRED_VAR") }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = getRequiredEnv("BENCH_REQUIRED_VAR")
	}
}

// BenchmarkBuildObjectKey benchmarks buildObjectKey
func BenchmarkBuildObjectKey(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = buildObjectKey("logs")
	}
}

// BenchmarkCheckErr benchmarks checkErr with no error
func BenchmarkCheckErr_NoError(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		checkErr(nil, false)
	}
}

// BenchmarkCheckErr benchmarks checkErr with error (no exit)
func BenchmarkCheckErr_WithError(b *testing.B) {
	err := errors.New("benchmark error")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		checkErr(err, false)
	}
}
