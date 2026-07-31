package handler

import (
	"myapi/internal/dto"
	"myapi/internal/response"
	"myapi/internal/service"
	"myapi/internal/vo"
	"net/http"

	"github.com/labstack/echo/v5"
)

type Login struct {
	LoginService *service.Login
}

func (h *Login) Login(c *echo.Context) error {
	var in dto.Login
	if err := c.Bind(&in); err != nil {
		return response.JsonError(c, 400, "用户名或密码错误")
	}
	if err := c.Validate(&in); err != nil {
		return response.JsonError(c, 400, "用户名或密码错误")
	}
	user, isValid, err := h.LoginService.AuthUser(in.Username, in.Password)
	if err != nil || !isValid {
		return response.JsonError(c, 400, "用户名或密码错误")
	}

	accessToken, refreshToken, expireIn, err := h.LoginService.GenerateTokens(user)
	if err != nil {
		return response.JsonError(c, 500, "服务器错误")
	}
	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		MaxAge:   h.LoginService.RefreshTokenExpireSeconds,
		Path:     "/",
		Domain:   "",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	return response.JsonData(c, vo.Login{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   expireIn,
	})
}

func (h *Login) Logout(c *echo.Context) error {
	if cookie, err := c.Cookie("refresh_token"); err == nil {
		h.LoginService.Logout(cookie.Value)
	}
	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		MaxAge:   -1,
		Path:     "/",
		Domain:   "",
		SameSite: http.SameSiteStrictMode,
	})
	return response.JsonData(c, "")
}

func (h *Login) Refresh(c *echo.Context) error {
	cookie, err := c.Cookie("refresh_token")
	if err != nil {
		return response.JsonError(c, 401, "请登录")
	}

	accessToken, refreshToken, expireIn, err := h.LoginService.RefreshTokens(cookie.Value)
	if err != nil {
		c.SetCookie(&http.Cookie{
			Name:     "refresh_token",
			Value:    "",
			MaxAge:   -1,
			Path:     "/",
			Domain:   "",
			SameSite: http.SameSiteStrictMode,
		})
		return response.JsonError(c, 401, "请登录")
	}

	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		MaxAge:   h.LoginService.RefreshTokenExpireSeconds,
		Path:     "/",
		Domain:   "",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	return response.JsonData(c, vo.Login{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   expireIn,
	})
}
