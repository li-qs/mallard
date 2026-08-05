package middleware

import (
	"log/slog"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func RequestLogger() echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogMethod:        true,
		LogURI:           true,
		LogStatus:        true,
		LogLatency:       true,
		LogRemoteIP:      true,
		LogContentLength: true,
		LogRequestID:     true,
		LogUserAgent:     true,
		LogValuesFunc: func(c *echo.Context, v middleware.RequestLoggerValues) error {
			args := []any{
				"method", v.Method,
				"uri", v.URI,
				"status", v.Status,
				"latency", v.Latency,
				"ip", v.RemoteIP,
				"content_length", v.ContentLength,
				"request_id", v.RequestID,
				"user_agent", v.UserAgent,
			}
			if v.Error != nil {
				args = append(args, "error", v.Error)
			}

			if v.Latency > time.Second {
				slog.Warn("slow request", args...)
			} else {
				slog.Info("request", args...)
			}

			return nil
		},
	})
}
