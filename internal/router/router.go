package router

import (
	"myapi/internal/config"
	"myapi/internal/handler"
	myMiddleware "myapi/internal/middleware"
	"myapi/internal/repository"
	"myapi/internal/service"

	"github.com/labstack/echo/v5"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func Register(e *echo.Echo, mongo *mongo.Client, cfg *config.Config) {
	userRepo := repository.NewUser(mongo)
	refreshTokenRepo := repository.NewRefreshToken(mongo)

	loginService := &service.Login{
		UserRepo:                  userRepo,
		RefreshTokenRepo:          refreshTokenRepo,
		JWTSecret:                 cfg.Auth.JWTSecret,
		AccessTokenExpireSeconds:  cfg.Auth.AccessTokenExpireSeconds,
		RefreshTokenExpireSeconds: cfg.Auth.RefreshTokenExpireSeconds,
	}

	userService := &service.User{
		UserRepo: userRepo,
	}

	login := handler.Login{LoginService: loginService}
	e.POST("/login", login.Login)
	e.POST("/logout", login.Logout)
	e.POST("/refresh", login.Refresh)

	protected := e.Group("")
	protected.Use(myMiddleware.Auth(cfg.Auth.JWTSecret))

	user := handler.User{UserService: userService, LoginService: loginService}
	protected.GET("/user", user.Get)
	protected.PUT("/user/password", user.UpdatePassword)
}
