package config

import (
	"log"

	"github.com/caarlos0/env/v6"
)

type EnvConfig struct {
	Version  string `env:"VERSION" envDefault:"version_not_set"`
	LogLevel string `env:"LOG_LEVEL" envDefault:"info"`
	HTTPPort string `env:"HTTP_INT_PORT" envDefault:"9090"`

	RequestHeaderMaxSize                 int `env:"REQUEST_HEADER_MAX_SIZE" envDefault:"10000"`
	RequestReadHeaderTimeoutMilliseconds int `env:"REQUEST_READ_HEADER_TIMEOUT_MILLISECONDS" envDefault:"2000"`

	DBDriverName string `env:"DB_DRIVER_NAME" envDefault:"postgres"`
	DBHost       string `env:"DB_HOST" envDefault:"localhost"`
	DBPort       int    `env:"DB_PORT" envDefault:"5432"`
	DBUsername   string `env:"DB_USERNAME" envDefault:"postgres"`
	DBPassword   string `env:"DB_PASSWORD" envDefault:"secret"`
	DBName       string `env:"DB_NAME" envDefault:"postgres"`
	DBSSLMode    string `env:"DB_SSL_MODE" envDefault:"disable"`
}

func GetConfigFromEnv() *EnvConfig {
	var config EnvConfig

	if err := env.Parse(&config); err != nil {
		log.Fatalf("unable to parse env config, error: %s", err)
	}

	return &config
}
