package middleware

import (
	"strings"

	"mallard/internal/reqctx"
	"mallard/internal/response"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
)

const (
	msgUnauthorized = "请先登录"
	msgTokenExpired = "登录已过期，请重新登录"
)

func Auth(jwtSecret string) echo.MiddlewareFunc {
	key := []byte(jwtSecret)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				return response.JsonError(c, 401, msgUnauthorized)
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return key, nil
			})
			if err != nil || !token.Valid {
				return response.JsonError(c, 401, msgTokenExpired)
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return response.JsonError(c, 401, msgTokenExpired)
			}

			uid, _ := claims["uid"].(string)
			username, _ := claims["username"].(string)
			if uid == "" || username == "" {
				return response.JsonError(c, 401, msgTokenExpired)
			}

			userID, err := bson.ObjectIDFromHex(uid)
			if err != nil {
				return response.JsonError(c, 401, msgTokenExpired)
			}

			reqctx.SetUser(c, &reqctx.UserCtx{
				ID:       userID,
				Username: username,
			})

			return next(c)
		}
	}
}
