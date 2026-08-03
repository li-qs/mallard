package dto

type Span struct {
	TraceID   string `json:"trace_id" validate:"required"`
	SpanID    string `json:"span_id" validate:"required"`
	ParentID  string `json:"parent_id"`
	Operation string `json:"operation" validate:"required"`
	StartTime int64  `json:"start_time" validate:"required"`
	Duration  int64  `json:"duration" validate:"required"`
	Status    int    `json:"status"`
	Error     string `json:"error"`
}

type Spans struct {
	Spans []Span `json:"spans" validate:"required,min=1"`
}
