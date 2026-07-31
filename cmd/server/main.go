package main

import (
	"context"
	"flag"
	"io"
	"log/slog"
	"myapi/internal/config"
	"myapi/internal/logger"
	myMiddleware "myapi/internal/middleware"
	"myapi/internal/router"
	"myapi/pkg/mysql"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Validator struct {
	validator *validator.Validate
}

func (v *Validator) Validate(i any) error {
	if err := v.validator.Struct(i); err != nil {
		return echo.ErrBadRequest.Wrap(err)
	}
	return nil
}

func main() {
	var configFile string
	flag.StringVar(&configFile, "config", "./config.yaml", "config file path")
	flag.Parse()

	v := viper.New()
	v.SetConfigFile(configFile)
	if err := v.ReadInConfig(); err != nil {
		panic(err.Error())
	}

	cfg, err := config.Load(v)
	if err != nil {
		panic(err.Error())
	}

	if err := logger.Init(cfg.Log.Level); err != nil {
		panic(err)
	}
	defer logger.Sync()

	db, err := mysql.Connect(cfg.Database.Main)
	if err != nil {
		logger.Fatal("init mysql failed", zap.Error(err))
	}
	defer db.Close()

	e := echo.New()
	e.Logger = slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelWarn}))
	e.Validator = &Validator{validator: validator.New()}

	e.Use(myMiddleware.RequestLogger(logger.GetLogger()))
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     cfg.Server.AllowOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD"},
		AllowHeaders:     []string{"Content-Type", "Authorization", "Origin", "Accept", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           86400,
	}))

	e.GET("/health", func(c *echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	router.Register(e, db, cfg)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	sc := echo.StartConfig{
		Address:         cfg.Server.ListenAddr,
		GracefulTimeout: 5 * time.Second,
		HideBanner:      true,
	}

	logger.Info("server starting", zap.String("address", cfg.Server.ListenAddr))
	if err := sc.Start(ctx, e); err != nil {
		logger.Fatal("failed to start server", zap.Error(err))
	}
}
