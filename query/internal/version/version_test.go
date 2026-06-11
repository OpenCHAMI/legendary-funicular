// SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestProgName tests the program name constant
func TestProgName(t *testing.T) {
	assert.Equal(t, "openchami-logq", ProgName)
	assert.NotEmpty(t, ProgName)
}

// TestVersionVariables tests that version variables exist
func TestVersionVariables(t *testing.T) {
	// All variables should be initialized (even if "unknown")
	assert.NotEmpty(t, Version)
	assert.NotEmpty(t, Tag)
	assert.NotEmpty(t, Branch)
	assert.NotEmpty(t, Commit)
	assert.NotEmpty(t, Date)
	assert.NotEmpty(t, GoVersion)
	assert.NotEmpty(t, GitState)
	assert.NotEmpty(t, BuildHost)
	assert.NotEmpty(t, BuildUser)
}

// TestVersionDefaults tests default values
func TestVersionDefaults(t *testing.T) {
	// Default values should be "unknown" unless set by build
	// This tests the package initialization
	tests := []struct {
		name  string
		value string
	}{
		{"Version", Version},
		{"Tag", Tag},
		{"Branch", Branch},
		{"Commit", Commit},
		{"Date", Date},
		{"GoVersion", GoVersion},
		{"GitState", GitState},
		{"BuildHost", BuildHost},
		{"BuildUser", BuildUser},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should be set to something (either "unknown" or actual value from build)
			assert.NotEmpty(t, tt.value, "%s should not be empty", tt.name)
		})
	}
}

// TestVersionVariablesAreExported tests that variables can be set
func TestVersionVariablesAreExported(t *testing.T) {
	// Save original values
	origVersion := Version
	origTag := Tag
	origBranch := Branch

	// Test that we can modify them (they're exported vars, not constants)
	Version = "1.0.0"
	Tag = "v1.0.0"
	Branch = "main"

	assert.Equal(t, "1.0.0", Version)
	assert.Equal(t, "v1.0.0", Tag)
	assert.Equal(t, "main", Branch)

	// Restore original values
	Version = origVersion
	Tag = origTag
	Branch = origBranch
}

// TestProgNameIsConstant tests that ProgName cannot be changed
func TestProgNameIsConstant(t *testing.T) {
	// This is a compile-time check - if ProgName is const, this test compiles
	const testName = ProgName
	assert.Equal(t, "openchami-logq", testName)
}

// TestVersionFormat tests version string format
func TestVersionFormat(t *testing.T) {
	// Version should either be "unknown" or follow semantic versioning
	if Version != "unknown" {
		// If set, should not be empty
		assert.NotEmpty(t, Version)
		// Could add more validation for semver format if needed
	}
}

// TestCommitFormat tests commit hash format
func TestCommitFormat(t *testing.T) {
	// Commit should either be "unknown" or a git hash
	if Commit != "unknown" {
		// If set, should be a reasonable length (git short hash is 7+ chars)
		assert.GreaterOrEqual(t, len(Commit), 7, "Commit hash should be at least 7 characters")
	}
}

// TestGitStateValues tests git state possible values
func TestGitStateValues(t *testing.T) {
	// GitState should be "unknown", "clean", or "dirty"
	validStates := []string{"unknown", "clean", "dirty"}
	assert.Contains(t, validStates, GitState, "GitState should be one of: %v", validStates)
}

// BenchmarkVersionAccess benchmarks accessing version variables
func BenchmarkVersionAccess(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Version
		_ = Tag
		_ = Branch
		_ = Commit
	}
}
