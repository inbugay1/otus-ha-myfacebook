package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/inbugay1/httprouter"
	"myfacebook/internal/config"
	"myfacebook/internal/httphandler"
	"myfacebook/internal/httpserver"
)

func main() {
	envConfig := config.GetConfigFromEnv()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

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
