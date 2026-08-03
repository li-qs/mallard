package service

import (
	"context"
	"mallard/internal/model"
	"mallard/internal/repository"
)

type Span struct {
	spanRepo *repository.Span
}

func NewSpan(spanRepo *repository.Span) *Span {
	return &Span{spanRepo: spanRepo}
}

func (s *Span) Report(ctx context.Context, spans []model.Span) error {
	return s.spanRepo.InsertMany(ctx, spans)
}

func (s *Span) GetByTraceID(ctx context.Context, traceID string) ([]model.Span, error) {
	return s.spanRepo.FindByTraceID(ctx, traceID)
}

func (s *Span) ListTraces(ctx context.Context, filter repository.TraceFilter, page, pageSize int) ([]model.TraceSummary, int64, error) {
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
