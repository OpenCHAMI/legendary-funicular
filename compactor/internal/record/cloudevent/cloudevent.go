// Package cloudevent
package cloudevent

import (
	"encoding/json"
	"fmt"
)

type Event struct {
	ID          *string         `json:"id" parquet:"id"`
	Source      *string         `json:"source" parquet:"source"`
	Timestamp   *string         `json:"time" parquet:"ts"`
	SpecVersion *string         `json:"specversion" parquet:"specversion"`
	Type        *string         `json:"type" parquet:"type"`
	DataRaw     json.RawMessage `json:"data" parquet:"-"`
	Data        string          `json:"-" parquet:"data"`
}

type CloudEvent struct {
	Host        *string `json:"host" parquet:"host"`
	ParseError  *string `json:"parse_error" parquet:"parse_error"`
	PayloadJSON string  `json:"payload_json" parquet:"payload_json"`
	Event       Event   `json:"cloudevent" parquet:"cloudevent"`
}

func Parse(b []byte) (CloudEvent, error) {
	var record CloudEvent
	err := json.Unmarshal(b, &record)
	if err == nil {
		record.PayloadJSON = string(b)
		record.Event.Data = string(record.Event.DataRaw)
	}
	return record, err
}
