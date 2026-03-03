// Package render
package render

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"

	"github.com/seantronsen/openchami-logq/query/internal/dev"
)

func validate(v any) error {

	if v == nil {
		return fmt.Errorf("nil value")
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.Struct:
		return nil

	case reflect.Slice, reflect.Array: // ensure non-empty and element type is struct
		elemType := rv.Type().Elem()
		if elemType.Kind() == reflect.Pointer {
			elemType = elemType.Elem()
		}

		if elemType.Kind() != reflect.Struct {
			return fmt.Errorf("slice element must be a struct, got %s", elemType.Kind())
		}
		return nil

	default:
		return fmt.Errorf("v must be a struct or slice/array of structs, got %s", rv.Kind())
	}
}

type Message struct {
	Message string `json:"message,omitempty"`
	Code    string `json:"code,omitempty"`
}

func Render(w io.Writer, f string, v any) error {
	if err := validate(v); err != nil {
		return err
	}

	switch f {
	case "ndjson":
		return renderNDJSON(w, v)
	case "json":
		return renderJSON(w, v)
	case "csv":
		// return renderCSV(w, v)
		return dev.NotImplemented()
	case "text":
		// return renderText(w, v)
		return dev.NotImplemented()
	default:
		return fmt.Errorf("unknown format: %s", f)
	}

}

func renderJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	return enc.Encode(v)
}

func renderNDJSON(w io.Writer, v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.Struct:
		return renderJSON(w, v)

	case reflect.Slice, reflect.Array:
		for i := range rv.Len() {
			elem := rv.Index(i).Interface()
			if err := renderJSON(w, elem); err != nil {
				return fmt.Errorf("ndjson: error at index %d: %w", i, err)
			}
		}
		return nil

	default:
		return fmt.Errorf("expected struct or slice of struct")
	}
}

// func renderCSV(w io.Writer, v any) error {
// 	return dev.NotImplemented()
// }
//
// func renderText(w io.Writer, v any) error {
// 	return dev.NotImplemented()
// }
