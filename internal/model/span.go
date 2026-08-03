package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

const (
	SpanDB   = "mallard_tracing"
	SpanColl = "span"
)

type Span struct {
	ID         bson.ObjectID `bson:"_id,omitempty"`
	AppID      bson.ObjectID `bson:"app_id"`
	TraceID    string        `bson:"trace_id"`
	SpanID     string        `bson:"span_id"`
	ParentID   string        `bson:"parent_id"`
	Operation  string        `bson:"operation"`
	StartTime  int64         `bson:"start_time"`
	Duration   int64         `bson:"duration"` // 单位：纳秒
	Status     int           `bson:"status"`   // HTTP 状态吗
	Error      string        `bson:"error,omitempty"`
	ReportedAt time.Time     `bson:"reported_at"`
}

type TraceSummary struct {
	TraceID    string          `bson:"trace_id"`
	AppIDs     []bson.ObjectID `bson:"app_ids"`
	Operation  string          `bson:"operation"`
	StartTime  int64           `bson:"start_time"`
	Duration   int64           `bson:"duration"`
	SpanCount  int             `bson:"span_count"`
	ErrorCount int             `bson:"error_count"`
}
