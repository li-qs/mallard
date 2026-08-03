package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Log         LogConfig         `mapstructure:"log"`
	Server      ServerConfig      `mapstructure:"server"`
	Mongodb     MongodbConfig     `mapstructure:"mongodb"`
	Auth        AuthConfig        `mapstructure:"auth"`
	TracingData TracingDataConfig `mapstructure:"tracing_data"`
}

type LogConfig struct {
	Level string `mapstructure:"level"`
}

type ServerConfig struct {
	ListenAddr   string   `mapstructure:"listen_addr"`
	AllowOrigins []string `mapstructure:"allow_origins"`
}

type MongodbConfig struct {
	Main string `mapstructure:"main"`
}

type AuthConfig struct {
	JWTSecret                 string `mapstructure:"jwt_secret"`
	AccessTokenExpireSeconds  int    `mapstructure:"access_token_expire_seconds"`
	RefreshTokenExpireSeconds int    `mapstructure:"refresh_token_expire_seconds"`
	CookieSecure              bool   `mapstructure:"cookie_secure"`
}

type TracingDataConfig struct {
	SpansExpireSeconds int `mapstructure:"spans_expire_seconds"`
}

func Load(v *viper.Viper) (*Config, error) {
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	setDefaults(&cfg)
	if !v.IsSet("auth.cookie_secure") {
		cfg.Auth.CookieSecure = true
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func setDefaults(cfg *Config) {
	if cfg.Server.ListenAddr == "" {
		cfg.Server.ListenAddr = ":8080"
	}
	if cfg.Auth.AccessTokenExpireSeconds == 0 {
		cfg.Auth.AccessTokenExpireSeconds = 900
	}
	if cfg.Auth.RefreshTokenExpireSeconds == 0 {
		cfg.Auth.RefreshTokenExpireSeconds = 604800
	}
}

func (c *Config) Validate() error {
	if c.Mongodb.Main == "" {
		return fmt.Errorf("mongodb.main is required")
	}
	if c.Auth.JWTSecret == "" {
		return fmt.Errorf("auth.jwt_secret is required")
	}
	return nil
}
