// SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

// Package render provides streaming encoders for CLI output.
//
// The core abstraction is a logical sequence of records. Encoders consume
// individual records via Encode and write them incrementally to an io.Writer.
// This design avoids materializing large slices in memory and supports
// constant-memory streaming for dump-scale operations.
//
// Each format defines how a record sequence is framed:
//
//   - ndjson: each Encode emits one JSON value per line.
//   - json: records are framed as a single JSON array; "[" is written on
//     creation, elements are comma-delimited, and "]" is written on Close.
//   - csv: records are written row-by-row, optionally with a header.
//   - text: records are formatted line-by-line.
//
// Encoders operate on single records. Collection flattening (e.g. slice
// handling) is performed by higher-level helpers such as Render, which
// provides a convenience wrapper for small in-memory results.
//
// For large datasets, callers should construct an Encoder directly and
// stream records via repeated Encode calls followed by Close.
package render

// use render for one off use cases
// use encoder for long stream-like use cases

import (
	"fmt"
	"io"
	"reflect"

	"github.com/OpenCHAMI/legendary-funicular/query/internal/dev"
)

type Encoder interface {
	Encode(v any) error
	Close() error
}

func BuildEncoder(w io.Writer, f string) (Encoder, error) {
	switch f {
	case "ndjson":
		return newEncoderNDJSON(w)
	case "json":
		return newEncoderJSON(w)
	case "csv":
		return nil, dev.NotImplemented()
	case "text":
		return nil, dev.NotImplemented()
	default:
		return nil, fmt.Errorf("unknown format: %s", f)
	}
}

type Message struct {
	Message string `json:"message,omitempty"`
	Code    string `json:"code,omitempty"`
}

func Render(w io.Writer, f string, v any) error {
	var enc Encoder
	var err error

	switch f {
	case "ndjson":
		enc, err = newEncoderNDJSON(w)
	case "json":
		enc, err = newEncoderJSON(w)
	case "csv":
		return dev.NotImplemented()
	case "text":
		return dev.NotImplemented()
	default:
		return fmt.Errorf("unknown format: %s", f)
	}
	if err != nil {
		return err
	}
	defer enc.Close() //nolint:errcheck // error on close in defer is non-critical

	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.Struct:
		return enc.Encode(v)

	case reflect.Slice, reflect.Array:
		for i := range rv.Len() {
			elem := rv.Index(i).Interface()
			if err := enc.Encode(elem); err != nil {
				return fmt.Errorf("encoder error at slice index %d: %w", i, err)
			}
		}
		return nil
	default:
		return fmt.Errorf("expected struct or slice of struct")
	}
}

func MaskSecrets(v any) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return
	}

	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return
	}

	rt := rv.Type()

	for i := range rv.NumField() {
		fieldVal := rv.Field(i)
		fieldType := rt.Field(i)

		if !fieldVal.CanSet() {
			continue
		}

		if fieldType.Tag.Get("secret") != "true" {
			continue
		}

		if fieldVal.Kind() != reflect.Pointer ||
			fieldVal.Type().Elem().Kind() != reflect.String {
			continue
		}

		if fieldVal.IsNil() {
			continue
		}

		masked := "************************"
		fieldVal.Set(reflect.ValueOf(&masked))
	}
}
