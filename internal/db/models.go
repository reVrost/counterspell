package db



type Log struct {
	ID         int64
	Timestamp  int64 // Changed from string to int64
	Level      string
	Message    string
	TraceID    string // Changed from sql.NullString
	SpanID     string // Changed from sql.NullString
	Attributes []byte // Changed from sql.NullString to []byte
}

type Span struct {
	SpanID       string
	TraceID      string
	ParentSpanID string // Changed from sql.NullString
	Name         string
	StartTime    int64 // Changed from string to int64
	EndTime      int64 // Changed from string to int64
	DurationNs   int64
	Attributes   []byte // Changed from sql.NullString to []byte
	ServiceName  string
	HasError     bool
}
