package handler

import (
	"context"
	"errors"
	"net/http"
	"time"

	"mallard/internal/dto"
	"mallard/internal/logger"
	"mallard/internal/response"
	"mallard/internal/service"
	"mallard/internal/vo"

	"github.com/labstack/echo/v5"
	"go.uber.org/zap"
)

type Login struct {
	LoginService *service.Login
	CookieSecure bool
}

func (h *Login) Login(c *echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	var in dto.Login
	if err := c.Bind(&in); err != nil {
		return response.JsonError(c, 400, "用户名或密码错误")
	}
	if err := c.Validate(&in); err != nil {
		return response.JsonError(c, 400, "用户名或密码错误")
	}

	user, isValid, err := h.LoginService.AuthUser(ctx, in.Username, in.Password)
	if err != nil {
		logger.Error("AuthUser failed", zap.Error(err))
		if errors.Is(err, service.ErrInvalidCredentials) {
			return response.JsonError(c, 401, "用户名或密码错误")
		}
		return response.JsonError(c, 500, "服务器错误")
	}
	if !isValid {
		return response.JsonError(c, 401, "用户名或密码错误")
	}

	accessToken, refreshToken, expireIn, err := h.LoginService.GenerateTokens(ctx, user)
	if err != nil {
		logger.Error("GenerateTokens failed", zap.Error(err))
		return response.JsonError(c, 500, "服务器错误")
	}
	setRefreshCookie(c, refreshToken, h.LoginService.RefreshTokenExpireSeconds, h.CookieSecure)
	return response.JsonData(c, vo.Login{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   expireIn,
	})
}

func (h *Login) Logout(c *echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	if cookie, err := c.Cookie("refresh_token"); err == nil {
		h.LoginService.Logout(ctx, cookie.Value)
	}
	setRefreshCookie(c, "", -1, h.CookieSecure)
	return response.JsonData(c, "")
}

func (h *Login) Refresh(c *echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	cookie, err := c.Cookie("refresh_token")
	if err != nil {
		return response.JsonError(c, 401, "请登录")
	}

	accessToken, refreshToken, expireIn, err := h.LoginService.RefreshTokens(ctx, cookie.Value)
	if err != nil {
		logger.Error("RefreshTokens failed", zap.Error(err))
		if errors.Is(err, service.ErrInvalidRefreshToken) {
			setRefreshCookie(c, "", -1, h.CookieSecure)
			return response.JsonError(c, 401, "请登录")
		}
		return response.JsonError(c, 500, "服务器错误")
	}

	setRefreshCookie(c, refreshToken, h.LoginService.RefreshTokenExpireSeconds, h.CookieSecure)
	return response.JsonData(c, vo.Login{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   expireIn,
	})
}

func setRefreshCookie(c *echo.Context, value string, maxAge int, secure bool) {
	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    value,
		MaxAge:   maxAge,
		Path:     "/",
		Domain:   "",
		Secure:   secure,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}
