package tracer

import (
	"fmt"
	"time"
)

type Span struct {
	SpanID    string `json:"span_id"`
	TraceID   string `json:"trace_id"`
	ParentID  string `json:"parent_id"`
	Service   string `json:"service"`
	Operation string `json:"operation"`
	StartTime int64  `json:"start_time"`
	Duration  int64  `json:"duration"` // 单位：纳秒
	Status    int    `json:"status"`   // HTTP 状态吗
	Error     string `json:"error,omitempty"`
}

func generateSpanID() string {
	return fmt.Sprintf("%x%x", time.Now().UnixNano(), time.Now().UnixNano()%1000000)
}
