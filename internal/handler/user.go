package handler

import (
	"myapi/internal/dto"
	"myapi/internal/model"
	"myapi/internal/response"
	"myapi/internal/service"
	"myapi/internal/vo"

	"github.com/labstack/echo/v5"
)

type User struct {
	UserService  *service.User
	LoginService *service.Login
}

func (h *User) Get(c *echo.Context) error {
	user, ok := c.Get("user").(*model.User)
	if !ok || user == nil || user.ID <= 0 {
		return response.JsonError(c, 403, "access denied")
	}

	return response.JsonData(c, vo.ToUser(user))
}

func (h *User) UpdatePassword(c *echo.Context) error {
	user, ok := c.Get("user").(*model.User)
	if !ok || user == nil || user.ID <= 0 {
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

	_, isValid, err := h.LoginService.AuthUser(user.Username, in.Password)
	if err != nil || !isValid {
		return response.JsonError(c, 400, "验证失败")
	}

	err = h.UserService.UpdatePassword(user.ID, in.NewPassword)
	if err != nil {
		return response.JsonError(c, 500, err.Error())
	}

	h.LoginService.RevokeAllUserTokens(user.ID)

	return response.JsonData(c, "")
}
