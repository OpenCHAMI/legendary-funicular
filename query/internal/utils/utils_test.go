// SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

package utils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGetEnv_ExistingVar tests GetEnv with a set environment variable
func TestGetEnv_ExistingVar(t *testing.T) {
	_ = os.Setenv("TEST_VAR", "test-value")
	defer func() { _ = os.Unsetenv("TEST_VAR") }()

	value, err := GetEnv("TEST_VAR")
	assert.NoError(t, err)
	assert.Equal(t, "test-value", value)
}

// TestGetEnv_MissingVar tests GetEnv with an unset environment variable
func TestGetEnv_MissingVar(t *testing.T) {
	_ = os.Unsetenv("NONEXISTENT_VAR")

	value, err := GetEnv("NONEXISTENT_VAR")
	assert.Error(t, err)
	assert.Empty(t, value)
	assert.Contains(t, err.Error(), "undefined required environment variable")
	assert.Contains(t, err.Error(), "NONEXISTENT_VAR")
}

// TestGetEnv_EmptyVar tests GetEnv with an empty environment variable
func TestGetEnv_EmptyVar(t *testing.T) {
	_ = os.Setenv("EMPTY_VAR", "")
	defer func() { _ = os.Unsetenv("EMPTY_VAR") }()

	value, err := GetEnv("EMPTY_VAR")
	assert.Error(t, err, "Empty value should be treated as missing")
	assert.Empty(t, value)
}

// TestGetEnv_MultipleVars tests GetEnv with multiple variables
func TestGetEnv_MultipleVars(t *testing.T) {
	tests := []struct {
		name     string
		varName  string
		varValue string
		wantErr  bool
	}{
		{"valid string", "VAR1", "value1", false},
		{"valid number", "VAR2", "12345", false},
		{"valid path", "VAR3", "/path/to/file", false},
		{"valid url", "VAR4", "http://localhost:9000", false},
		{"missing", "VAR5", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.varValue != "" {
				_ = os.Setenv(tt.varName, tt.varValue)
				defer func() { _ = os.Unsetenv(tt.varName) }()
			} else {
				_ = os.Unsetenv(tt.varName)
			}

			value, err := GetEnv(tt.varName)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, value)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.varValue, value)
			}
		})
	}
}

// TestGetEnvFatal_ExistingVar tests GetEnvFatal with a set variable
func TestGetEnvFatal_ExistingVar(t *testing.T) {
	_ = os.Setenv("FATAL_TEST_VAR", "fatal-value")
	defer func() { _ = os.Unsetenv("FATAL_TEST_VAR") }()

	// This should not exit
	value := GetEnvFatal("FATAL_TEST_VAR")
	assert.Equal(t, "fatal-value", value)
}

// TestGetEnvFatal_MissingVar tests that GetEnvFatal would exit
// Note: We can't actually test the exit behavior without mocking os.Exit
// or using a subprocess. This test documents the expected behavior.
func TestGetEnvFatal_MissingVar(t *testing.T) {
	t.Skip("Cannot test os.Exit() behavior without subprocess or mock")

	// If we could test this, it would:
	// 1. Call GetEnvFatal with missing var
	// 2. Verify it logs an error
	// 3. Verify it calls os.Exit(1)

	// To properly test this, we'd need to refactor GetEnvFatal to accept
	// a dependency-injected exit function, or test it via subprocess.
}

// TestCheckFatal_NoError tests CheckFatal with nil error
func TestCheckFatal_NoError(t *testing.T) {
	// This should not exit or panic
	CheckFatal(nil)
	// If we get here, test passed
}

// TestCheckFatal_WithError tests that CheckFatal would exit
// Note: Same limitation as TestGetEnvFatal_MissingVar
func TestCheckFatal_WithError(t *testing.T) {
	t.Skip("Cannot test os.Exit() behavior without subprocess or mock")

	// If we could test this, it would:
	// 1. Call CheckFatal with an error
	// 2. Verify it logs the error via slog
	// 3. Verify it calls os.Exit(1)
}

// BenchmarkGetEnv_ExistingVar benchmarks GetEnv with existing variable
func BenchmarkGetEnv_ExistingVar(b *testing.B) {
	_ = os.Setenv("BENCH_VAR", "bench-value")
	defer func() { _ = os.Unsetenv("BENCH_VAR") }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetEnv("BENCH_VAR")
	}
}

// BenchmarkGetEnv_MissingVar benchmarks GetEnv with missing variable
func BenchmarkGetEnv_MissingVar(b *testing.B) {
	_ = os.Unsetenv("MISSING_BENCH_VAR")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetEnv("MISSING_BENCH_VAR")
	}
}

// BenchmarkGetEnvFatal benchmarks GetEnvFatal with existing variable
func BenchmarkGetEnvFatal(b *testing.B) {
	_ = os.Setenv("BENCH_FATAL_VAR", "bench-fatal-value")
	defer func() { _ = os.Unsetenv("BENCH_FATAL_VAR") }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetEnvFatal("BENCH_FATAL_VAR")
	}
}
