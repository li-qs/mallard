package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"mallard/internal/config"
	"mallard/internal/domain/app"
	"mallard/internal/domain/health"
	"mallard/internal/domain/span"
	"mallard/internal/infra/repo"
	myMiddleware "mallard/internal/middleware"
	"mallard/internal/server"
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

	appRepo := repo.NewAppRepo(client)
	spanRepo := repo.NewSpanRepo(client, int32(cfg.SpanTTL))

	if err := repo.Migrate(ctx, appRepo, spanRepo); err != nil {
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

	appService := app.NewService(appRepo, appsCache, &app.ServiceOptions{})
	spanService := span.NewService(spanRepo)

	register := func(e *echo.Echo) {
		e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins:     cfg.CollectorAllowOrigins,
			AllowMethods:     []string{"POST", "OPTIONS", "HEAD"},
			AllowHeaders:     []string{"Content-Type", "Authorization", "Origin", "Accept", "X-Request-ID"},
			AllowCredentials: true,
			MaxAge:           86400,
		}))

		healthHandler := health.NewHandler(client)
		spanHandler := span.NewHandler(spanService)

		noAuth := e.Group("")
		noAuth.GET("/health", healthHandler.Liveness)
		noAuth.GET("/ready", healthHandler.Readiness)

		g := e.Group("/api/v1")
		g.Use(myMiddleware.AppAuth(appService))
		g.POST("/spans", spanHandler.ReportSpans)
	}

	if err := server.Run(ctx, cfg.CollectorAddr, cfg.LogLevel, register); err != nil {
		log.Fatalf("collector failed: %v", err)
	}
}
