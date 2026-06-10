// SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

// Package record
package record

import (
	"bufio"
	"github.com/OpenCHAMI/legendary-funicular/compactor/internal/record/cloudevent"
	"github.com/OpenCHAMI/legendary-funicular/compactor/internal/record/syslog"
	"io"
)

type Record interface {
	syslog.Syslog | cloudevent.CloudEvent
}
type Parser[T Record] func(b []byte) (T, error)

func FromStream[T Record](stream io.Reader, parser Parser[T]) ([]T, error) {
	var records []T
	reader := bufio.NewReader(stream)
	for {
		bytes, err := reader.ReadBytes('\n')
		if err == io.EOF || len(bytes) == 0 {
			break
		}
		if err != nil {
			return records, err
		}
		s, err := parser(bytes)
		if err != nil {
			return records, err
		}
		records = append(records, s)
	}

	return records, nil
}
