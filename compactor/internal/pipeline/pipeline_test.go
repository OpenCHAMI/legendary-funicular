// SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

package pipeline

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockReadCloser wraps an io.Reader with a Close method
type mockReadCloser struct {
	io.Reader
	closed bool
}

func (m *mockReadCloser) Close() error {
	m.closed = true
	return nil
}

// TestNewR_NoTransforms tests creating pipeline with no transforms
func TestNewR_NoTransforms(t *testing.T) {
	input := &mockReadCloser{Reader: strings.NewReader("test data")}

	pipe, err := NewR(input)

	require.NoError(t, err)
	assert.NotNil(t, pipe)
	assert.Len(t, pipe.stages, 1)

	// Read should work
	data, err := io.ReadAll(pipe)
	require.NoError(t, err)
	assert.Equal(t, "test data", string(data))

	pipe.Close()
}

// TestNewR_SingleTransform tests pipeline with single transform
func TestNewR_SingleTransform(t *testing.T) {
	input := &mockReadCloser{Reader: strings.NewReader("hello")}

	// Transform that wraps reader
	transform := func(r io.Reader) (io.ReadCloser, error) {
		return &mockReadCloser{Reader: r}, nil
	}

	pipe, err := NewR(input, transform)

	require.NoError(t, err)
	assert.NotNil(t, pipe)
	assert.Len(t, pipe.stages, 2)

	// Read should work through pipeline
	data, err := io.ReadAll(pipe)
	require.NoError(t, err)
	assert.Equal(t, "hello", string(data))

	pipe.Close()
}

// TestNewR_MultipleTransforms tests pipeline with multiple transforms
func TestNewR_MultipleTransforms(t *testing.T) {
	input := &mockReadCloser{Reader: strings.NewReader("test")}

	transform1 := func(r io.Reader) (io.ReadCloser, error) {
		return &mockReadCloser{Reader: r}, nil
	}

	transform2 := func(r io.Reader) (io.ReadCloser, error) {
		return &mockReadCloser{Reader: r}, nil
	}

	transform3 := func(r io.Reader) (io.ReadCloser, error) {
		return &mockReadCloser{Reader: r}, nil
	}

	pipe, err := NewR(input, transform1, transform2, transform3)

	require.NoError(t, err)
	assert.NotNil(t, pipe)
	assert.Len(t, pipe.stages, 4) // input + 3 transforms

	// Read should work through all stages
	data, err := io.ReadAll(pipe)
	require.NoError(t, err)
	assert.Equal(t, "test", string(data))

	pipe.Close()
}

// TestNewR_TransformError tests error handling in transform
func TestNewR_TransformError(t *testing.T) {
	input := &mockReadCloser{Reader: strings.NewReader("test")}
	stage2 := &mockReadCloser{Reader: input}

	transform1 := func(r io.Reader) (io.ReadCloser, error) {
		return stage2, nil
	}

	// Transform that returns error
	transform2 := func(r io.Reader) (io.ReadCloser, error) {
		return nil, io.ErrUnexpectedEOF
	}

	pipe, err := NewR(input, transform1, transform2)

	assert.Error(t, err)
	assert.Nil(t, pipe)
	assert.Equal(t, io.ErrUnexpectedEOF, err)

	// Stages should be closed on error
	assert.True(t, input.closed)
	assert.True(t, stage2.closed)
}

// TestPipelineReadCloser_Close tests closing pipeline
func TestPipelineReadCloser_Close(t *testing.T) {
	stage1 := &mockReadCloser{Reader: strings.NewReader("test")}
	stage2 := &mockReadCloser{Reader: stage1}
	stage3 := &mockReadCloser{Reader: stage2}

	pipe := &PipelineReadCloser{
		stages: []io.ReadCloser{stage1, stage2, stage3},
	}

	pipe.Close()

	// All stages should be closed (in reverse order)
	assert.True(t, stage1.closed)
	assert.True(t, stage2.closed)
	assert.True(t, stage3.closed)

	// Stages should be cleared
	assert.Len(t, pipe.stages, 0)
}

// TestPipelineReadCloser_CloseEmpty tests closing empty pipeline
func TestPipelineReadCloser_CloseEmpty(t *testing.T) {
	pipe := &PipelineReadCloser{
		stages: []io.ReadCloser{},
	}

	// Should not panic
	pipe.Close()

	assert.Len(t, pipe.stages, 0)
}

// TestPipelineReadCloser_Read tests reading from pipeline
func TestPipelineReadCloser_Read(t *testing.T) {
	input := &mockReadCloser{Reader: strings.NewReader("hello world")}

	pipe := &PipelineReadCloser{
		stages: []io.ReadCloser{input},
	}

	buf := make([]byte, 5)
	n, err := pipe.Read(buf)

	require.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, "hello", string(buf))

	pipe.Close()
}

// TestPipelineReadCloser_ReadAll tests reading all data from pipeline
func TestPipelineReadCloser_ReadAll(t *testing.T) {
	input := &mockReadCloser{Reader: strings.NewReader("full data content")}

	pipe := &PipelineReadCloser{
		stages: []io.ReadCloser{input},
	}

	data, err := io.ReadAll(pipe)

	require.NoError(t, err)
	assert.Equal(t, "full data content", string(data))

	pipe.Close()
}

// TestPipelineReadCloser_ReadFromLastStage tests that Read uses last stage
func TestPipelineReadCloser_ReadFromLastStage(t *testing.T) {
	stage1 := &mockReadCloser{Reader: strings.NewReader("first")}
	stage2 := &mockReadCloser{Reader: strings.NewReader("second")}
	stage3 := &mockReadCloser{Reader: strings.NewReader("third")}

	pipe := &PipelineReadCloser{
		stages: []io.ReadCloser{stage1, stage2, stage3},
	}

	data, err := io.ReadAll(pipe)

	require.NoError(t, err)
	// Should read from last stage (stage3)
	assert.Equal(t, "third", string(data))

	pipe.Close()
}

// TestTransformR_Identity tests identity transform
func TestTransformR_Identity(t *testing.T) {
	input := &mockReadCloser{Reader: strings.NewReader("unchanged")}

	// Identity transform - passes through unchanged
	identity := func(r io.Reader) (io.ReadCloser, error) {
		return io.NopCloser(r), nil
	}

	pipe, err := NewR(input, identity)

	require.NoError(t, err)

	data, err := io.ReadAll(pipe)
	require.NoError(t, err)
	assert.Equal(t, "unchanged", string(data))

	pipe.Close()
}

// TestTransformR_BufferedReader tests transform with buffering
func TestTransformR_BufferedReader(t *testing.T) {
	input := &mockReadCloser{Reader: strings.NewReader("buffered data")}

	// Transform that adds buffering
	addBuffer := func(r io.Reader) (io.ReadCloser, error) {
		// Read all into buffer and return new reader
		data, err := io.ReadAll(r)
		if err != nil {
			return nil, err
		}
		return io.NopCloser(bytes.NewReader(data)), nil
	}

	pipe, err := NewR(input, addBuffer)

	require.NoError(t, err)

	data, err := io.ReadAll(pipe)
	require.NoError(t, err)
	assert.Equal(t, "buffered data", string(data))

	pipe.Close()
}

// TestTransformR_ChainedTransforms tests chaining multiple transforms
func TestTransformR_ChainedTransforms(t *testing.T) {
	input := &mockReadCloser{Reader: strings.NewReader("original")}

	// Multiple transforms in sequence
	transforms := []TransformR{
		func(r io.Reader) (io.ReadCloser, error) {
			return io.NopCloser(r), nil
		},
		func(r io.Reader) (io.ReadCloser, error) {
			return io.NopCloser(r), nil
		},
		func(r io.Reader) (io.ReadCloser, error) {
			return io.NopCloser(r), nil
		},
	}

	pipe, err := NewR(input, transforms...)

	require.NoError(t, err)
	assert.Len(t, pipe.stages, 4) // input + 3 transforms

	data, err := io.ReadAll(pipe)
	require.NoError(t, err)
	assert.Equal(t, "original", string(data))

	pipe.Close()
}

// TestNewR_ErrorClosesStages tests that error in transform closes previous stages
func TestNewR_ErrorClosesStages(t *testing.T) {
	stage1 := &mockReadCloser{Reader: strings.NewReader("test")}
	stage2 := &mockReadCloser{Reader: stage1}

	transform1 := func(r io.Reader) (io.ReadCloser, error) {
		return stage2, nil
	}

	// This transform will fail
	transform2 := func(r io.Reader) (io.ReadCloser, error) {
		return nil, io.ErrClosedPipe
	}

	_, err := NewR(stage1, transform1, transform2)

	assert.Error(t, err)
	assert.Equal(t, io.ErrClosedPipe, err)

	// All stages should be closed due to error cleanup
	assert.True(t, stage1.closed)
	assert.True(t, stage2.closed)
}

// TestPipelineReadCloser_MultipleClose tests calling Close multiple times
func TestPipelineReadCloser_MultipleClose(t *testing.T) {
	stage1 := &mockReadCloser{Reader: strings.NewReader("test")}

	pipe := &PipelineReadCloser{
		stages: []io.ReadCloser{stage1},
	}

	pipe.Close()
	assert.True(t, stage1.closed)
	assert.Len(t, pipe.stages, 0)

	// Second close should not panic
	pipe.Close()
	assert.Len(t, pipe.stages, 0)
}

// BenchmarkNewR_NoTransforms benchmarks pipeline creation with no transforms
func BenchmarkNewR_NoTransforms(b *testing.B) {
	data := strings.NewReader("benchmark data")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = data.Seek(0, io.SeekStart)
		input := &mockReadCloser{Reader: data}
		pipe, _ := NewR(input)
		pipe.Close()
	}
}

// BenchmarkNewR_SingleTransform benchmarks pipeline with single transform
func BenchmarkNewR_SingleTransform(b *testing.B) {
	data := strings.NewReader("benchmark data")
	transform := func(r io.Reader) (io.ReadCloser, error) {
		return io.NopCloser(r), nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = data.Seek(0, io.SeekStart)
		input := &mockReadCloser{Reader: data}
		pipe, _ := NewR(input, transform)
		pipe.Close()
	}
}

// BenchmarkNewR_MultipleTransforms benchmarks pipeline with multiple transforms
func BenchmarkNewR_MultipleTransforms(b *testing.B) {
	data := strings.NewReader("benchmark data")
	transforms := []TransformR{
		func(r io.Reader) (io.ReadCloser, error) { return io.NopCloser(r), nil },
		func(r io.Reader) (io.ReadCloser, error) { return io.NopCloser(r), nil },
		func(r io.Reader) (io.ReadCloser, error) { return io.NopCloser(r), nil },
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = data.Seek(0, io.SeekStart)
		input := &mockReadCloser{Reader: data}
		pipe, _ := NewR(input, transforms...)
		pipe.Close()
	}
}

// BenchmarkPipelineRead benchmarks reading from pipeline
func BenchmarkPipelineRead(b *testing.B) {
	content := strings.Repeat("benchmark data ", 1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		input := &mockReadCloser{Reader: strings.NewReader(content)}
		pipe, _ := NewR(input)
		_, _ = io.Copy(io.Discard, pipe)
		pipe.Close()
	}
}

// BenchmarkPipelineRead_WithTransforms benchmarks reading through transforms
func BenchmarkPipelineRead_WithTransforms(b *testing.B) {
	content := strings.Repeat("benchmark data ", 1000)
	transforms := []TransformR{
		func(r io.Reader) (io.ReadCloser, error) { return io.NopCloser(r), nil },
		func(r io.Reader) (io.ReadCloser, error) { return io.NopCloser(r), nil },
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		input := &mockReadCloser{Reader: strings.NewReader(content)}
		pipe, _ := NewR(input, transforms...)
		_, _ = io.Copy(io.Discard, pipe)
		pipe.Close()
	}
}
