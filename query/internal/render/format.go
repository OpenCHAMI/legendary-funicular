package render

type Format string

const (
	FormatNDJSON Format = "ndjson"
	FormatJSON Format = "json"
	FormatCSV Format = "csv"
	FormatText Format = "text"
)
