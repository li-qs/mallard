package middleware

import (
	"strings"

	"myapi/internal/model"
	"myapi/internal/response"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v5"
)

func Auth(jwtSecret string) echo.MiddlewareFunc {
	key := []byte(jwtSecret)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				return response.JsonError(c, 401, "请登录")
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return key, nil
			})
			if err != nil || !token.Valid {
				return response.JsonError(c, 401, "请登录")
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return response.JsonError(c, 401, "请登录")
			}

			uid, _ := claims["uid"].(float64)
			username, _ := claims["username"].(string)
			if uid == 0 || username == "" {
				return response.JsonError(c, 401, "请登录")
			}

			c.Set("user", &model.User{
				ID:       int64(uid),
				Username: username,
			})

			return next(c)
		}
	}
}
