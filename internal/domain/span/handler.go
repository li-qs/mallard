package span

import (
	"context"
	"log/slog"
	"mallard/internal/reqctx"
	"mallard/internal/request"
	"mallard/internal/response"
	"strconv"
	"time"

	"github.com/labstack/echo/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Handler struct {
	srv *Service
}

func NewHandler(srv *Service) *Handler {
	return &Handler{srv: srv}
}

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

type ReportSpansRequest struct {
	Spans []Span `json:"spans" validate:"required,min=1"`
}

type ReportSpansResponse struct {
	Accepted int `json:"accepted"`
}

func (h *Handler) ReportSpans(c *echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	app, ok := reqctx.GetApp(c)
	if !ok || app == nil || app.ID == bson.NilObjectID {
		return response.JsonError(c, 401, "app ID 或 secret 不匹配")
	}

	var req ReportSpansRequest
	if err := c.Bind(&req); err != nil {
		return response.JsonError(c, 400, err.Error())
	}
	if err := c.Validate(&req); err != nil {
		return response.JsonError(c, 400, err.Error())
	}

	now := time.Now()
	spans := make([]SpanEntity, 0, len(req.Spans))
	for _, s := range req.Spans {
		spans = append(spans, SpanEntity{
			AppID:      app.ID,
			TraceID:    s.TraceID,
			SpanID:     s.SpanID,
			ParentID:   s.ParentID,
			Operation:  s.Operation,
			StartTime:  s.StartTime,
			Duration:   s.Duration,
			Status:     s.Status,
			Error:      s.Error,
			ReportedAt: now,
		})
	}

	if err := h.srv.Report(ctx, spans); err != nil {
		slog.Error("srv.Report failed", "error", err)
		return response.JsonError(c, 500, "服务器错误")
	}

	return response.JsonData(c, ReportSpansResponse{Accepted: len(spans)})
}

type GetTraceResponse struct {
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

func (h *Handler) GetTrace(c *echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	traceID := c.Param("trace_id")
	if traceID == "" {
		return response.JsonError(c, 400, "缺少 trace_id")
	}

	spans, err := h.srv.GetByTraceID(ctx, traceID)
	if err != nil {
		slog.Error("srv.GetByTraceID failed", "error", err)
		return response.JsonError(c, 500, "服务器错误")
	}

	list := make([]GetTraceResponse, 0, len(spans))
	for i := range spans {
		span := spans[i]
		list = append(list, GetTraceResponse{
			ID:         span.ID.Hex(),
			AppID:      span.AppID.Hex(),
			TraceID:    span.TraceID,
			SpanID:     span.SpanID,
			ParentID:   span.ParentID,
			IsRoot:     span.ParentID == "",
			Operation:  span.Operation,
			StartTime:  span.StartTime,
			Duration:   span.Duration,
			Status:     span.Status,
			Error:      span.Error,
			ReportedAt: span.ReportedAt.UnixMilli(),
		})
	}
	return response.JsonData(c, list)
}

type ListTracesResponse struct {
	TraceID    string   `json:"trace_id"`
	AppIDs     []string `json:"app_ids"`
	Operation  string   `json:"operation"`
	StartTime  int64    `json:"start_time"`
	Duration   int64    `json:"duration"`
	SpanCount  int      `json:"span_count"`
	ErrorCount int      `json:"error_count"`
	HasError   bool     `json:"has_error"`
}

func (h *Handler) ListTraces(c *echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	filter := TraceFilter{}
	if v := c.QueryParam("app_id"); v != "" {
		id, err := bson.ObjectIDFromHex(v)
		if err != nil {
			return response.JsonError(c, 400, "无效的 app_id")
		}
		filter.AppID = &id
	}
	if v := c.QueryParam("operation"); v != "" {
		filter.Operation = v
	}
	if v := c.QueryParam("status"); v != "" {
		s, err := strconv.Atoi(v)
		if err != nil || (s != 1 && s != 2) {
			return response.JsonError(c, 400, "无效的 status（1=成功，2=错误）")
		}
		filter.Status = &s
	}
	if v := c.QueryParam("trace_id"); v != "" {
		filter.TraceID = v
	}
	if v := c.QueryParam("start_time_gt"); v != "" {
		ts, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return response.JsonError(c, 400, "无效的 start_time_gt")
		}
		filter.StartTimeGT = ts
	}
	if v := c.QueryParam("start_time_lt"); v != "" {
		ts, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return response.JsonError(c, 400, "无效的 start_time_lt")
		}
		filter.StartTimeLT = ts
	}

	page, pageSize := request.ParsePagination(c)
	traces, total, err := h.srv.ListTraces(ctx, filter, page, pageSize)
	if err != nil {
		slog.Error("srv.ListTraces failed", "error", err)
		return response.JsonError(c, 500, "服务器错误")
	}

	list := make([]ListTracesResponse, 0, len(traces))
	for i := range traces {
		trace := traces[i]

		appIDs := make([]string, 0, len(trace.AppIDs))
		for _, id := range trace.AppIDs {
			appIDs = append(appIDs, id.Hex())
		}

		list = append(list, ListTracesResponse{
			TraceID:    trace.TraceID,
			AppIDs:     appIDs,
			Operation:  trace.Operation,
			StartTime:  trace.StartTime,
			Duration:   trace.Duration,
			SpanCount:  trace.SpanCount,
			ErrorCount: trace.ErrorCount,
			HasError:   trace.ErrorCount > 0,
		})
	}
	return response.JsonList(c, list, page, pageSize, int(total))
}
