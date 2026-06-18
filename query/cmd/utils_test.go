// SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBuildCfg_WithAllEnvVars tests BuildCfg with all environment variables set
func TestBuildCfg_WithAllEnvVars(t *testing.T) {
	// Set up all required environment variables
	_ = os.Setenv("S3_ACCESS_KEY", "test-access-key")
	_ = os.Setenv("S3_SECRET_KEY", "test-secret-key")
	_ = os.Setenv("S3_REGION", "us-west-2")
	_ = os.Setenv("S3_ENDPOINT", "http://localhost:9000")
	_ = os.Setenv("S3_SSL", "true")
	_ = os.Setenv("S3_BUCKET_NDJSON", "test-ndjson-bucket")
	_ = os.Setenv("S3_BUCKET_PARQUET", "test-parquet-bucket")

	defer func() {
		_ = os.Unsetenv("S3_ACCESS_KEY")
		_ = os.Unsetenv("S3_SECRET_KEY")
		_ = os.Unsetenv("S3_REGION")
		_ = os.Unsetenv("S3_ENDPOINT")
		_ = os.Unsetenv("S3_SSL")
		_ = os.Unsetenv("S3_BUCKET_NDJSON")
		_ = os.Unsetenv("S3_BUCKET_PARQUET")
	}()

	cfg := BuildCfg()

	require.NotNil(t, cfg)

	// Verify required fields from env vars
	assert.NotNil(t, cfg.S3KeyAccess)
	assert.Equal(t, "test-access-key", *cfg.S3KeyAccess)
	assert.NotNil(t, cfg.S3KeySecret)
	assert.Equal(t, "test-secret-key", *cfg.S3KeySecret)

	// Verify optional fields that were set
	assert.NotNil(t, cfg.S3Region)
	assert.Equal(t, "us-west-2", *cfg.S3Region)
	assert.NotNil(t, cfg.S3Endpoint)
	assert.Equal(t, "http://localhost:9000", *cfg.S3Endpoint)
	assert.True(t, cfg.S3SSL)
	assert.NotNil(t, cfg.S3BucketNDJSON)
	assert.Equal(t, "test-ndjson-bucket", *cfg.S3BucketNDJSON)
	assert.NotNil(t, cfg.S3BucketParquet)
	assert.Equal(t, "test-parquet-bucket", *cfg.S3BucketParquet)
}

// TestBuildCfg_WithMinimalEnvVars tests BuildCfg with only required env vars
func TestBuildCfg_WithMinimalEnvVars(t *testing.T) {
	// Set up only required environment variables
	_ = os.Setenv("S3_ACCESS_KEY", "minimal-access-key")
	_ = os.Setenv("S3_SECRET_KEY", "minimal-secret-key")

	defer func() {
		_ = os.Unsetenv("S3_ACCESS_KEY")
		_ = os.Unsetenv("S3_SECRET_KEY")
	}()

	cfg := BuildCfg()

	require.NotNil(t, cfg)

	// Verify required fields
	assert.NotNil(t, cfg.S3KeyAccess)
	assert.Equal(t, "minimal-access-key", *cfg.S3KeyAccess)
	assert.NotNil(t, cfg.S3KeySecret)
	assert.Equal(t, "minimal-secret-key", *cfg.S3KeySecret)

	// Verify defaults are applied
	assert.NotNil(t, cfg.S3Region)
	assert.Equal(t, "us-east-1", *cfg.S3Region) // Default
	assert.NotNil(t, cfg.S3Endpoint)
	assert.Equal(t, "localhost:7070", *cfg.S3Endpoint) // Default
	assert.False(t, cfg.S3SSL)                         // Default
	assert.NotNil(t, cfg.S3BucketNDJSON)
	assert.Equal(t, "openchami-logs-raw", *cfg.S3BucketNDJSON) // Default
	assert.NotNil(t, cfg.S3BucketParquet)
	assert.Equal(t, "openchami-logs-daily", *cfg.S3BucketParquet) // Default
}

// TestBuildCfg_EnvOverridesDefaults tests that env vars override defaults
func TestBuildCfg_EnvOverridesDefaults(t *testing.T) {
	// Set required vars plus one optional to override default
	_ = os.Setenv("S3_ACCESS_KEY", "override-access-key")
	_ = os.Setenv("S3_SECRET_KEY", "override-secret-key")
	_ = os.Setenv("S3_REGION", "eu-west-1") // Override default us-east-1

	defer func() {
		_ = os.Unsetenv("S3_ACCESS_KEY")
		_ = os.Unsetenv("S3_SECRET_KEY")
		_ = os.Unsetenv("S3_REGION")
	}()

	cfg := BuildCfg()

	require.NotNil(t, cfg)

	// Verify override worked
	assert.NotNil(t, cfg.S3Region)
	assert.Equal(t, "eu-west-1", *cfg.S3Region)

	// Verify other defaults still applied
	assert.NotNil(t, cfg.S3Endpoint)
	assert.Equal(t, "localhost:7070", *cfg.S3Endpoint)
}

// TestBuildCfg_SSLParsing tests SSL boolean parsing
func TestBuildCfg_SSLParsing(t *testing.T) {
	tests := []struct {
		name     string
		sslValue string
		expected bool
	}{
		{"SSL true", "true", true},
		{"SSL false", "false", false},
		{"SSL empty", "", false},      // Empty means not set, default applies
		{"SSL invalid", "yes", false}, // Invalid value, not "true"
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = os.Setenv("S3_ACCESS_KEY", "test-access")
			_ = os.Setenv("S3_SECRET_KEY", "test-secret")
			if tt.sslValue != "" {
				_ = os.Setenv("S3_SSL", tt.sslValue)
			}

			defer func() {
				_ = os.Unsetenv("S3_ACCESS_KEY")
				_ = os.Unsetenv("S3_SECRET_KEY")
				_ = os.Unsetenv("S3_SSL")
			}()

			cfg := BuildCfg()
			assert.Equal(t, tt.expected, cfg.S3SSL)
		})
	}
}

// TestBuildCfg_EmptyEnvVars tests behavior with empty environment variables
func TestBuildCfg_EmptyEnvVars(t *testing.T) {
	// Set required vars
	_ = os.Setenv("S3_ACCESS_KEY", "test-access")
	_ = os.Setenv("S3_SECRET_KEY", "test-secret")

	// Set optional vars to empty (should be ignored)
	_ = os.Setenv("S3_REGION", "")
	_ = os.Setenv("S3_ENDPOINT", "")

	defer func() {
		_ = os.Unsetenv("S3_ACCESS_KEY")
		_ = os.Unsetenv("S3_SECRET_KEY")
		_ = os.Unsetenv("S3_REGION")
		_ = os.Unsetenv("S3_ENDPOINT")
	}()

	cfg := BuildCfg()

	require.NotNil(t, cfg)

	// Empty env vars should be ignored, defaults should apply
	assert.NotNil(t, cfg.S3Region)
	assert.Equal(t, "us-east-1", *cfg.S3Region) // Default, not empty
	assert.NotNil(t, cfg.S3Endpoint)
	assert.Equal(t, "localhost:7070", *cfg.S3Endpoint) // Default, not empty
}

// TestBuildCfg_OptionChainOrder tests that options are applied in correct order
func TestBuildCfg_OptionChainOrder(t *testing.T) {
	// Set all env vars
	_ = os.Setenv("S3_ACCESS_KEY", "test-access")
	_ = os.Setenv("S3_SECRET_KEY", "test-secret")
	_ = os.Setenv("S3_REGION", "us-west-2")
	_ = os.Setenv("S3_ENDPOINT", "http://custom:9000")
	_ = os.Setenv("S3_BUCKET_NDJSON", "custom-ndjson")
	_ = os.Setenv("S3_BUCKET_PARQUET", "custom-parquet")

	defer func() {
		_ = os.Unsetenv("S3_ACCESS_KEY")
		_ = os.Unsetenv("S3_SECRET_KEY")
		_ = os.Unsetenv("S3_REGION")
		_ = os.Unsetenv("S3_ENDPOINT")
		_ = os.Unsetenv("S3_BUCKET_NDJSON")
		_ = os.Unsetenv("S3_BUCKET_PARQUET")
	}()

	cfg := BuildCfg()

	require.NotNil(t, cfg)

	// Verify that env vars override defaults (option chain: defaults first, then env vars)
	assert.Equal(t, "us-west-2", *cfg.S3Region)             // Env var, not default us-east-1
	assert.Equal(t, "http://custom:9000", *cfg.S3Endpoint)  // Env var, not default localhost:7070
	assert.Equal(t, "custom-ndjson", *cfg.S3BucketNDJSON)   // Env var, not default
	assert.Equal(t, "custom-parquet", *cfg.S3BucketParquet) // Env var, not default
}

// TestBuildCfg_RealWorldScenarios tests realistic deployment scenarios
func TestBuildCfg_RealWorldScenarios(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		description string
	}{
		{
			name: "Local development",
			envVars: map[string]string{
				"S3_ACCESS_KEY": "minioadmin",
				"S3_SECRET_KEY": "minioadmin",
			},
			description: "Local MinIO with defaults",
		},
		{
			name: "Production AWS",
			envVars: map[string]string{
				"S3_ACCESS_KEY":     "AKIAIOSFODNN7EXAMPLE",
				"S3_SECRET_KEY":     "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
				"S3_REGION":         "us-east-1",
				"S3_ENDPOINT":       "https://s3.us-east-1.amazonaws.com",
				"S3_SSL":            "true",
				"S3_BUCKET_NDJSON":  "prod-logs-raw",
				"S3_BUCKET_PARQUET": "prod-logs-daily",
			},
			description: "Production AWS S3 with SSL",
		},
		{
			name: "Staging environment",
			envVars: map[string]string{
				"S3_ACCESS_KEY":     "staging-access",
				"S3_SECRET_KEY":     "staging-secret",
				"S3_REGION":         "us-west-2",
				"S3_ENDPOINT":       "http://staging-s3:9000",
				"S3_BUCKET_NDJSON":  "staging-logs-raw",
				"S3_BUCKET_PARQUET": "staging-logs-daily",
			},
			description: "Staging with custom buckets",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tt.envVars {
				_ = os.Setenv(key, value)
			}

			defer func() {
				for key := range tt.envVars {
					_ = os.Unsetenv(key)
				}
			}()

			cfg := BuildCfg()
			require.NotNil(t, cfg, tt.description)

			// Verify required fields are set
			assert.NotNil(t, cfg.S3KeyAccess)
			assert.NotNil(t, cfg.S3KeySecret)
			assert.NotNil(t, cfg.S3Region)
			assert.NotNil(t, cfg.S3Endpoint)
			assert.NotNil(t, cfg.S3BucketNDJSON)
			assert.NotNil(t, cfg.S3BucketParquet)
		})
	}
}

// TestBuildCfg_Idempotency tests that multiple calls produce consistent results
func TestBuildCfg_Idempotency(t *testing.T) {
	_ = os.Setenv("S3_ACCESS_KEY", "idempotent-access")
	_ = os.Setenv("S3_SECRET_KEY", "idempotent-secret")
	_ = os.Setenv("S3_REGION", "ap-southeast-1")

	defer func() {
		_ = os.Unsetenv("S3_ACCESS_KEY")
		_ = os.Unsetenv("S3_SECRET_KEY")
		_ = os.Unsetenv("S3_REGION")
	}()

	cfg1 := BuildCfg()
	cfg2 := BuildCfg()

	// Both configs should have the same values (though different pointers)
	require.NotNil(t, cfg1)
	require.NotNil(t, cfg2)

	assert.Equal(t, *cfg1.S3KeyAccess, *cfg2.S3KeyAccess)
	assert.Equal(t, *cfg1.S3KeySecret, *cfg2.S3KeySecret)
	assert.Equal(t, *cfg1.S3Region, *cfg2.S3Region)
	assert.Equal(t, cfg1.S3SSL, cfg2.S3SSL)
}

// BenchmarkBuildCfg benchmarks the full BuildCfg function
func BenchmarkBuildCfg(b *testing.B) {
	_ = os.Setenv("S3_ACCESS_KEY", "bench-access")
	_ = os.Setenv("S3_SECRET_KEY", "bench-secret")

	defer func() {
		_ = os.Unsetenv("S3_ACCESS_KEY")
		_ = os.Unsetenv("S3_SECRET_KEY")
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = BuildCfg()
	}
}

// BenchmarkBuildCfg_WithAllEnvVars benchmarks with all env vars set
func BenchmarkBuildCfg_WithAllEnvVars(b *testing.B) {
	_ = os.Setenv("S3_ACCESS_KEY", "bench-access")
	_ = os.Setenv("S3_SECRET_KEY", "bench-secret")
	_ = os.Setenv("S3_REGION", "us-west-2")
	_ = os.Setenv("S3_ENDPOINT", "http://localhost:9000")
	_ = os.Setenv("S3_SSL", "true")
	_ = os.Setenv("S3_BUCKET_NDJSON", "bench-ndjson")
	_ = os.Setenv("S3_BUCKET_PARQUET", "bench-parquet")

	defer func() {
		_ = os.Unsetenv("S3_ACCESS_KEY")
		_ = os.Unsetenv("S3_SECRET_KEY")
		_ = os.Unsetenv("S3_REGION")
		_ = os.Unsetenv("S3_ENDPOINT")
		_ = os.Unsetenv("S3_SSL")
		_ = os.Unsetenv("S3_BUCKET_NDJSON")
		_ = os.Unsetenv("S3_BUCKET_PARQUET")
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = BuildCfg()
	}
}
