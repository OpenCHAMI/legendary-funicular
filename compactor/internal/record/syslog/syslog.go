// SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

// Package syslog
package syslog

import "encoding/json"

// todo: revisit and reconsider if all fields should be pointers

type Syslog struct {
	// CORE
	Timestamp *string `json:"ts" parquet:"ts"`
	Service   *string `json:"service" parquet:"service"`
	Host      *string `json:"host" parquet:"host"`
	Level     *string `json:"level" parquet:"level"`
	Message   *string `json:"msg" parquet:"msg"`
	Data      *string `json:"data" parquet:"data"`
	// OPENCHAMI SPECIFIC
	XName       *string `json:"xname" parquet:"xname"`
	IDComponent *string `json:"component_id" parquet:"component_id"`
	IDNode      *string `json:"node_id" parquet:"node_id"`
	IDTrace     *string `json:"trace_id" parquet:"trace_id"`
	IDRequest   *string `json:"request_id" parquet:"request_id"`
	RequestURI  *string `json:"request_uri" parquet:"request_uri"`
	RequestUser *string `json:"request_user" parquet:"request_user"`
	// PARSING
	ParseError  *string `json:"parse_error" parquet:"parse_error"`
	PayloadJSON *string `json:"payload_json" parquet:"payload_json"`
}

func Parse(b []byte) (Syslog, error) {
	var record Syslog
	err := json.Unmarshal(b, &record)
	if err == nil {
		jsonStr := string(b)
		record.PayloadJSON = &jsonStr
	}
	return record, err
}
