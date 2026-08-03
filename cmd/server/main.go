package main

import (
	"context"
	"flag"
	"io"
	"log/slog"
	"mallard/internal/config"
	"mallard/internal/logger"
	myMiddleware "mallard/internal/middleware"
	"mallard/internal/repository"
	"mallard/internal/router"
	"mallard/internal/service"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
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
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var configFile string
	flag.StringVar(&configFile, "config", "./config.yaml", "config file path")
	flag.Parse()

	v := viper.New()
	v.SetConfigFile(configFile)
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
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

	mongo, err := mongo.Connect(options.Client().ApplyURI(cfg.Mongodb.Main))
	if err != nil {
		logger.Fatal("init mongodb failed", zap.Error(err))
	}
	defer mongo.Disconnect(ctx)

	pingCtx, cancelPing := context.WithTimeout(ctx, 5*time.Second)
	defer cancelPing()
	if err := mongo.Ping(pingCtx, nil); err != nil {
		logger.Fatal("ping mongodb failed", zap.Error(err))
	}

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

	userRepo := repository.NewUser(mongo)
	refreshTokenRepo := repository.NewRefreshToken(mongo, int32(cfg.Auth.RefreshTokenExpireSeconds))
	appRepo := repository.NewApp(mongo)
	spanRepo := repository.NewSpan(mongo, int32(cfg.TracingData.SpansExpireSeconds))

	if err := repository.Migrate(ctx, userRepo, refreshTokenRepo, appRepo, spanRepo); err != nil {
		logger.Fatal("migrate failed", zap.Error(err))
	}

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
	appService := service.NewApp(appRepo)
	spanService := service.NewSpan(spanRepo)

	router.Register(e, &router.Deps{
		LoginService: loginService,
		UserService:  userService,
		AppService:   appService,
		SpanService:  spanService,
		Mongo:        mongo,
	}, cfg)

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
