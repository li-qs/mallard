package router

import (
	"myapi/internal/config"
	"myapi/internal/handler"
	myMiddleware "myapi/internal/middleware"
	"myapi/internal/repository"
	"myapi/internal/service"
	"myapi/pkg/mysql"

	"github.com/labstack/echo/v5"
)

func Register(e *echo.Echo, db *mysql.DB, cfg *config.Config) {
	userRepo := &repository.User{DB: db}
	refreshTokenRepo := &repository.RefreshToken{DB: db}

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
