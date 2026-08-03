package vo

import "mallard/internal/model"

type SpanReport struct {
	Accepted int `json:"accepted"`
}

type Span struct {
	ID         string `json:"id"`
	AppID      string `json:"app_id"`
	TraceID    string `json:"trace_id"`
	SpanID     string `json:"span_id"`
	ParentID   string `json:"parent_id"`
	IsRoot     bool   `json:"is_root"`
	Operation  string `json:"operation"`
	StartTime  int64  `json:"start_time"`
	Duration   int64  `json:"duration"`
	Status     int    `json:"status"`
	Error      string `json:"error"`
	ReportedAt int64  `json:"reported_at"`
}

func ToSpan(m *model.Span) Span {
	return Span{
		ID:         m.ID.Hex(),
		AppID:      m.AppID.Hex(),
		TraceID:    m.TraceID,
		SpanID:     m.SpanID,
		ParentID:   m.ParentID,
		IsRoot:     m.ParentID == "",
		Operation:  m.Operation,
		StartTime:  m.StartTime,
		Duration:   m.Duration,
		Status:     m.Status,
		Error:      m.Error,
		ReportedAt: m.ReportedAt.UnixMilli(),
	}
}

type TraceSummary struct {
	TraceID    string   `json:"trace_id"`
	AppIDs     []string `json:"app_ids"`
	Operation  string   `json:"operation"`
	StartTime  int64    `json:"start_time"`
	Duration   int64    `json:"duration"`
	SpanCount  int      `json:"span_count"`
	ErrorCount int      `json:"error_count"`
	HasError   bool     `json:"has_error"`
}

func ToTraceSummary(m *model.TraceSummary) TraceSummary {
	appIDs := make([]string, 0, len(m.AppIDs))
	for _, id := range m.AppIDs {
		appIDs = append(appIDs, id.Hex())
	}
	return TraceSummary{
		TraceID:    m.TraceID,
		AppIDs:     appIDs,
		Operation:  m.Operation,
		StartTime:  m.StartTime,
		Duration:   m.Duration,
		SpanCount:  m.SpanCount,
		ErrorCount: m.ErrorCount,
		HasError:   m.ErrorCount > 0,
	}
}
