// Package pipeline
package pipeline

import (
	"io"
	"slices"
)

type TransformR func(i io.Reader) (io.ReadCloser, error)
type PipelineReadCloser struct {
	stages []io.ReadCloser
}

func NewR(i io.ReadCloser, transforms ...TransformR) (*PipelineReadCloser, error) {
	var pipe PipelineReadCloser
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

func (iop *PipelineReadCloser) Close() {
	for _, s := range slices.Backward(iop.stages) {
		s.Close() // squelch errors... for now
	}
	iop.stages = make([]io.ReadCloser, 0)
}

func (iop PipelineReadCloser) Read(p []byte) (int, error) {
	return iop.stages[len(iop.stages)-1].Read(p)
}

// type TransformW func(i io.WriteCloser) (io.WriteCloser, error)
// type PipelineWriteCloser struct {
// 	stages []io.WriteCloser
// }
// 
// func NewW(i io.WriteCloser, transforms ...TransformW) (*PipelineWriteCloser, error) {
// 	var pipe PipelineWriteCloser
// 	pipe.stages = append(pipe.stages, i)
// 	for ix, transform := range transforms {
// 		prev := pipe.stages[ix]
// 		next, err := transform(prev)
// 		pipe.stages = append(pipe.stages, next)
// 		if err != nil {
// 			pipe.Close()
// 			return nil, err
// 		}
// 	}
// 	return &pipe, nil
// }
// 
// func (iop *PipelineWriteCloser) Close() {
// 	for _, s := range iop.stages {
// 		s.Close() // squelch errors... for now
// 	}
// 	iop.stages = make([]io.WriteCloser, 0)
// }
// 
// func (iop PipelineWriteCloser) Write(p []byte) (int, error) {
// 	return iop.stages[len(iop.stages)-1].Write(p)
// }
