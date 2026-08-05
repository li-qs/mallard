package config

import (
	"fmt"
	"os"

	"go.yaml.in/yaml/v4"
)

const (
	defaultServerAddr    = ":9010"
	defaultCollectorAddr = ":9011"
	defaultAccessTTL     = 900
	defaultRefreshTTL    = 604800
	defaultSpanTTL       = 604800
	defaultCookieSecure  = true
)

type Config struct {
	ServerAddr            string   `yaml:"server_addr"`
	CollectorAddr         string   `yaml:"collector_addr"`
	AllowOrigins          []string `yaml:"allow_origins"`
	CollectorAllowOrigins []string `yaml:"collector_allow_origins"`
	LogLevel              string   `yaml:"log_level"`
	MongoURI              string   `yaml:"mongo_uri"`
	JWTSecret             string   `yaml:"jwt_secret"`
	AccessTTL             int      `yaml:"access_ttl"`
	RefreshTTL            int      `yaml:"refresh_ttl"`
	SpanTTL               int      `yaml:"span_ttl"`
	CookieSecure          *bool    `yaml:"cookie_secure"`
}

func LoadFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if cfg.ServerAddr == "" {
		cfg.ServerAddr = defaultServerAddr
	}
	if cfg.CollectorAddr == "" {
		cfg.CollectorAddr = defaultCollectorAddr
	}
	if cfg.AccessTTL == 0 {
		cfg.AccessTTL = defaultAccessTTL
	}
	if cfg.RefreshTTL == 0 {
		cfg.RefreshTTL = defaultRefreshTTL
	}
	if cfg.SpanTTL == 0 {
		cfg.SpanTTL = defaultSpanTTL
	}
	if cfg.CookieSecure == nil {
		secure := defaultCookieSecure
		cfg.CookieSecure = &secure
	}

	if cfg.MongoURI == "" {
		return nil, fmt.Errorf("mongo_uri is required")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("jwt_secret is required")
	}

	return &cfg, nil
}
