package span

import (
	"context"
)

type Service struct {
	spanRepo SpanRepoImpl
}

func NewService(spanRepo SpanRepoImpl) *Service {
	return &Service{spanRepo: spanRepo}
}

func (s *Service) Report(ctx context.Context, spans []SpanEntity) error {
	return s.spanRepo.InsertMany(ctx, spans)
}

func (s *Service) GetByTraceID(ctx context.Context, traceID string) ([]SpanEntity, error) {
	return s.spanRepo.GetByTraceID(ctx, traceID)
}

func (s *Service) ListTraces(ctx context.Context, filter TraceFilter, page, pageSize int) ([]TraceEntity, int64, error) {
	traces, err := s.spanRepo.ListTraces(ctx, filter, int64((page-1)*pageSize), int64(pageSize))
	if err != nil {
		return nil, 0, err
	}
	total, err := s.spanRepo.CountTraces(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	return traces, total, nil
}
