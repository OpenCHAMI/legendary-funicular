// SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

package opts

import (
	"bytes"
	"os"
	"testing"

	"github.com/OpenCHAMI/legendary-funicular/query/internal/config"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a test config
func testConfig() *config.Config {
	cfg, _ := config.New(
		config.WithDefaults(),
		config.WithS3KeyAccess("test-access"),
		config.WithS3KeySecret("test-secret"),
	)
	return cfg
}

// Helper function to create a cobra command with flags
func testCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use: "test",
		Run: func(cmd *cobra.Command, args []string) {},
	}

	rootCmd.PersistentFlags().String("scope", "all", "Data scope")
	rootCmd.PersistentFlags().String("stream", "logs", "Data stream")
	rootCmd.PersistentFlags().String("format", "json", "Output format")
	rootCmd.PersistentFlags().String("output", "", "Output file")

	return rootCmd
}

// TestFromCobraCmd_DefaultFlags tests FromCobraCmd with default flag values
func TestFromCobraCmd_DefaultFlags(t *testing.T) {
	cmd := testCommand()

	opts := FromCobraCmd(cmd)

	assert.Equal(t, "all", opts.Scope)
	assert.Equal(t, "logs", opts.Stream)
	assert.Equal(t, "json", opts.Format)
	assert.Equal(t, os.Stdout, opts.Output)
}

// TestFromCobraCmd_CustomFlags tests FromCobraCmd with custom flag values
func TestFromCobraCmd_CustomFlags(t *testing.T) {
	cmd := testCommand()
	_ = cmd.PersistentFlags().Set("scope", "compacted")
	_ = cmd.PersistentFlags().Set("stream", "events")
	_ = cmd.PersistentFlags().Set("format", "table")

	opts := FromCobraCmd(cmd)

	assert.Equal(t, "compacted", opts.Scope)
	assert.Equal(t, "events", opts.Stream)
	assert.Equal(t, "table", opts.Format)
	assert.Equal(t, os.Stdout, opts.Output)
}

// TestFromCobraCmd_OutputFile tests FromCobraCmd with output file
func TestFromCobraCmd_OutputFile(t *testing.T) {
	tmpFile := "/tmp/test-output-" + t.Name() + ".txt"
	defer func() { _ = os.Remove(tmpFile) }()

	cmd := testCommand()
	_ = cmd.PersistentFlags().Set("output", tmpFile)

	opts := FromCobraCmd(cmd)

	assert.NotEqual(t, os.Stdout, opts.Output)
	assert.NotNil(t, opts.Output)

	// Verify we can write to the output
	_, err := opts.Output.Write([]byte("test"))
	assert.NoError(t, err)
}

// TestValidate_ValidStreams tests validation with valid streams
func TestValidate_ValidStreams(t *testing.T) {
	validStreams := []string{"logs", "events"}

	for _, stream := range validStreams {
		t.Run(stream, func(t *testing.T) {
			opts := &Opts{
				Stream: stream,
				Scope:  "all",
				Format: "json",
				Output: os.Stdout,
			}

			err := opts.validate()
			assert.NoError(t, err)
		})
	}
}

// TestValidate_InvalidStreams tests validation with invalid streams
func TestValidate_InvalidStreams(t *testing.T) {
	invalidStreams := []string{"invalid", "metrics", "traces", "", "LOGS"}

	for _, stream := range invalidStreams {
		t.Run(stream, func(t *testing.T) {
			opts := &Opts{
				Stream: stream,
				Scope:  "all",
				Format: "json",
				Output: os.Stdout,
			}

			err := opts.validate()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid data stream")
			assert.Contains(t, err.Error(), stream)
		})
	}
}

// TestValidate_ValidScopes tests validation with valid scopes
func TestValidate_ValidScopes(t *testing.T) {
	validScopes := []string{"all", "compacted", "recent"}

	for _, scope := range validScopes {
		t.Run(scope, func(t *testing.T) {
			opts := &Opts{
				Stream: "logs",
				Scope:  scope,
				Format: "json",
				Output: os.Stdout,
			}

			err := opts.validate()
			assert.NoError(t, err)
		})
	}
}

// TestValidate_InvalidScopes tests validation with invalid scopes
func TestValidate_InvalidScopes(t *testing.T) {
	invalidScopes := []string{"invalid", "archived", "", "ALL"}

	for _, scope := range invalidScopes {
		t.Run(scope, func(t *testing.T) {
			opts := &Opts{
				Stream: "logs",
				Scope:  scope,
				Format: "json",
				Output: os.Stdout,
			}

			err := opts.validate()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid data scope")
			assert.Contains(t, err.Error(), scope)
		})
	}
}

// TestBuildSources_ScopeAll tests BuildSources with scope=all
func TestBuildSources_ScopeAll(t *testing.T) {
	cfg := testConfig()
	opts := &Opts{
		Stream: "logs",
		Scope:  "all",
		Format: "json",
		Output: os.Stdout,
	}

	sources, err := opts.BuildSources(cfg)
	require.NoError(t, err)

	require.Len(t, sources, 2, "scope=all should return both parquet and ndjson sources")

	// Check parquet source
	assert.Contains(t, sources[0], "read_parquet")
	assert.Contains(t, sources[0], *cfg.S3BucketParquet)
	assert.Contains(t, sources[0], "logs")
	assert.Contains(t, sources[0], "union_by_name")

	// Check ndjson source
	assert.Contains(t, sources[1], "read_json")
	assert.Contains(t, sources[1], *cfg.S3BucketNDJSON)
	assert.Contains(t, sources[1], "logs")
	assert.Contains(t, sources[1], ".ndjson.zst")
}

// TestBuildSources_ScopeCompacted tests BuildSources with scope=compacted
func TestBuildSources_ScopeCompacted(t *testing.T) {
	cfg := testConfig()
	opts := &Opts{
		Stream: "logs",
		Scope:  "compacted",
		Format: "json",
		Output: os.Stdout,
	}

	sources, err := opts.BuildSources(cfg)
	require.NoError(t, err)

	require.Len(t, sources, 1, "scope=compacted should return only parquet source")
	assert.Contains(t, sources[0], "read_parquet")
	assert.Contains(t, sources[0], *cfg.S3BucketParquet)
}

// TestBuildSources_ScopeRecent tests BuildSources with scope=recent
func TestBuildSources_ScopeRecent(t *testing.T) {
	cfg := testConfig()
	opts := &Opts{
		Stream: "logs",
		Scope:  "recent",
		Format: "json",
		Output: os.Stdout,
	}

	sources, err := opts.BuildSources(cfg)
	require.NoError(t, err)

	require.Len(t, sources, 1, "scope=recent should return only ndjson source")
	assert.Contains(t, sources[0], "read_json")
	assert.Contains(t, sources[0], *cfg.S3BucketNDJSON)
	assert.Contains(t, sources[0], ".ndjson.zst")
}

// TestBuildSources_StreamLogs tests BuildSources with stream=logs
func TestBuildSources_StreamLogs(t *testing.T) {
	cfg := testConfig()
	opts := &Opts{
		Stream: "logs",
		Scope:  "all",
		Format: "json",
		Output: os.Stdout,
	}

	sources, err := opts.BuildSources(cfg)
	require.NoError(t, err)

	for _, source := range sources {
		assert.Contains(t, source, "logs", "All sources should reference logs stream")
	}
}

// TestBuildSources_StreamEvents tests BuildSources with stream=events
func TestBuildSources_StreamEvents(t *testing.T) {
	cfg := testConfig()
	opts := &Opts{
		Stream: "events",
		Scope:  "all",
		Format: "json",
		Output: os.Stdout,
	}

	sources, err := opts.BuildSources(cfg)
	require.NoError(t, err)

	for _, source := range sources {
		assert.Contains(t, source, "events", "All sources should reference events stream")
	}
}

// TestBuildSources_AllCombinations tests all valid scope/stream combinations
func TestBuildSources_AllCombinations(t *testing.T) {
	cfg := testConfig()

	tests := []struct {
		scope         string
		stream        string
		expectedCount int
	}{
		{"all", "logs", 2},
		{"all", "events", 2},
		{"compacted", "logs", 1},
		{"compacted", "events", 1},
		{"recent", "logs", 1},
		{"recent", "events", 1},
	}

	for _, tt := range tests {
		t.Run(tt.scope+"_"+tt.stream, func(t *testing.T) {
			opts := &Opts{
				Stream: tt.stream,
				Scope:  tt.scope,
				Format: "json",
				Output: os.Stdout,
			}

			sources, err := opts.BuildSources(cfg)
			require.NoError(t, err)
			assert.Len(t, sources, tt.expectedCount)

			for _, source := range sources {
				assert.Contains(t, source, tt.stream)
			}
		})
	}
}

// TestBuildSources_ParquetPath tests parquet path construction
func TestBuildSources_ParquetPath(t *testing.T) {
	cfg := testConfig()
	opts := &Opts{
		Stream: "logs",
		Scope:  "compacted",
		Format: "json",
		Output: os.Stdout,
	}

	sources, err := opts.BuildSources(cfg)
	require.NoError(t, err)

	expectedPath := "s3://openchami-logs-daily/logs/**/*.parquet"
	assert.Contains(t, sources[0], expectedPath)
}

// TestBuildSources_NDJSONPath tests ndjson path construction
func TestBuildSources_NDJSONPath(t *testing.T) {
	cfg := testConfig()
	opts := &Opts{
		Stream: "logs",
		Scope:  "recent",
		Format: "json",
		Output: os.Stdout,
	}

	sources, err := opts.BuildSources(cfg)
	require.NoError(t, err)

	expectedPath := "s3://openchami-logs-raw/logs/**/*.ndjson.zst"
	assert.Contains(t, sources[0], expectedPath)
}

// TestOpts_OutputWriter tests that Output field can be used for writing
func TestOpts_OutputWriter(t *testing.T) {
	var buf bytes.Buffer
	opts := &Opts{
		Stream: "logs",
		Scope:  "all",
		Format: "json",
		Output: &buf,
	}

	testData := []byte("test output data")
	n, err := opts.Output.Write(testData)

	assert.NoError(t, err)
	assert.Equal(t, len(testData), n)
	assert.Equal(t, testData, buf.Bytes())
}

// TestOpts_MultipleWrites tests multiple writes to Output
func TestOpts_MultipleWrites(t *testing.T) {
	var buf bytes.Buffer
	opts := &Opts{
		Stream: "logs",
		Scope:  "all",
		Format: "json",
		Output: &buf,
	}

	writes := [][]byte{
		[]byte("line 1\n"),
		[]byte("line 2\n"),
		[]byte("line 3\n"),
	}

	for _, data := range writes {
		_, err := opts.Output.Write(data)
		assert.NoError(t, err)
	}

	expected := "line 1\nline 2\nline 3\n"
	assert.Equal(t, expected, buf.String())
}

// BenchmarkValidate benchmarks the validate function
func BenchmarkValidate(b *testing.B) {
	opts := &Opts{
		Stream: "logs",
		Scope:  "all",
		Format: "json",
		Output: os.Stdout,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = opts.validate() // Ignore error in benchmark
	}
}

// BenchmarkBuildSources_All benchmarks BuildSources with scope=all
func BenchmarkBuildSources_All(b *testing.B) {
	cfg := testConfig()
	opts := &Opts{
		Stream: "logs",
		Scope:  "all",
		Format: "json",
		Output: os.Stdout,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = opts.BuildSources(cfg) // Ignore errors in benchmark
	}
}

// BenchmarkBuildSources_Compacted benchmarks BuildSources with scope=compacted
func BenchmarkBuildSources_Compacted(b *testing.B) {
	cfg := testConfig()
	opts := &Opts{
		Stream: "logs",
		Scope:  "compacted",
		Format: "json",
		Output: os.Stdout,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = opts.BuildSources(cfg) // Ignore errors in benchmark
	}
}

// BenchmarkBuildSources_Recent benchmarks BuildSources with scope=recent
func BenchmarkBuildSources_Recent(b *testing.B) {
	cfg := testConfig()
	opts := &Opts{
		Stream: "logs",
		Scope:  "recent",
		Format: "json",
		Output: os.Stdout,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = opts.BuildSources(cfg) // Ignore errors in benchmark
	}
}
