package main

import (
	"context"
	"flag"
	"io/fs"
	"log"
	"os"
	"os/signal"
	"syscall"

	"mallard/internal/config"
	"mallard/internal/domain/app"
	"mallard/internal/domain/health"
	"mallard/internal/domain/span"
	"mallard/internal/domain/user"
	"mallard/internal/infra/repo"
	myMiddleware "mallard/internal/middleware"
	"mallard/internal/server"
	"mallard/internal/web"
	"mallard/pkg/mongodb"

	"github.com/dgraph-io/ristretto/v2"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func main() {
	var configFile string
	flag.StringVar(&configFile, "config", "./config.yaml", "config file path")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.LoadFile(configFile)
	if err != nil {
		log.Fatalf("load config failed: %v", err)
	}

	client, err := mongodb.Connect(ctx, cfg.MongoURI)
	if err != nil {
		log.Fatalf("connect mongodb failed: %v", err)
	}
	defer client.Disconnect(ctx)

	userRepo := repo.NewUserRepo(client)
	tokenRepo := repo.NewTokenRepo(client, int32(cfg.RefreshTTL))
	appRepo := repo.NewAppRepo(client)
	spanRepo := repo.NewSpanRepo(client, int32(cfg.SpanTTL))

	if err := repo.Migrate(ctx, userRepo, tokenRepo, appRepo, spanRepo); err != nil {
		log.Fatalf("migrate failed: %v", err)
	}

	appsCache, err := ristretto.NewCache(&ristretto.Config[string, app.AppEntity]{
		NumCounters: 1e6,
		MaxCost:     1 << 20,
		BufferItems: 64,
	})
	if err != nil {
		log.Fatalf("new cache failed: %v", err)
	}

	userService := user.NewService(userRepo, tokenRepo, &user.ServiceOptions{
		JWTSecret:                 cfg.JWTSecret,
		TokenSalt:                 cfg.TokenSalt,
		AccessTokenExpireSeconds:  cfg.AccessTTL,
		RefreshTokenExpireSeconds: cfg.RefreshTTL,
	})
	appService := app.NewService(appRepo, appsCache, &app.ServiceOptions{})
	spanService := span.NewService(spanRepo)

	register := func(e *echo.Echo) {
		e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins:     cfg.AllowOrigins,
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD"},
			AllowHeaders:     []string{"Content-Type", "Authorization", "Origin", "Accept", "X-Request-ID"},
			AllowCredentials: true,
			MaxAge:           86400,
		}))

		healthHandler := health.NewHandler(client)
		userHandler := user.NewHandler(userService, *cfg.CookieSecure)
		appHandler := app.NewHandler(appService)
		spanHandler := span.NewHandler(spanService)

		noAuth := e.Group("")
		noAuth.GET("/health", healthHandler.Liveness)
		noAuth.GET("/ready", healthHandler.Readiness)

		api := e.Group("/api")
		api.POST("/login", userHandler.Login, middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
			Store: middleware.NewRateLimiterMemoryStore(10),
			IdentifierExtractor: func(c *echo.Context) (string, error) {
				return c.RealIP(), nil
			},
		}))
		api.POST("/logout", userHandler.Logout)
		api.POST("/refresh", userHandler.RefreshToken)

		g := e.Group("/api")
		g.Use(myMiddleware.Auth(cfg.JWTSecret))

		g.GET("/user", userHandler.UserInfo)
		g.PUT("/user/password", userHandler.UpdatePassword)

		g.POST("/app", appHandler.Add)
		g.GET("/app", appHandler.List)
		g.PUT("/app/:id/ip-allow-list", appHandler.UpdateIPAllowList)
		g.PUT("/app/:id/secret", appHandler.UpdateSecret)
		g.DELETE("/app/:id", appHandler.Delete)

		g.GET("/traces", spanHandler.ListTraces)
		g.GET("/traces/:trace_id", spanHandler.GetTrace)

		sub, err := fs.Sub(web.Dist, "dist")
		if err != nil {
			log.Fatalf("web fs sub failed: %v", err)
		}
		ui := web.SPA(sub)
		e.GET("/", ui)
		e.GET("/*", ui)
	}

	if err := server.Run(ctx, cfg.ServerAddr, cfg.LogLevel, register); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
