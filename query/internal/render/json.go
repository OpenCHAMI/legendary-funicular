package render

import (
	"encoding/json"
	"io"
)

type encoderJSON struct {
	needsComma bool
	writer     io.Writer
	instance   *json.Encoder
}

func (e *encoderJSON) Encode(v any) error {
	if e.needsComma {
		if _, err := e.writer.Write([]byte{','}); err != nil {
			return err
		}
	}
	if err := e.instance.Encode(v); err != nil {
		return err
	}
	e.needsComma = true
	return nil

}
func (e *encoderJSON) Close() error {
	_, err := e.writer.Write([]byte{']'})
	return err
}

func newEncoderJSON(w io.Writer) (Encoder, error) {
	enc := &encoderJSON{instance: json.NewEncoder(w), writer: w}
	if _, err := w.Write([]byte{'['}); err != nil {
		return nil, err
	}
	return enc, nil
}
