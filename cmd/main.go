package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/inbugay1/httprouter"
	"myfacebook/internal/apiclient"
	"myfacebook/internal/apiv1/handler"
	apiv1middleware "myfacebook/internal/apiv1/middleware"
	"myfacebook/internal/config"
	"myfacebook/internal/db"
	"myfacebook/internal/dialogapiclient"
	"myfacebook/internal/httpclient"
	"myfacebook/internal/httphandler"
	httproutermiddleware "myfacebook/internal/httprouter/middleware"
	"myfacebook/internal/httpserver"
	internalapihandler "myfacebook/internal/internalapi/handler"
	internalapimiddleware "myfacebook/internal/internalapi/middleware"
	"myfacebook/internal/repository/rest"
	sqlxrepo "myfacebook/internal/repository/sqlx"
)

func main() {
	if err := run(); err != nil {
		slog.Error(fmt.Sprintf("Application error: %s", err))

		os.Exit(1)
	}
}

func run() error {
	envConfig := config.GetConfigFromEnv()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel(envConfig.LogLevel),
	}))
	slog.SetDefault(logger)

	appDB := db.New(db.Config{
		DriverName:         envConfig.DBDriverName,
		Host:               envConfig.DBHost,
		Port:               envConfig.DBPort,
		User:               envConfig.DBUsername,
		Password:           envConfig.DBPassword,
		DBName:             envConfig.DBName,
		SSLMode:            envConfig.DBSSLMode,
		MaxOpenConnections: envConfig.DBMaxOpenConnections,
		MigrationPath:      "./storage/migrations",
	})

	if err := appDB.Connect(context.Background()); err != nil {
		return fmt.Errorf("cannot connect to appDB: %w", err)
	}

	defer func() {
		if err := appDB.Disconnect(); err != nil {
			slog.Error(fmt.Sprintf("Failed to disconnect from app db: %s", err))
		}
	}()

	if err := appDB.Migrate(); err != nil {
		return fmt.Errorf("appDB migration failed: %w", err)
	}

	httpClient := httpclient.New(&httpclient.Config{
		InsecureSkipVerify: true,
	})
	apiClient := apiclient.New(envConfig.DialogAPIBaseURL, httpClient)
	dialogAPIClient := dialogapiclient.New(apiClient)

	userRepository := sqlxrepo.NewUserRepository(appDB)
	sqlxDialogRepository := sqlxrepo.NewDialogRepository(appDB)
	restDialogRepository := rest.NewDialogRepository(dialogAPIClient)

	router := httprouter.New(httprouter.NewRegexRouteFactory())

	requestResponseMiddleware := httproutermiddleware.NewRequestResponseLog()

	apiv1ErrorResponseMiddleware := apiv1middleware.NewErrorResponse()
	apiv1ErrorLogMiddleware := apiv1middleware.NewErrorLog()
	apiv1AuthMiddleware := apiv1middleware.NewAuth(userRepository)

	router.Use(requestResponseMiddleware)

	router.Get("/health", &httphandler.Health{})

	router.Group(func(router httprouter.Router) {
		router.Use(apiv1ErrorResponseMiddleware, apiv1ErrorLogMiddleware)

		router.Post("/user/register", &handler.Register{UserRepository: userRepository})
		router.Get(`/user/{id:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}}`,
			&handler.GetUser{UserRepository: userRepository})
		router.Post("/login", &handler.Login{
			UserRepository: userRepository,
		})
		router.Get("/user/search", &handler.SearchUser{
			UserRepository: userRepository,
		})
		router.Get(`/user/findByToken/{token:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}}`,
			&handler.FindUserByToken{UserRepository: userRepository})

		router.Group(func(router httprouter.Router) {
			router.Use(apiv1AuthMiddleware)

			router.Put(`/friend/set/{id:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}}`, &handler.SetFriend{
				UserRepository: userRepository,
			})

			router.Put(`/friend/delete/{id:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}}`, &handler.DeleteFriend{
				UserRepository: userRepository,
			})

			router.Post(`/dialog/{user_id:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}}/send`, &handler.SendDialog{
				SqlxDialogRepository: sqlxDialogRepository,
				RestDialogRepository: restDialogRepository,
			})

			router.Get(`/dialog/{user_id:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}}/list`, &handler.ListDialog{
				DialogRepository: restDialogRepository,
			})
		})
	})

	internalAPIErrorResponseMiddleware := internalapimiddleware.NewErrorResponse()
	internalAPIErrorLogMiddleware := internalapimiddleware.NewErrorLog()

	router.Group(func(router httprouter.Router) {
		router.Use(internalAPIErrorResponseMiddleware, internalAPIErrorLogMiddleware)

		router.Get(`/int/user/findByToken/{token:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}}`,
			&internalapihandler.FindUserByToken{UserRepository: userRepository})
	})

	httpServer := httpserver.New(httpserver.Config{
		Port:                          envConfig.HTTPPort,
		RequestMaxHeaderBytes:         envConfig.RequestHeaderMaxSize,
		ReadHeaderTimeoutMilliseconds: envConfig.RequestReadHeaderTimeoutMilliseconds,
	}, router)

	httpServerErrCh := httpServer.Start()
	defer httpServer.Shutdown()

	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt, syscall.SIGTERM)

	select {
	case osSignal := <-osSignals:
		slog.Info(fmt.Sprintf("got signal from OS: %v. Exit...", osSignal))
	case err := <-httpServerErrCh:
		return fmt.Errorf("http server error: %w", err)
	}

	return nil
}

func logLevel(lvl string) slog.Level {
	switch lvl {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	}

	return slog.LevelInfo
}
