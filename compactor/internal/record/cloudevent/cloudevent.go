// Package cloudevent
package cloudevent

import "encoding/json"

type CloudEvent struct {
	ID            *string         `json:"id" parquet:"id"`
	Source        *string         `json:"source" parquet:"source"`
	Timestamp     *string         `json:"ts" parquet:"ts"`
	SpecVersion   *string         `json:"spec_version" parquet:"spec_version"`
	Type          *string         `json:"type" parquet:"type"`
	CloudEventRaw json.RawMessage `json:"cloudevent" parquet:"-"`
	CloudEvent    string          `json:"-" parquet:"cloudevent"`
	Host          *string         `json:"host"`
	PayloadJSON   string          `json:"-" parquet:"payload_json"`
	ParseError    *string         `json:"parse_error" parquet:"parse_error"`
}

func Parse(b []byte) (CloudEvent, error) {
	var record CloudEvent
	err := json.Unmarshal(b, &record)
	if err == nil {
		record.PayloadJSON = string(b)
		record.CloudEvent = string(record.CloudEventRaw)
	}
	return record, err
}
