package health

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Handler struct {
	mongoDB *mongo.Client
}

func NewHandler(client *mongo.Client) *Handler {
	return &Handler{mongoDB: client}
}

func (h *Handler) Liveness(c *echo.Context) error {
	return c.String(http.StatusOK, "OK")
}

func (h *Handler) Readiness(c *echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
	defer cancel()

	if err := h.mongoDB.Ping(ctx, nil); err != nil {
		return c.String(http.StatusServiceUnavailable, "unavailable")
	}
	return c.String(http.StatusOK, "OK")
}
