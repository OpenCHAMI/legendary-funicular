// SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

package zio

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/klauspost/compress/zstd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper to create zstd compressed data
func compressZstd(tb testing.TB, data string) []byte {
	tb.Helper()

	var buf bytes.Buffer
	w, err := zstd.NewWriter(&buf)
	require.NoError(tb, err)

	_, err = w.Write([]byte(data))
	require.NoError(tb, err)

	err = w.Close()
	require.NoError(tb, err)

	return buf.Bytes()
}

// TestNewUnzstdStream_ValidData tests decompression of valid zstd data
func TestNewUnzstdStream_ValidData(t *testing.T) {
	original := "hello world, this is a test"
	compressed := compressZstd(t, original)

	reader := bytes.NewReader(compressed)
	uz, err := NewUnzstdStream(reader)

	require.NoError(t, err)
	assert.NotNil(t, uz)

	// Read decompressed data
	data, err := io.ReadAll(uz)
	require.NoError(t, err)
	assert.Equal(t, original, string(data))

	err = uz.Close()
	assert.NoError(t, err)
}

// TestNewUnzstdStream_EmptyData tests decompression of empty data
func TestNewUnzstdStream_EmptyData(t *testing.T) {
	compressed := compressZstd(t, "")

	reader := bytes.NewReader(compressed)
	uz, err := NewUnzstdStream(reader)

	require.NoError(t, err)

	data, err := io.ReadAll(uz)
	require.NoError(t, err)
	assert.Equal(t, "", string(data))

	_ = uz.Close() // Ignore error in test
}

// TestNewUnzstdStream_LargeData tests decompression of large data
func TestNewUnzstdStream_LargeData(t *testing.T) {
	// Create large data (1MB)
	original := strings.Repeat("The quick brown fox jumps over the lazy dog. ", 20000)
	compressed := compressZstd(t, original)

	// Verify compression worked
	compressionRatio := float64(len(original)) / float64(len(compressed))
	t.Logf("Compression ratio: %.2fx (%d bytes -> %d bytes)",
		compressionRatio, len(original), len(compressed))
	assert.Greater(t, compressionRatio, 2.0, "Should achieve at least 2x compression")

	reader := bytes.NewReader(compressed)
	uz, err := NewUnzstdStream(reader)

	require.NoError(t, err)

	data, err := io.ReadAll(uz)
	require.NoError(t, err)
	assert.Equal(t, original, string(data))

	_ = uz.Close() // Ignore error in test
}

// TestNewUnzstdStream_InvalidData tests error handling with invalid data
func TestNewUnzstdStream_InvalidData(t *testing.T) {
	// Invalid zstd data
	reader := strings.NewReader("this is not compressed data")
	uz, err := NewUnzstdStream(reader)

	require.NoError(t, err) // NewUnzstdStream doesn't validate header

	// Error should occur during Read
	_, err = io.ReadAll(uz)
	assert.Error(t, err)

	_ = uz.Close() // Ignore error in test cleanup
}

// TestNewUnzstdStream_PartialRead tests reading data in chunks
func TestNewUnzstdStream_PartialRead(t *testing.T) {
	original := "hello world, this is a longer test message for partial reading"
	compressed := compressZstd(t, original)

	reader := bytes.NewReader(compressed)
	uz, err := NewUnzstdStream(reader)

	require.NoError(t, err)

	// Read in small chunks
	buf := make([]byte, 10)
	var result bytes.Buffer

	for {
		n, err := uz.Read(buf)
		if n > 0 {
			result.Write(buf[:n])
		}
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
	}

	assert.Equal(t, original, result.String())

	_ = uz.Close() // Ignore error in test
}

// TestUnzstdStream_Close tests Close method
func TestUnzstdStream_Close(t *testing.T) {
	compressed := compressZstd(t, "test data")
	reader := bytes.NewReader(compressed)

	uz, err := NewUnzstdStream(reader)
	require.NoError(t, err)

	// Close should not error
	err = uz.Close()
	assert.NoError(t, err)
}

// TestUnzstdStream_MultipleClose tests calling Close multiple times
func TestUnzstdStream_MultipleClose(t *testing.T) {
	compressed := compressZstd(t, "test data")
	reader := bytes.NewReader(compressed)

	uz, err := NewUnzstdStream(reader)
	require.NoError(t, err)

	// First close
	err = uz.Close()
	assert.NoError(t, err)

	// Second close should not panic
	err = uz.Close()
	assert.NoError(t, err)
}

// TestUnzstdStream_ReadAfterClose tests reading after close
func TestUnzstdStream_ReadAfterClose(t *testing.T) {
	compressed := compressZstd(t, "test data")
	reader := bytes.NewReader(compressed)

	uz, err := NewUnzstdStream(reader)
	require.NoError(t, err)

	_ = uz.Close() // Ignore error in test cleanup

	// Read after close should error
	buf := make([]byte, 100)
	_, err = uz.Read(buf)
	assert.Error(t, err)
}

// TestNewUnzstdStream_MultipleStreams tests multiple concurrent streams
func TestNewUnzstdStream_MultipleStreams(t *testing.T) {
	data1 := "first stream data"
	data2 := "second stream data"
	data3 := "third stream data"

	compressed1 := compressZstd(t, data1)
	compressed2 := compressZstd(t, data2)
	compressed3 := compressZstd(t, data3)

	uz1, err := NewUnzstdStream(bytes.NewReader(compressed1))
	require.NoError(t, err)

	uz2, err := NewUnzstdStream(bytes.NewReader(compressed2))
	require.NoError(t, err)

	uz3, err := NewUnzstdStream(bytes.NewReader(compressed3))
	require.NoError(t, err)

	// Read from all streams
	result1, err := io.ReadAll(uz1)
	require.NoError(t, err)

	result2, err := io.ReadAll(uz2)
	require.NoError(t, err)

	result3, err := io.ReadAll(uz3)
	require.NoError(t, err)

	assert.Equal(t, data1, string(result1))
	assert.Equal(t, data2, string(result2))
	assert.Equal(t, data3, string(result3))

	_ = uz1.Close() // Ignore errors in test
	_ = uz2.Close()
	_ = uz3.Close()
}

// TestNewUnzstdStream_JSONData tests decompression of JSON data
func TestNewUnzstdStream_JSONData(t *testing.T) {
	jsonData := `{"id":1,"message":"test","timestamp":"2026-06-10T12:00:00Z","nested":{"key":"value"}}`
	compressed := compressZstd(t, jsonData)

	reader := bytes.NewReader(compressed)
	uz, err := NewUnzstdStream(reader)

	require.NoError(t, err)

	data, err := io.ReadAll(uz)
	require.NoError(t, err)
	assert.Equal(t, jsonData, string(data))

	_ = uz.Close() // Ignore error in test cleanup
}

// TestNewUnzstdStream_BinaryData tests decompression of binary data
func TestNewUnzstdStream_BinaryData(t *testing.T) {
	// Binary data with various byte values
	binaryData := []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD, 0x7F, 0x80}

	var buf bytes.Buffer
	w, err := zstd.NewWriter(&buf)
	require.NoError(t, err)

	_, err = w.Write(binaryData)
	require.NoError(t, err)

	err = w.Close()
	require.NoError(t, err)

	compressed := buf.Bytes()

	reader := bytes.NewReader(compressed)
	uz, err := NewUnzstdStream(reader)

	require.NoError(t, err)

	data, err := io.ReadAll(uz)
	require.NoError(t, err)
	assert.Equal(t, binaryData, data)

	_ = uz.Close() // Ignore error in test cleanup
}

// BenchmarkNewUnzstdStream_Small benchmarks small data decompression
func BenchmarkNewUnzstdStream_Small(b *testing.B) {
	data := "hello world"
	compressed := compressZstd(b, data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := bytes.NewReader(compressed)
		uz, _ := NewUnzstdStream(reader)
		_, _ = io.Copy(io.Discard, uz) // Ignore errors in benchmark
		_ = uz.Close()                 // Ignore errors in benchmark
	}
}

// BenchmarkNewUnzstdStream_Medium benchmarks medium data decompression
func BenchmarkNewUnzstdStream_Medium(b *testing.B) {
	data := strings.Repeat("The quick brown fox jumps over the lazy dog. ", 100)
	compressed := compressZstd(b, data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := bytes.NewReader(compressed)
		uz, _ := NewUnzstdStream(reader)
		_, _ = io.Copy(io.Discard, uz) // Ignore errors in benchmark
		_ = uz.Close()                 // Ignore errors in benchmark
	}
}

// BenchmarkNewUnzstdStream_Large benchmarks large data decompression
func BenchmarkNewUnzstdStream_Large(b *testing.B) {
	data := strings.Repeat("The quick brown fox jumps over the lazy dog. ", 10000)
	compressed := compressZstd(b, data)

	b.SetBytes(int64(len(data)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := bytes.NewReader(compressed)
		uz, _ := NewUnzstdStream(reader)
		_, _ = io.Copy(io.Discard, uz) // Ignore errors in benchmark
		_ = uz.Close()                 // Ignore errors in benchmark
	}
}

// BenchmarkNewUnzstdStream_JSON benchmarks JSON data decompression
func BenchmarkNewUnzstdStream_JSON(b *testing.B) {
	jsonData := `{"id":1,"message":"test message","timestamp":"2026-06-10T12:00:00Z","level":"INFO","host":"server01","service":"api","nested":{"key1":"value1","key2":"value2","key3":"value3"}}`
	compressed := compressZstd(b, jsonData)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := bytes.NewReader(compressed)
		uz, _ := NewUnzstdStream(reader)
		_, _ = io.Copy(io.Discard, uz) // Ignore errors in benchmark
		_ = uz.Close()                 // Ignore errors in benchmark
	}
}

// BenchmarkUnzstdStream_Read benchmarks Read method
func BenchmarkUnzstdStream_Read(b *testing.B) {
	data := strings.Repeat("benchmark data ", 1000)
	compressed := compressZstd(b, data)

	reader := bytes.NewReader(compressed)
	uz, _ := NewUnzstdStream(reader)
	defer func() { _ = uz.Close() }() // Ignore error in benchmark cleanup

	buf := make([]byte, 4096)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = reader.Seek(0, io.SeekStart)
		uz, _ = NewUnzstdStream(reader)

		for {
			_, err := uz.Read(buf)
			if err == io.EOF {
				break
			}
		}
	}
}
