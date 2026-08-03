package handler

import (
	"context"
	"errors"
	"mallard/internal/dto"
	"mallard/internal/logger"
	"mallard/internal/repository"
	"mallard/internal/response"
	"mallard/internal/service"
	"mallard/internal/vo"
	"time"

	"github.com/labstack/echo/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"
)

type App struct {
	AppService *service.App
}

func (h *App) Add(c *echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	var in dto.AddApp
	if err := c.Bind(&in); err != nil {
		return response.JsonError(c, 400, err.Error())
	}
	if err := c.Validate(&in); err != nil {
		return response.JsonError(c, 400, err.Error())
	}

	id, secret, err := h.AppService.Add(ctx, in.AppName, in.IPAllowList)
	if err != nil {
		logger.Error("AppService.Add failed", zap.Error(err))
		if errors.Is(err, service.ErrAppExists) {
			return response.JsonError(c, 400, "app 名称已存在")
		}
		return response.JsonError(c, 500, "服务器错误")
	}

	return response.JsonData(c, vo.AppAdd{
		ID:          id.Hex(),
		AppName:     in.AppName,
		Secret:      secret,
		IPAllowList: in.IPAllowList,
	})
}

func (h *App) List(c *echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	page, pageSize := parsePagination(c)

	filter := repository.AppFilter{
		AppName: c.QueryParam("app_name"),
		ID:      c.QueryParam("id"),
	}

	apps, total, err := h.AppService.List(ctx, filter, int64(page), int64(pageSize))
	if err != nil {
		logger.Error("AppService.List failed", zap.Error(err))
		if errors.Is(err, repository.ErrInvalidIDPrefix) {
			return response.JsonError(c, 400, "无效的 app id")
		}
		return response.JsonError(c, 500, "服务器错误")
	}

	list := make([]vo.App, 0, len(apps))
	for i := range apps {
		list = append(list, vo.ToApp(&apps[i]))
	}
	return response.JsonList(c, list, vo.Pagination{
		Page:     page,
		PageSize: pageSize,
		Total:    int(total),
	})
}

func (h *App) UpdateIPAllowList(c *echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	appID, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		return response.JsonError(c, 400, "无效的 app id")
	}

	var in dto.UpdateApp
	if err := c.Bind(&in); err != nil {
		return response.JsonError(c, 400, err.Error())
	}

	if err := h.AppService.UpdateIPAllowList(ctx, appID, in.IPAllowList); err != nil {
		logger.Error("AppService.UpdateIPAllowList failed", zap.Error(err))
		if errors.Is(err, mongo.ErrNoDocuments) {
			return response.JsonError(c, 404, "app 不存在")
		}
		return response.JsonError(c, 500, "服务器错误")
	}

	return response.JsonData(c, "")
}

func (h *App) UpdateSecret(c *echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	appID, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		return response.JsonError(c, 400, "无效的 app id")
	}

	secret, err := h.AppService.UpdateSecret(ctx, appID)
	if err != nil {
		logger.Error("AppService.UpdateSecret failed", zap.Error(err))
		if errors.Is(err, mongo.ErrNoDocuments) {
			return response.JsonError(c, 404, "app 不存在")
		}
		return response.JsonError(c, 500, "服务器错误")
	}

	return response.JsonData(c, vo.AppSecret{
		ID:     appID.Hex(),
		Secret: secret,
	})
}

func (h *App) Delete(c *echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	appID, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		return response.JsonError(c, 400, "无效的 app id")
	}

	if err := h.AppService.Delete(ctx, appID); err != nil {
		logger.Error("AppService.Delete failed", zap.Error(err))
		return response.JsonError(c, 500, "服务器错误")
	}

	return response.JsonData(c, "")
}
