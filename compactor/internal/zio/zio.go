// SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

// Package zio
package zio

import (
	"io"

	"github.com/klauspost/compress/zstd"
)

type UnzstdStream struct {
	stream *zstd.Decoder
}

func NewUnzstdStream(stream io.Reader) (io.ReadCloser, error) {
	uz, err := zstd.NewReader(stream)
	return UnzstdStream{stream: uz}, err
}

func (s UnzstdStream) Close() error {
	s.stream.Close()
	return nil // abide by iopipe
}

func (s UnzstdStream) Read(p []byte) (int, error) {
	return s.stream.Read(p)
}
