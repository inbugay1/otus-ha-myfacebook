package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/inbugay1/httprouter"
	"myfacebook/internal/config"
	"myfacebook/internal/db"
	"myfacebook/internal/httphandler"
	"myfacebook/internal/httpserver"
)

func main() {
	envConfig := config.GetConfigFromEnv()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	db := db.New(db.Config{
		DriverName:    envConfig.DBDriverName,
		Host:          envConfig.DBHost,
		Port:          envConfig.DBPort,
		User:          envConfig.DBUsername,
		Password:      envConfig.DBPassword,
		DBName:        envConfig.DBName,
		MigrationPath: "./storage/migrations",
	})

	if err := db.Connect(context.Background()); err != nil {
		slog.Error(fmt.Sprintf("cannot connect to db: %s", err))
		os.Exit(1)
	}
	defer db.Disconnect()

	if err := db.Migrate(); err != nil {
		slog.Error(fmt.Sprintf("db migration failed: %s", err))
		os.Exit(1)
	}

	router := httprouter.New()

	router.Get("/health", &httphandler.Health{})

	httpServer := httpserver.New(httpserver.Config{
		Port:                          envConfig.HTTPPort,
		RequestMaxHeaderBytes:         envConfig.RequestHeaderMaxSize,
		ReadHeaderTimeoutMilliseconds: envConfig.RequestReadHeaderTimeoutMilliseconds,
	}, router)

	httpServer.Start()
	defer httpServer.Shutdown()

	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt, syscall.SIGTERM)

	slog.Info("got signal from OS: %v. Exit...", <-osSignals)
}
