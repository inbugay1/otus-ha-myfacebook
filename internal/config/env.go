package config

import (
	"log"

	"github.com/caarlos0/env/v6"
)

type EnvConfig struct {
	Version     string `env:"VERSION" envDefault:"version_not_set"`
	ServiceName string `env:"SERVICE_NAME" envDefault:"myfacebook"`
	LogLevel    string `env:"LOG_LEVEL" envDefault:"info"`
	HTTPPort    string `env:"HTTP_INT_PORT" envDefault:"9092"`

	RequestHeaderMaxSize                 int `env:"REQUEST_HEADER_MAX_SIZE" envDefault:"10000"`
	RequestReadHeaderTimeoutMilliseconds int `env:"REQUEST_READ_HEADER_TIMEOUT_MILLISECONDS" envDefault:"2000"`

	WriteDBDriverName         string `env:"WRITE_DB_DRIVER_NAME" envDefault:"postgres"`
	WriteDBHost               string `env:"WRITE_DB_HOST" envDefault:"localhost"`
	WriteDBPort               int    `env:"WRITE_DB_PORT" envDefault:"5432"`
	WriteDBUsername           string `env:"WRITE_DB_USERNAME" envDefault:"postgres"`
	WriteDBPassword           string `env:"WRITE_DB_PASSWORD" envDefault:"secret"`
	WriteDBName               string `env:"WRITE_DB_NAME" envDefault:"myfacebook"`
	WriteDBSSLMode            string `env:"WRITE_DB_SSL_MODE" envDefault:"disable"`
	WriteDBMaxOpenConnections int    `env:"WRITE_DB_MAX_OPEN_CONNECTIONS" envDefault:"10"`

	ReadDBDriverName         string `env:"READ_DB_DRIVER_NAME" envDefault:"postgres"`
	ReadDBHost               string `env:"READ_DB_HOST" envDefault:"localhost"`
	ReadDBPort               int    `env:"READ_DB_PORT" envDefault:"5432"`
	ReadDBUsername           string `env:"READ_DB_USERNAME" envDefault:"postgres"`
	ReadDBPassword           string `env:"READ_DB_PASSWORD" envDefault:"secret"`
	ReadDBName               string `env:"READ_DB_NAME" envDefault:"myfacebook"`
	ReadDBSSLMode            string `env:"READ_DB_SSL_MODE" envDefault:"disable"`
	ReadDBMaxOpenConnections int    `env:"READ_DB_MAX_OPEN_CONNECTIONS" envDefault:"10"`

	MyfacebookDialogAPIBaseURL string `env:"MYFACEBOOK_DIALOG_API_BASE_URL" envDefault:"http://localhost:9091"`

	OTelExporterType         string `env:"OTEL_EXPORTER_TYPE" envDefault:"stdout"`
	OTelExporterOTLPEndpoint string `env:"OTEL_EXPORTER_OTLP_ENDPOINT" envDefault:"localhost:4318"`
}

func GetConfigFromEnv() *EnvConfig {
	var config EnvConfig

	if err := env.Parse(&config); err != nil {
		log.Fatalf("unable to parse env config, error: %s", err)
	}

	return &config
}
