package app

import (
	"context"
	"errors"
	"log/slog"
	"mallard/internal/request"
	"mallard/internal/response"
	"time"

	"github.com/labstack/echo/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Handler struct {
	srv *Service
}

func NewHandler(srv *Service) *Handler {
	return &Handler{srv: srv}
}

type AddRequest struct {
	AppName     string   `json:"app_name" validate:"required"`
	IPAllowList []string `json:"ip_allow_list"`
}

type AddResponse struct {
	ID          string   `json:"id"`
	AppName     string   `json:"app_name"`
	Secret      string   `json:"secret"`
	IPAllowList []string `json:"ip_allow_list"`
}

func (h *Handler) Add(c *echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	var req AddRequest
	if err := c.Bind(&req); err != nil {
		return response.JsonError(c, 400, err.Error())
	}
	if err := c.Validate(&req); err != nil {
		return response.JsonError(c, 400, err.Error())
	}

	app, secret, err := h.srv.Add(ctx, req.AppName, req.IPAllowList)
	if err != nil {
		slog.Error("srv.Add failed", "error", err)
		if errors.Is(err, ErrAppExists) {
			return response.JsonError(c, 400, "app 名称已存在")
		}
		return response.JsonError(c, 500, "服务器错误")
	}

	return response.JsonData(c, AddResponse{
		ID:          app.ID.Hex(),
		AppName:     app.AppName,
		Secret:      secret,
		IPAllowList: app.IPAllowList,
	})
}

type ListResponse struct {
	ID          string   `json:"id"`
	AppName     string   `json:"app_name"`
	IPAllowList []string `json:"ip_allow_list"`
	CreatedAt   int64    `json:"created_at"`
	UpdatedAt   int64    `json:"updated_at"`
}

func (h *Handler) List(c *echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	page, pageSize := request.ParsePagination(c)

	filter := AppFilter{
		AppName: c.QueryParam("app_name"),
		ID:      c.QueryParam("id"),
	}

	apps, total, err := h.srv.List(ctx, filter, int64(page), int64(pageSize))
	if err != nil {
		slog.Error("srv.List failed", "error", err)
		return response.JsonError(c, 400, "无效的 app id")
	}

	list := make([]ListResponse, 0, len(apps))
	for i := range apps {
		app := apps[i]
		list = append(list, ListResponse{
			ID:          app.ID.Hex(),
			AppName:     app.AppName,
			IPAllowList: app.IPAllowList,
			CreatedAt:   app.CreatedAt.UnixMilli(),
			UpdatedAt:   app.UpdatedAt.UnixMilli(),
		})
	}
	return response.JsonList(c, list, page, pageSize, int(total))
}

type UpdateRequest struct {
	IPAllowList []string `json:"ip_allow_list"`
}

func (h *Handler) UpdateIPAllowList(c *echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	appID, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		return response.JsonError(c, 404, "ID 无效")
	}

	var req UpdateRequest
	if err := c.Bind(&req); err != nil {
		return response.JsonError(c, 400, err.Error())
	}

	if err := h.srv.UpdateIPAllowList(ctx, appID, req.IPAllowList); err != nil {
		slog.Error("srv.UpdateIPAllowList failed", "error", err)
		if errors.Is(err, mongo.ErrNoDocuments) {
			return response.JsonError(c, 404, "app 不存在")
		}
		return response.JsonError(c, 500, "服务器错误")
	}

	return response.JsonData(c, "")
}

type UpdateSecretResponse struct {
	ID     string `json:"id"`
	Secret string `json:"secret"`
}

func (h *Handler) UpdateSecret(c *echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	appID, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		return response.JsonError(c, 404, "ID 无效")
	}

	secret, err := h.srv.UpdateSecret(ctx, appID)
	if err != nil {
		slog.Error("srv.UpdateSecret failed", "error", err)
		if errors.Is(err, mongo.ErrNoDocuments) {
			return response.JsonError(c, 404, "app 不存在")
		}
		return response.JsonError(c, 500, "服务器错误")
	}

	return response.JsonData(c, UpdateSecretResponse{
		ID:     appID.Hex(),
		Secret: secret,
	})
}

func (h *Handler) Delete(c *echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	appID, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		return response.JsonError(c, 404, "ID 无效")
	}

	if err := h.srv.Delete(ctx, appID); err != nil {
		slog.Error("srv.Delete failed", "error", err)
		return response.JsonError(c, 500, "服务器错误")
	}

	return response.JsonData(c, "")
}
