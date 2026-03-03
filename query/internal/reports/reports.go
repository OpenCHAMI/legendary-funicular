// Package reports
package reports

import (
	"github.com/seantronsen/openchami-logq/query/internal/render"
	"github.com/seantronsen/openchami-logq/query/internal/sql"
)

// todo: future dev should enhance these canned recipes
// by allowing user specifed constraints (e.g.,
// where clauses, order, group, etc.)
//
// SELECT *
// FROM (<base query>) AS report
// WHERE ...
// ORDER BY ...

type Report interface {
	Name() string
	Description() string
	BuildQueryString() string

	// todo: this likely will have a lot... of repeated logic.
	// if we do a V2, we should abstract this out further.
	QueryRecords(engine *sql.Engine, sources []string, enc render.Encoder) error
}

var Registry = []Report{
	&ReportFindParseErrors{},
	&ReportFindServiceErrors{},
}
