package server

import (
	"context"
	"io"
	"log/slog"
	"os"
	"time"

	myMiddleware "mallard/internal/middleware"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

type RegisterFunc func(e *echo.Echo)

func Run(ctx context.Context, listenAddr, logLevel string, register RegisterFunc) error {
	opts := &slog.HandlerOptions{}
	if logLevel == "debug" {
		opts.Level = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, opts)))

	e := echo.New()
	e.Logger = slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelWarn}))
	e.Validator = NewValidator()

	e.Use(middleware.Recover())
	e.Use(myMiddleware.RequestLogger())

	register(e)

	sc := echo.StartConfig{
		Address:         listenAddr,
		GracefulTimeout: 5 * time.Second,
		HideBanner:      true,
	}
	slog.Info("server starting", "address", listenAddr)
	return sc.Start(ctx, e)
}
