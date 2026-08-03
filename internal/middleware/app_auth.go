package middleware

import (
	"context"
	"encoding/base64"
	"net/netip"
	"strings"
	"time"

	"mallard/internal/logger"
	"mallard/internal/response"
	"mallard/internal/service"

	"github.com/labstack/echo/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.uber.org/zap"
)

const (
	msgInvalidSecret = "app ID 或 secret 不匹配"
	msgIPNotAllowed  = "IP 不在 allow list 内"
)

func AppAuth(s *service.App) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Basic ") {
				return response.JsonError(c, 401, msgInvalidSecret)
			}

			raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(authHeader, "Basic "))
			if err != nil {
				return response.JsonError(c, 401, msgInvalidSecret)
			}

			parts := strings.Split(string(raw), ":")
			if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
				return response.JsonError(c, 401, msgInvalidSecret)
			}

			appID, err := bson.ObjectIDFromHex(parts[0])
			if err != nil {
				return response.JsonError(c, 401, msgInvalidSecret)
			}
			appSecret := parts[1]

			ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
			defer cancel()

			app, valid, err := s.ValidSecret(ctx, appID, appSecret)
			if err != nil {
				logger.Error("AppAuth failed", zap.Error(err))
				return response.JsonError(c, 401, msgInvalidSecret)
			}
			if !valid {
				return response.JsonError(c, 401, msgInvalidSecret)
			}
			if !ipAllowed(c.RealIP(), app.IPAllowList) {
				return response.JsonError(c, 403, msgIPNotAllowed)
			}

			c.Set("app", app)

			return next(c)
		}
	}
}

func ipAllowed(ip string, ipAllowList []string) bool {
	if len(ipAllowList) == 0 {
		return true
	}

	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return false
	}

	for _, entry := range ipAllowList {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		if prefix, err := netip.ParsePrefix(entry); err == nil {
			if prefix.Contains(addr) {
				return true
			}
			continue
		}
		if allowed, err := netip.ParseAddr(entry); err == nil && allowed.Compare(addr) == 0 {
			return true
		}
	}
	return false
}
