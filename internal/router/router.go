package router

import (
	"mallard/internal/config"
	"mallard/internal/handler"
	myMiddleware "mallard/internal/middleware"
	"mallard/internal/service"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Deps struct {
	LoginService *service.Login
	UserService  *service.User
	AppService   *service.App
	SpanService  *service.Span
	Mongo        *mongo.Client
}

func Register(e *echo.Echo, deps *Deps, cfg *config.Config) {
	health := handler.Health{Mongo: deps.Mongo}
	e.GET("/health", health.Liveness)
	e.GET("/ready", health.Readiness)

	login := handler.Login{LoginService: deps.LoginService, CookieSecure: cfg.Auth.CookieSecure}
	e.POST("/login", login.Login, middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Store: middleware.NewRateLimiterMemoryStore(10),
		IdentifierExtractor: func(c *echo.Context) (string, error) {
			return c.RealIP(), nil
		},
	}))
	e.POST("/logout", login.Logout)
	e.POST("/refresh", login.Refresh)

	manager := e.Group("")
	manager.Use(myMiddleware.Auth(cfg.Auth.JWTSecret))

	user := handler.User{UserService: deps.UserService, LoginService: deps.LoginService}
	manager.GET("/user", user.Get)
	manager.PUT("/user/password", user.UpdatePassword)

	app := handler.App{AppService: deps.AppService}
	manager.POST("/app", app.Add)
	manager.GET("/app", app.List)
	manager.PUT("/app/:id/ip-allow-list", app.UpdateIPAllowList)
	manager.PUT("/app/:id/secret", app.UpdateSecret)
	manager.DELETE("/app/:id", app.Delete)

	appAuth := e.Group("")
	appAuth.Use(myMiddleware.AppAuth(deps.AppService))

	span := handler.Span{SpanService: deps.SpanService}
	appAuth.POST("/spans", span.Report)

	manager.GET("/traces/:trace_id", span.GetTrace)
	manager.GET("/traces", span.ListTraces)
}
