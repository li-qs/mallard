package span

import (
	"context"
)

type SpanRepoImpl interface {
	InsertMany(ctx context.Context, spans []SpanEntity) error
	GetByTraceID(ctx context.Context, traceID string) ([]SpanEntity, error)
	ListTraces(ctx context.Context, filter TraceFilter, skip, limit int64) ([]TraceEntity, error)
	CountTraces(ctx context.Context, filter TraceFilter) (int64, error)
}
