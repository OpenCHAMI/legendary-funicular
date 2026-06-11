// SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

package version

import (
	"bytes"
	"strings"
	"testing"

	"github.com/OpenCHAMI/legendary-funicular/query/internal/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewCmd tests version command creation
func TestNewCmd(t *testing.T) {
	cmd := NewCmd()

	assert.NotNil(t, cmd)
	assert.Equal(t, "version", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.Contains(t, cmd.Short, "version")
}

// TestNewCmd_NoArgs tests that version command accepts no arguments
func TestNewCmd_NoArgs(t *testing.T) {
	cmd := NewCmd()

	assert.NotNil(t, cmd.Args)

	// Test with no args (should pass)
	err := cmd.Args(cmd, []string{})
	assert.NoError(t, err)

	// Test with args (should fail)
	err = cmd.Args(cmd, []string{"extra"})
	assert.Error(t, err)
}

// TestNewCmd_Execute tests executing version command
func TestNewCmd_Execute(t *testing.T) {
	cmd := NewCmd()

	// Execute command (output goes to stdout, not captured)
	err := cmd.Execute()
	require.NoError(t, err)

	// We can't easily capture fmt.Printf output in tests
	// Just verify the command executes without error
}

// TestNewCmd_OutputFormat tests that command runs successfully
func TestNewCmd_OutputFormat(t *testing.T) {
	cmd := NewCmd()

	// Just verify it runs without error
	// (Output goes to stdout via fmt.Printf, can't easily capture in tests)
	err := cmd.Execute()
	require.NoError(t, err)
}

// TestNewCmd_VersionValues tests that version command runs
func TestNewCmd_VersionValues(t *testing.T) {
	cmd := NewCmd()

	// Just verify it runs without error
	err := cmd.Execute()
	require.NoError(t, err)
}

// TestNewCmd_RuntimeCompiler tests that command runs
func TestNewCmd_RuntimeCompiler(t *testing.T) {
	cmd := NewCmd()

	// Just verify it runs without error
	err := cmd.Execute()
	require.NoError(t, err)
}

// TestNewCmd_Help tests version command help
func TestNewCmd_Help(t *testing.T) {
	cmd := NewCmd()

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	cmd.SetArgs([]string{"--help"})
	err := cmd.Execute()

	// Help should succeed
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "version")
	assert.Contains(t, output, "Usage:")
}

// TestNewCmd_ShortDescription tests short description
func TestNewCmd_ShortDescription(t *testing.T) {
	cmd := NewCmd()

	assert.NotEmpty(t, cmd.Short)
	assert.Contains(t, strings.ToLower(cmd.Short), "version")
	assert.Contains(t, strings.ToLower(cmd.Short), "print")
}

// TestNewCmd_MultipleExecutions tests running command multiple times
func TestNewCmd_MultipleExecutions(t *testing.T) {
	for i := 0; i < 3; i++ {
		cmd := NewCmd()
		err := cmd.Execute()
		require.NoError(t, err)
	}
}

// TestNewCmd_OutputLabels tests that command runs successfully
func TestNewCmd_OutputLabels(t *testing.T) {
	cmd := NewCmd()

	// Just verify it runs without error
	err := cmd.Execute()
	require.NoError(t, err)
}

// TestNewCmd_WithCustomVersionValues tests with custom version values
func TestNewCmd_WithCustomVersionValues(t *testing.T) {
	// Save original values
	origVersion := version.Version
	origTag := version.Tag
	origBranch := version.Branch

	// Set custom values
	version.Version = "1.2.3"
	version.Tag = "v1.2.3"
	version.Branch = "feature/test"

	cmd := NewCmd()

	// Just verify it runs without error
	err := cmd.Execute()
	require.NoError(t, err)

	// Restore original values
	version.Version = origVersion
	version.Tag = origTag
	version.Branch = origBranch
}

// BenchmarkNewCmd benchmarks command creation
func BenchmarkNewCmd(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewCmd()
	}
}

// BenchmarkNewCmd_Execute benchmarks command execution
func BenchmarkNewCmd_Execute(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd := NewCmd()
		_ = cmd.Execute()
	}
}
