package render

import (
	"encoding/json"
	"io"
)

type encoderNDJSON struct {
	instance *json.Encoder
}

func (e *encoderNDJSON) Encode(v any) error {
	return e.instance.Encode(v)
}
func (e *encoderNDJSON) Close() (err error) {
	return nil // no op
}

func newEncoderNDJSON(w io.Writer) (enc Encoder, err error) {
	enc = &encoderNDJSON{instance: json.NewEncoder(w)}
	return enc, err
}
