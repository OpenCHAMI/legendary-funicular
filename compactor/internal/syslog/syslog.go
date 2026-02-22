// Package syslog
package syslog

import "encoding/json"

type Syslog struct {
	Timestamp   *string `json:"timestamp" parquet:"ts"`
	Service     *string `json:"appname" parquet:"service"`
	Host        *string `json:"host" parquet:"host"`
	Severity    *string `json:"severity" parquet:"level"`
	Message     *string `json:"message" parquet:"msg"`
	XName       *string `json:"xname" parquet:"xname"`
	IDComponent *string `json:"component_id" parquet:"component_id"`
	IDNode      *string `json:"node_id" parquet:"node_id"`
	IDTrace     *string `json:"trace_id" parquet:"trace_id"`
	IDRequest   *string `json:"request_id" parquet:"request_id"`
	RequestUser *string `json:"request_user" parquet:"request_user"`
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
