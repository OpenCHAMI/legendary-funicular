// Package render
package render

import (
	"fmt"
	"io"
	"reflect"

	"github.com/seantronsen/openchami-logq/query/internal/dev"
)

type Response struct {
	Message string `json:"message,omitempty"`
	Code    string `json:"code,omitempty"`
}

func Render(w io.Writer, f Format, v any) error {

	if v == nil {
		return fmt.Errorf("nil value")
	}

	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Pointer {
		if val.IsNil() {
			return fmt.Errorf("nil pointer")
		}
		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < val.Len(); i++ {
			elem := val.Index(i)
			if err := Render(w, f, elem.Interface()); err != nil {
				return err
			}
		}
		return nil

	case reflect.Struct:
		switch f {
		case FormatNDJSON:
			return renderNDJSON(w, v)
		case FormatJSON:
			return renderJSON(w, v)
		case FormatCSV:
			return renderCSV(w, v)
		case FormatText:
			return renderText(w, v)
		default:
			return fmt.Errorf("unknown format: %s", f)
		}

	default:
		return fmt.Errorf("expected struct or slice of struct")
	}
}

func renderNDJSON(w io.Writer, v any) error {
	return dev.NotImplemented()
}

func renderJSON(w io.Writer, v any) error {
	return dev.NotImplemented()
}

func renderCSV(w io.Writer, v any) error {
	return dev.NotImplemented()
}

func renderText(w io.Writer, v any) error {
	return dev.NotImplemented()
}

func ParseFormat(s string) (Format, error) {
	switch s {
	case "ndjson":
		return FormatNDJSON, nil
	case "json":
		return FormatJSON, nil
	case "csv":
		return FormatCSV, nil
	case "text":
		return FormatText, nil
	default:
		return "", fmt.Errorf("invalid format: %s", s)
	}
}
