package handler

import (
	"context"
	"mallard/internal/dto"
	"mallard/internal/logger"
	"mallard/internal/model"
	"mallard/internal/repository"
	"mallard/internal/response"
	"mallard/internal/service"
	"mallard/internal/vo"
	"strconv"
	"time"

	"github.com/labstack/echo/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.uber.org/zap"
)

type Span struct {
	SpanService *service.Span
}

func (h *Span) Report(c *echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	app, ok := c.Get("app").(*model.App)
	if !ok || app == nil || app.ID == bson.NilObjectID {
		return response.JsonError(c, 401, "app ID 或 secret 不匹配")
	}

	var in dto.Spans
	if err := c.Bind(&in); err != nil {
		return response.JsonError(c, 400, err.Error())
	}
	if err := c.Validate(&in); err != nil {
		return response.JsonError(c, 400, err.Error())
	}

	now := time.Now()
	spans := make([]model.Span, 0, len(in.Spans))
	for _, s := range in.Spans {
		spans = append(spans, model.Span{
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

	if err := h.SpanService.Report(ctx, spans); err != nil {
		logger.Error("SpanService.Report failed", zap.Error(err))
		return response.JsonError(c, 500, "服务器错误")
	}

	return response.JsonData(c, vo.SpanReport{Accepted: len(spans)})
}

func (h *Span) GetTrace(c *echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	traceID := c.Param("trace_id")
	if traceID == "" {
		return response.JsonError(c, 400, "缺少 trace_id")
	}

	spans, err := h.SpanService.GetByTraceID(ctx, traceID)
	if err != nil {
		logger.Error("SpanService.GetByTraceID failed", zap.Error(err))
		return response.JsonError(c, 500, "服务器错误")
	}

	list := make([]vo.Span, 0, len(spans))
	for i := range spans {
		list = append(list, vo.ToSpan(&spans[i]))
	}
	return response.JsonData(c, list)
}

func (h *Span) ListTraces(c *echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	filter := repository.TraceFilter{}
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

	page, pageSize := parsePagination(c)
	traces, total, err := h.SpanService.ListTraces(ctx, filter, page, pageSize)
	if err != nil {
		logger.Error("SpanService.ListTraces failed", zap.Error(err))
		return response.JsonError(c, 500, "服务器错误")
	}

	list := make([]vo.TraceSummary, 0, len(traces))
	for i := range traces {
		list = append(list, vo.ToTraceSummary(&traces[i]))
	}
	return response.JsonList(c, list, vo.Pagination{
		Page:     page,
		PageSize: pageSize,
		Total:    int(total),
	})
}
