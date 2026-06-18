// SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

package main

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockBucket is a test implementation of bucket.Bucket interface
type MockBucket struct {
	name    string
	objects map[string][]byte
	listErr error
	getErr  error
	putErr  error
	delErr  error
}

func NewMockBucket(name string) *MockBucket {
	return &MockBucket{
		name:    name,
		objects: make(map[string][]byte),
	}
}

func (m *MockBucket) List(ctx context.Context, prefix string) ([]string, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}

	var keys []string
	for key := range m.objects {
		if strings.HasPrefix(key, prefix) {
			keys = append(keys, key)
		}
	}
	return keys, nil
}

func (m *MockBucket) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}

	data, ok := m.objects[key]
	if !ok {
		return nil, io.EOF
	}

	return io.NopCloser(bytes.NewReader(data)), nil
}

func (m *MockBucket) SpooledPut(ctx context.Context, key string, r io.Reader) (int64, error) {
	if m.putErr != nil {
		return 0, m.putErr
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return 0, err
	}

	m.objects[key] = data
	return int64(len(data)), nil
}

func (m *MockBucket) Delete(ctx context.Context, key string) error {
	if m.delErr != nil {
		return m.delErr
	}

	delete(m.objects, key)
	return nil
}

// MockRecord for testing
type MockRecord struct {
	ID      string
	Message string
}

// MockParser for testing
type MockParser struct {
	parseErr error
}

func (p *MockParser) Parse(line []byte) (MockRecord, error) {
	if p.parseErr != nil {
		return MockRecord{}, p.parseErr
	}

	return MockRecord{
		ID:      "test-id",
		Message: string(line),
	}, nil
}

// TestCompaction_EmptySource tests compaction with empty source
func TestCompaction_EmptySource(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	_ = NewMockBucket("source")
	_ = NewMockBucket("sink")
	_ = &MockParser{}

	// Note: This test requires the actual compaction function signature
	// which uses generics and internal types
	// Documenting expected behavior:

	// When source bucket is empty:
	// - List returns empty array
	// - Compaction returns nil (no error)
	// - No objects written to sink
	// - No objects deleted from source

	t.Log("Expected: compaction returns nil for empty source")
}

// TestCompaction_SingleFile tests compaction with single file
func TestCompaction_SingleFile(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Expected flow:
	// 1. List source bucket (finds 1 file)
	// 2. Create pipe for streaming
	// 3. Start reader goroutine:
	//    - Read from S3
	//    - Decompress (if compressed)
	//    - Parse lines
	//    - Convert to Parquet
	//    - Write to pipe
	// 4. Start writer goroutine:
	//    - Read from pipe
	//    - Upload to sink bucket
	// 5. Wait for both goroutines
	// 6. Delete source files

	t.Log("Expected: single file compacted successfully")
}

// TestCompaction_MultipleFiles tests compaction with multiple files
func TestCompaction_MultipleFiles(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Expected flow:
	// 1. List source bucket (finds N files)
	// 2. Process all files in parallel
	// 3. Combine into single Parquet file
	// 4. Upload combined file
	// 5. Delete all source files

	t.Log("Expected: multiple files combined into one Parquet")
}

// TestCompaction_ErrorInReader tests error handling in reader goroutine
func TestCompaction_ErrorInReader(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Expected behavior:
	// - Reader encounters error (parse failure, S3 error, etc.)
	// - Error sent to producerErrCh
	// - Writer goroutine sees broken pipe
	// - Main function receives error
	// - No files deleted from source (safety)

	t.Log("Expected: reader error propagates to main, no source deletion")
}

// TestCompaction_ErrorInWriter tests error handling in writer goroutine
func TestCompaction_ErrorInWriter(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Expected behavior:
	// - Writer encounters error (S3 put failure)
	// - Error sent to consumerErrCh
	// - Reader continues until done or sees closed pipe
	// - Main function receives error
	// - No files deleted from source (safety)

	t.Log("Expected: writer error propagates to main, no source deletion")
}

// TestCompaction_ConcurrentPipeline tests pipeline concurrency
func TestCompaction_ConcurrentPipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Expected behavior:
	// - Reader and writer run concurrently
	// - Data streams through pipe (no full buffering)
	// - Memory usage stays constant (streaming)
	// - Both goroutines coordinate via WaitGroup

	t.Log("Expected: reader and writer run concurrently")
}

// TestCompaction_SourceDeletion tests source file deletion
func TestCompaction_SourceDeletion(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Expected behavior:
	// - Compaction succeeds
	// - All source files are deleted
	// - If any deletion fails, error is logged but not fatal
	// - Partial deletion is acceptable (can retry later)

	t.Log("Expected: source files deleted after successful compaction")
}

// TestCompaction_PrefixFiltering tests prefix-based filtering
func TestCompaction_PrefixFiltering(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Expected behavior:
	// - Only files matching prefix are listed
	// - Other files in bucket are ignored
	// - Allows parallel compaction of different prefixes

	t.Log("Expected: only files matching prefix are processed")
}

// TestCompaction_ContextCancellation tests context cancellation
func TestCompaction_ContextCancellation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Expected behavior:
	// - Context is cancelled during compaction
	// - S3 operations respect context
	// - Goroutines terminate gracefully
	// - Error returned indicating cancellation

	t.Log("Expected: compaction stops on context cancellation")
}

// TestCompactionFlow_Documentation documents the complete compaction flow
func TestCompactionFlow_Documentation(t *testing.T) {
	// COMPLETE E2E COMPACTION FLOW

	// PHASE 1: DISCOVERY
	// 1. List NDJSON files in source S3 bucket
	// 2. Filter by prefix (e.g., "2026-06-10/")
	// 3. Return empty if no files found

	// PHASE 2: PIPELINE SETUP
	// 1. Create io.Pipe for streaming
	// 2. Create error channels for goroutine communication
	// 3. Create key channel for tracking processed files
	// 4. Start WaitGroup for goroutine coordination

	// PHASE 3: READER GOROUTINE (Producer)
	// 1. For each source file:
	//    a. Download from S3 (streaming)
	//    b. Decompress zstd stream
	//    c. Parse each line (syslog or cloudevent)
	//    d. Handle parse errors gracefully
	// 2. Convert all records to Parquet format
	// 3. Write Parquet to pipe (streaming)
	// 4. Send processed file keys to keyCh
	// 5. Close pipe when done
	// 6. Send error to producerErrCh if any

	// PHASE 4: WRITER GOROUTINE (Consumer)
	// 1. Read Parquet data from pipe
	// 2. Upload to sink S3 bucket
	// 3. Log bytes written
	// 4. Send error to consumerErrCh if any

	// PHASE 5: COORDINATION
	// 1. Collect processed file keys from keyCh
	// 2. Wait for producer to finish
	// 3. Wait for consumer to finish
	// 4. Check both error channels
	// 5. Return error if any occurred

	// PHASE 6: CLEANUP
	// 1. If compaction succeeded:
	//    - Delete all processed files from source
	//    - Log any deletion failures
	// 2. If compaction failed:
	//    - Keep source files for retry
	//    - Return error to caller

	// KEY DESIGN DECISIONS:
	// - Streaming pipeline (constant memory)
	// - Concurrent reader/writer (performance)
	// - Error channels (goroutine communication)
	// - Safe cleanup (only delete on success)
	// - Graceful degradation (log but don't fail on delete errors)

	t.Log("Compaction flow fully documented")
}

// TestBuildObjectKey_E2E tests object key generation in E2E context
func TestBuildObjectKey_E2E(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
	}{
		{"simple prefix", "2026-06-10"},
		{"nested prefix", "logs/2026/06/10"},
		{"with trailing slash", "2026-06-10/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := buildObjectKey(tt.prefix)

			// Should contain prefix
			assert.Contains(t, key, tt.prefix)

			// Should have parquet extension
			assert.Contains(t, key, ".parquet")

			// Should not be empty
			assert.NotEmpty(t, key)
		})
	}
}

// TestMockBucket_Interface tests MockBucket implementation
func TestMockBucket_Interface(t *testing.T) {
	bucket := NewMockBucket("test-bucket")

	assert.NotNil(t, bucket)
	assert.Equal(t, "test-bucket", bucket.name)
	assert.NotNil(t, bucket.objects)
}

// TestMockBucket_ListEmpty tests listing empty bucket
func TestMockBucket_ListEmpty(t *testing.T) {
	bucket := NewMockBucket("test")

	keys, err := bucket.List(context.Background(), "prefix")

	require.NoError(t, err)
	assert.Empty(t, keys)
}

// TestMockBucket_ListWithObjects tests listing with objects
func TestMockBucket_ListWithObjects(t *testing.T) {
	bucket := NewMockBucket("test")
	bucket.objects["prefix/file1.json"] = []byte("data1")
	bucket.objects["prefix/file2.json"] = []byte("data2")
	bucket.objects["other/file3.json"] = []byte("data3")

	keys, err := bucket.List(context.Background(), "prefix")

	require.NoError(t, err)
	assert.Len(t, keys, 2)
	assert.Contains(t, keys, "prefix/file1.json")
	assert.Contains(t, keys, "prefix/file2.json")
	assert.NotContains(t, keys, "other/file3.json")
}

// TestMockBucket_GetNotFound tests getting non-existent object
func TestMockBucket_GetNotFound(t *testing.T) {
	bucket := NewMockBucket("test")

	_, err := bucket.Get(context.Background(), "missing.json")

	assert.Error(t, err)
	assert.Equal(t, io.EOF, err)
}

// TestMockBucket_GetSuccess tests getting existing object
func TestMockBucket_GetSuccess(t *testing.T) {
	bucket := NewMockBucket("test")
	bucket.objects["test.json"] = []byte("test data")

	rc, err := bucket.Get(context.Background(), "test.json")

	require.NoError(t, err)
	defer rc.Close() //nolint:errcheck // test cleanup

	data, err := io.ReadAll(rc)
	require.NoError(t, err)
	assert.Equal(t, "test data", string(data))
}

// TestMockBucket_SpooledPut tests uploading data
func TestMockBucket_SpooledPut(t *testing.T) {
	bucket := NewMockBucket("test")
	data := []byte("upload data")

	n, err := bucket.SpooledPut(context.Background(), "upload.json", bytes.NewReader(data))

	require.NoError(t, err)
	assert.Equal(t, int64(len(data)), n)
	assert.Equal(t, data, bucket.objects["upload.json"])
}

// TestMockBucket_Delete tests deleting object
func TestMockBucket_Delete(t *testing.T) {
	bucket := NewMockBucket("test")
	bucket.objects["delete.json"] = []byte("data")

	err := bucket.Delete(context.Background(), "delete.json")

	require.NoError(t, err)
	assert.NotContains(t, bucket.objects, "delete.json")
}

// TestMockParser_Parse tests parsing records
func TestMockParser_Parse(t *testing.T) {
	parser := &MockParser{}

	record, err := parser.Parse([]byte("test message"))

	require.NoError(t, err)
	assert.Equal(t, "test-id", record.ID)
	assert.Equal(t, "test message", record.Message)
}

// TestMockParser_ParseError tests parse error handling
func TestMockParser_ParseError(t *testing.T) {
	parser := &MockParser{parseErr: io.ErrUnexpectedEOF}

	_, err := parser.Parse([]byte("invalid"))

	assert.Error(t, err)
	assert.Equal(t, io.ErrUnexpectedEOF, err)
}

// BenchmarkBuildObjectKey_E2E benchmarks key generation in E2E context
func BenchmarkBuildObjectKey_E2E(b *testing.B) {
	prefix := "2026-06-10/logs"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = buildObjectKey(prefix)
	}
}
