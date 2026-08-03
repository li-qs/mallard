package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Health struct {
	Mongo *mongo.Client
}

func (h *Health) Liveness(c *echo.Context) error {
	return c.String(http.StatusOK, "OK")
}

func (h *Health) Readiness(c *echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
	defer cancel()

	if err := h.Mongo.Ping(ctx, nil); err != nil {
		return c.String(http.StatusServiceUnavailable, "unavailable")
	}
	return c.String(http.StatusOK, "OK")
}
