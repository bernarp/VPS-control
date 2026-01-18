package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Database  DatabaseConfig  `yaml:"database"`
	Storage   StorageConfig   `yaml:"storage"`
	NATS      NATSConfig      `yaml:"nats"`
	JWT       JWTConfig       `yaml:"jwt"`
	RateLimit RateLimitConfig `yaml:"rate_limit"`
	Cookie    CookieConfig    `yaml:"cookie"`
}

type StorageConfig struct {
	LocalDBPath string `yaml:"local_db_path"`
}

type NATSConfig struct {
	URL     string        `yaml:"url"`
	Timeout time.Duration `yaml:"timeout"`
}

type ServerConfig struct {
	Port  string `yaml:"port"`
	Debug bool   `yaml:"debug"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
	SSLMode  string `yaml:"ssl_mode"`
}

func (db DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		db.User, db.Password, db.Host, db.Port, db.Name, db.SSLMode,
	)
}

type JWTConfig struct {
	Secret string        `yaml:"secret"`
	Issuer string        `yaml:"issuer"`
	TTL    time.Duration `yaml:"ttl"`
}

type RateLimitConfig struct {
	Auth AuthRateLimit `yaml:"auth"`
	API  APIRateLimit  `yaml:"api"`
}

type AuthRateLimit struct {
	MaxAttempts int           `yaml:"max_attempts"`
	BlockTime   time.Duration `yaml:"block_time"`
}

type APIRateLimit struct {
	Limit  int           `yaml:"limit"`
	Window time.Duration `yaml:"window"`
}

type CookieConfig struct {
	Name     string `yaml:"name"`
	Secure   bool   `yaml:"secure"`
	HttpOnly bool   `yaml:"http_only"`
	SameSite string `yaml:"same_site"`
}

func Load(path string) (*Config, error) {
	// #nosec G304
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	expanded := os.ExpandEnv(string(data))
	var cfg Config
	if err := yaml.Unmarshal([]byte(expanded), &cfg); err != nil {
		return nil, err
	}
	cfg.JWT.Secret = os.Getenv("JWT_SECRET")
	cfg.JWT.Issuer = getEnvOrDefault("JWT_ISSUER", cfg.JWT.Issuer)

	if cfg.Storage.LocalDBPath == "" {
		cfg.Storage.LocalDBPath = "./data/tokens.db"
	}

	return &cfg, nil
}

func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
