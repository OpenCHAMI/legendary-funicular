// SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

// Package cloudevent
package cloudevent

import "encoding/json"

type CloudEvent struct {
	// Core
	Timestamp *string `json:"ts" parquet:"ts"`

	TransportMethod   *string `json:"transport_method" parquet:"transport_method"`
	TransportMetadata *string `json:"transport_metadata" parquet:"transport_metadata"`

	Event            *string `json:"cloudevent" parquet:"cloudevent"`
	EventID          *string `json:"cloudevent_id" parquet:"cloudevent_id"`
	EventSource      *string `json:"cloudevent_source" parquet:"cloudevent_source"`
	EventType        *string `json:"cloudevent_type" parquet:"cloudevent_type"`
	EventSpecversion *string `json:"cloudevent_specversion" parquet:"cloudevent_specversion"`
	EventBinding     *string `json:"cloudevent_binding" parquet:"cloudevent_binding"`

	// OPENCHAMI SPECIFIC
	XName       *string `json:"xname" parquet:"xname"`
	IDComponent *string `json:"component_id" parquet:"component_id"`
	IDNode      *string `json:"node_id" parquet:"node_id"`
	IDTrace     *string `json:"trace_id" parquet:"trace_id"`
	IDRequest   *string `json:"request_id" parquet:"request_id"`
	RequestURI  *string `json:"request_uri" parquet:"request_uri"`
	RequestUser *string `json:"request_user" parquet:"request_user"`

	// PARSING
	Raw         *string `json:"raw" parquet:"raw"` // raw as operated on by datadog vector
	ParseError  *string `json:"parse_error" parquet:"parse_error"`
	PayloadJSON *string `json:"payload_json" parquet:"payload_json"`
}

func Parse(b []byte) (CloudEvent, error) {
	var record CloudEvent
	err := json.Unmarshal(b, &record)
	if err == nil {
		jsonStr := string(b)
		record.PayloadJSON = &jsonStr
	}
	return record, err
}
