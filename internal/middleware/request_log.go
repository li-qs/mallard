package middleware

import (
	"time"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"go.uber.org/zap"
)

func RequestLogger(logger *zap.Logger) echo.MiddlewareFunc {
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
			fields := []zap.Field{
				zap.String("method", v.Method),
				zap.String("uri", v.URI),
				zap.Int("status", v.Status),
				zap.Duration("latency", v.Latency),
				zap.String("ip", v.RemoteIP),
				zap.String("content_length", v.ContentLength),
				zap.String("request_id", v.RequestID),
				zap.String("user_agent", v.UserAgent),
			}

			if v.Error != nil {
				fields = append(fields, zap.Error(v.Error))
			}

			if v.Latency > time.Second {
				logger.Warn("slow request", fields...)
			} else {
				logger.Info("request", fields...)
			}

			return nil
		},
	})
}
