// Package record
package record

import (
	"bufio"
	"io"
	"github.com/seantronsen/openchami-logq/compactor/internal/record/cloudevent"
	"github.com/seantronsen/openchami-logq/compactor/internal/record/syslog"
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
