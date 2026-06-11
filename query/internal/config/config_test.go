// SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNew_WithDefaults tests creating config with defaults
func TestNew_WithDefaults(t *testing.T) {
	cfg, err := New(
		WithDefaults(),
		WithS3KeyAccess("test-access-key"),
		WithS3KeySecret("test-secret-key"),
	)

	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Verify defaults
	assert.False(t, cfg.S3SSL)
	assert.NotNil(t, cfg.S3Region)
	assert.Equal(t, "us-east-1", *cfg.S3Region)
	assert.NotNil(t, cfg.S3Endpoint)
	assert.Equal(t, "localhost:7070", *cfg.S3Endpoint)
	assert.NotNil(t, cfg.S3BucketNDJSON)
	assert.Equal(t, "openchami-logs-raw", *cfg.S3BucketNDJSON)
	assert.NotNil(t, cfg.S3BucketParquet)
	assert.Equal(t, "openchami-logs-daily", *cfg.S3BucketParquet)

	// Verify required fields set
	assert.NotNil(t, cfg.S3KeyAccess)
	assert.Equal(t, "test-access-key", *cfg.S3KeyAccess)
	assert.NotNil(t, cfg.S3KeySecret)
	assert.Equal(t, "test-secret-key", *cfg.S3KeySecret)
}

// TestNew_MissingRequired tests that missing required fields cause errors
func TestNew_MissingRequired(t *testing.T) {
	tests := []struct {
		name    string
		opts    []Option
		wantErr string
	}{
		{
			name:    "missing all required",
			opts:    []Option{},
			wantErr: "s3_region is required",
		},
		{
			name: "missing access key",
			opts: []Option{
				WithDefaults(),
				WithS3KeySecret("secret"),
			},
			wantErr: "s3_key_access is required",
		},
		{
			name: "missing secret key",
			opts: []Option{
				WithDefaults(),
				WithS3KeyAccess("access"),
			},
			wantErr: "s3_key_secret is required",
		},
		{
			name: "missing region",
			opts: []Option{
				WithS3KeyAccess("access"),
				WithS3KeySecret("secret"),
				WithS3Endpoint("localhost:7070"),
				WithS3BucketNDJSON("bucket1"),
				WithS3BucketParquet("bucket2"),
			},
			wantErr: "s3_region is required",
		},
		{
			name: "missing endpoint",
			opts: []Option{
				WithS3KeyAccess("access"),
				WithS3KeySecret("secret"),
				WithS3Region("us-east-1"),
				WithS3BucketNDJSON("bucket1"),
				WithS3BucketParquet("bucket2"),
			},
			wantErr: "s3_endpoint is required",
		},
		{
			name: "missing ndjson bucket",
			opts: []Option{
				WithS3KeyAccess("access"),
				WithS3KeySecret("secret"),
				WithS3Region("us-east-1"),
				WithS3Endpoint("localhost:7070"),
				WithS3BucketParquet("bucket2"),
			},
			wantErr: "s3_bucket_ndjson is required",
		},
		{
			name: "missing parquet bucket",
			opts: []Option{
				WithS3KeyAccess("access"),
				WithS3KeySecret("secret"),
				WithS3Region("us-east-1"),
				WithS3Endpoint("localhost:7070"),
				WithS3BucketNDJSON("bucket1"),
			},
			wantErr: "s3_bucket_parquet is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := New(tt.opts...)
			assert.Error(t, err)
			assert.Nil(t, cfg)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

// TestNew_EmptyStringsTreatedAsMissing tests empty strings are invalid
func TestNew_EmptyStringsTreatedAsMissing(t *testing.T) {
	tests := []struct {
		name    string
		opts    []Option
		wantErr string
	}{
		{
			name: "empty access key",
			opts: []Option{
				WithDefaults(),
				WithS3KeyAccess(""),
				WithS3KeySecret("secret"),
			},
			wantErr: "s3_key_access is required",
		},
		{
			name: "empty secret key",
			opts: []Option{
				WithDefaults(),
				WithS3KeyAccess("access"),
				WithS3KeySecret(""),
			},
			wantErr: "s3_key_secret is required",
		},
		{
			name: "empty region",
			opts: []Option{
				WithS3KeyAccess("access"),
				WithS3KeySecret("secret"),
				WithS3Region(""),
				WithS3Endpoint("localhost:7070"),
				WithS3BucketNDJSON("bucket1"),
				WithS3BucketParquet("bucket2"),
			},
			wantErr: "s3_region is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := New(tt.opts...)
			assert.Error(t, err)
			assert.Nil(t, cfg)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

// TestNew_OptionOrdering tests that option order doesn't matter
func TestNew_OptionOrdering(t *testing.T) {
	// Create config with options in different orders
	cfg1, err1 := New(
		WithDefaults(),
		WithS3KeyAccess("access"),
		WithS3KeySecret("secret"),
	)

	cfg2, err2 := New(
		WithS3KeySecret("secret"),
		WithS3KeyAccess("access"),
		WithDefaults(),
	)

	cfg3, err3 := New(
		WithS3KeyAccess("access"),
		WithDefaults(),
		WithS3KeySecret("secret"),
	)

	require.NoError(t, err1)
	require.NoError(t, err2)
	require.NoError(t, err3)

	// All should be valid configs (values may differ due to overrides)
	assert.NotNil(t, cfg1)
	assert.NotNil(t, cfg2)
	assert.NotNil(t, cfg3)
}

// TestNew_OptionOverrides tests that later options override earlier ones
func TestNew_OptionOverrides(t *testing.T) {
	cfg, err := New(
		WithDefaults(),
		WithS3KeyAccess("access"),
		WithS3KeySecret("secret"),
		WithS3Region("us-west-2"), // Override default
		WithS3SSL(true),           // Override default
	)

	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Verify overrides took effect
	assert.Equal(t, "us-west-2", *cfg.S3Region)
	assert.True(t, cfg.S3SSL)
}

// TestNew_FullCustomConfig tests creating fully custom config
func TestNew_FullCustomConfig(t *testing.T) {
	cfg, err := New(
		WithS3KeyAccess("AKIAIOSFODNN7EXAMPLE"),
		WithS3KeySecret("wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"),
		WithS3Region("us-west-2"),
		WithS3Endpoint("https://s3.us-west-2.amazonaws.com"),
		WithS3BucketNDJSON("my-ndjson-bucket"),
		WithS3BucketParquet("my-parquet-bucket"),
		WithS3SSL(true),
	)

	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "AKIAIOSFODNN7EXAMPLE", *cfg.S3KeyAccess)
	assert.Equal(t, "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY", *cfg.S3KeySecret)
	assert.Equal(t, "us-west-2", *cfg.S3Region)
	assert.Equal(t, "https://s3.us-west-2.amazonaws.com", *cfg.S3Endpoint)
	assert.Equal(t, "my-ndjson-bucket", *cfg.S3BucketNDJSON)
	assert.Equal(t, "my-parquet-bucket", *cfg.S3BucketParquet)
	assert.True(t, cfg.S3SSL)
}

// TestNew_MultipleRegions tests various AWS regions
func TestNew_MultipleRegions(t *testing.T) {
	regions := []string{
		"us-east-1",
		"us-west-2",
		"eu-west-1",
		"ap-southeast-1",
		"ca-central-1",
	}

	for _, region := range regions {
		t.Run(region, func(t *testing.T) {
			cfg, err := New(
				WithS3KeyAccess("access"),
				WithS3KeySecret("secret"),
				WithS3Region(region),
				WithS3Endpoint("localhost:7070"),
				WithS3BucketNDJSON("bucket1"),
				WithS3BucketParquet("bucket2"),
			)

			require.NoError(t, err)
			assert.Equal(t, region, *cfg.S3Region)
		})
	}
}

// TestNew_SSLToggle tests SSL enable/disable
func TestNew_SSLToggle(t *testing.T) {
	tests := []struct {
		name      string
		sslValue  bool
		wantValue bool
	}{
		{"SSL enabled", true, true},
		{"SSL disabled", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := New(
				WithDefaults(),
				WithS3KeyAccess("access"),
				WithS3KeySecret("secret"),
				WithS3SSL(tt.sslValue),
			)

			require.NoError(t, err)
			assert.Equal(t, tt.wantValue, cfg.S3SSL)
		})
	}
}

// TestValidate_NilConfig tests validation of nil config
func TestValidate_NilConfig(t *testing.T) {
	err := validate(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config is nil")
}

// TestRef tests the ref helper function
func TestRef(t *testing.T) {
	// String
	s := "test"
	sPtr := ref(s)
	assert.NotNil(t, sPtr)
	assert.Equal(t, "test", *sPtr)

	// Int
	i := 42
	iPtr := ref(i)
	assert.NotNil(t, iPtr)
	assert.Equal(t, 42, *iPtr)

	// Bool
	b := true
	bPtr := ref(b)
	assert.NotNil(t, bPtr)
	assert.True(t, *bPtr)
}

// TestReq tests the req validation helper
func TestReq(t *testing.T) {
	tests := []struct {
		name    string
		ptr     *string
		field   string
		wantErr bool
	}{
		{"valid string", ref("value"), "field", false},
		{"nil pointer", nil, "field", true},
		{"empty string", ref(""), "field", true},
		{"whitespace only", ref("   "), "field", false}, // Whitespace is valid
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := req(tt.ptr, tt.field)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.field)
				assert.Contains(t, err.Error(), "is required")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// BenchmarkNew_WithDefaults benchmarks config creation with defaults
func BenchmarkNew_WithDefaults(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = New(
			WithDefaults(),
			WithS3KeyAccess("access"),
			WithS3KeySecret("secret"),
		)
	}
}

// BenchmarkNew_FullCustom benchmarks config creation with all custom options
func BenchmarkNew_FullCustom(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = New(
			WithS3KeyAccess("access"),
			WithS3KeySecret("secret"),
			WithS3Region("us-west-2"),
			WithS3Endpoint("localhost:7070"),
			WithS3BucketNDJSON("bucket1"),
			WithS3BucketParquet("bucket2"),
			WithS3SSL(true),
		)
	}
}

// BenchmarkValidate benchmarks config validation
func BenchmarkValidate(b *testing.B) {
	cfg := &Config{
		S3SSL:           true,
		S3Region:        ref("us-east-1"),
		S3Endpoint:      ref("localhost:7070"),
		S3BucketNDJSON:  ref("bucket1"),
		S3BucketParquet: ref("bucket2"),
		S3KeyAccess:     ref("access"),
		S3KeySecret:     ref("secret"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validate(cfg)
	}
}
