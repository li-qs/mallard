package handler

import (
	"context"
	"errors"
	"mallard/internal/dto"
	"mallard/internal/model"
	"mallard/internal/response"
	"mallard/internal/service"
	"mallard/internal/vo"
	"time"

	"github.com/labstack/echo/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type User struct {
	UserService  *service.User
	LoginService *service.Login
}

func (h *User) Get(c *echo.Context) error {
	user, ok := c.Get("user").(*model.User)
	if !ok || user == nil || user.ID == bson.NilObjectID {
		return response.JsonError(c, 403, "access denied")
	}

	return response.JsonData(c, vo.ToUser(user))
}

func (h *User) UpdatePassword(c *echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	user, ok := c.Get("user").(*model.User)
	if !ok || user == nil || user.ID == bson.NilObjectID {
		return response.JsonError(c, 403, "access denied")
	}

	var in dto.UpdatePassword
	if err := c.Bind(&in); err != nil {
		return response.JsonError(c, 400, err.Error())
	}
	if err := c.Validate(&in); err != nil {
		return response.JsonError(c, 400, err.Error())
	}

	if in.Password == "" || in.NewPassword == "" {
		return response.JsonError(c, 400, "参数错误")
	}

	_, isValid, err := h.LoginService.AuthUser(ctx, user.Username, in.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			return response.JsonError(c, 400, "验证失败")
		}
		return response.JsonError(c, 500, "服务器错误")
	}
	if !isValid {
		return response.JsonError(c, 400, "验证失败")
	}

	err = h.UserService.UpdatePassword(ctx, user.ID, in.NewPassword)
	if err != nil {
		return response.JsonError(c, 500, err.Error())
	}

	h.LoginService.RevokeAllUserTokens(ctx, user.ID)

	return response.JsonData(c, "")
}
