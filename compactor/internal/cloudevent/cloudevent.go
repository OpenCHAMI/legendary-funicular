// Package cloudevent
package cloudevent

import "encoding/json"

type CloudEvent struct {
	ID          *string `json:"id"`
	Source      *string `json:"source"`
	Timestamp   *string `json:"ts"`
	SpecVersion *string `json:"spec_version"`
	Type        *string `json:"type"`
	CloudEvent  *string `json:"cloudevent"`
	Host        *string `json:"host"`

	// collector specific
	PayloadJSON *string `json:"payload_json"`
	ParseError  *string `json:"parse_error"`
}

func Parse(b []byte) (CloudEvent, error) {
	var record CloudEvent
	err := json.Unmarshal(b, &record)
	return record, err
}
