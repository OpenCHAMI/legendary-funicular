// Package iopipe
package iopipe

import (
	"io"
	"slices"
)

type TransformR func(i io.Reader) (io.ReadCloser, error)
type IOPipeReadCloser struct {
	stages []io.ReadCloser
}

func NewR(i io.ReadCloser, transforms ...TransformR) (*IOPipeReadCloser, error) {
	var pipe IOPipeReadCloser
	pipe.stages = append(pipe.stages, i)
	for ix, transform := range transforms {
		prev := pipe.stages[ix]
		next, err := transform(prev)
		pipe.stages = append(pipe.stages, next)
		if err != nil {
			pipe.Close()
			return nil, err
		}
	}
	return &pipe, nil
}

func (iop *IOPipeReadCloser) Close() {
	for _, s := range slices.Backward(iop.stages) {
		s.Close() // squelch errors... for now
	}
	iop.stages = make([]io.ReadCloser, 0)
}

func (iop IOPipeReadCloser) Read(p []byte) (int, error) {
	return iop.stages[len(iop.stages)-1].Read(p)
}

type TransformW func(i io.WriteCloser) (io.WriteCloser, error)
type IOPipeWriteCloser struct {
	stages []io.WriteCloser
}

func NewW(i io.WriteCloser, transforms ...TransformW) (*IOPipeWriteCloser, error) {
	var pipe IOPipeWriteCloser
	pipe.stages = append(pipe.stages, i)
	for ix, transform := range transforms {
		prev := pipe.stages[ix]
		next, err := transform(prev)
		pipe.stages = append(pipe.stages, next)
		if err != nil {
			pipe.Close()
			return nil, err
		}
	}
	return &pipe, nil
}

func (iop *IOPipeWriteCloser) Close() {
	for _, s := range iop.stages {
		s.Close() // squelch errors... for now
	}
	iop.stages = make([]io.WriteCloser, 0)
}

func (iop IOPipeWriteCloser) Write(p []byte) (int, error) {
	return iop.stages[len(iop.stages)-1].Write(p)
}
