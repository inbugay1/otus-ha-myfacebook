package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/inbugay1/httprouter"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"myfacebook/internal/apiclient"
	"myfacebook/internal/apiv1/handler"
	apiv1middleware "myfacebook/internal/apiv1/middleware"
	"myfacebook/internal/config"
	"myfacebook/internal/connectionwatcher"
	"myfacebook/internal/db"
	"myfacebook/internal/httpclient"
	"myfacebook/internal/httphandler"
	httproutermiddleware "myfacebook/internal/httprouter/middleware"
	"myfacebook/internal/httpserver"
	internalapihandler "myfacebook/internal/internalapi/handler"
	internalapimiddleware "myfacebook/internal/internalapi/middleware"
	"myfacebook/internal/myfacebookdialogapiclient"
	"myfacebook/internal/postfanoutservice"
	"myfacebook/internal/postfeedcache"
	"myfacebook/internal/rdb"
	"myfacebook/internal/repository/rest"
	sqlxrepo "myfacebook/internal/repository/sqlx"
	"myfacebook/internal/rmq"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Application error: %s", err)
	}
}

func run() error {
	envConfig := config.GetConfigFromEnv()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel(envConfig.LogLevel),
	}))
	slog.SetDefault(logger)

	ctx := context.Background()

	tracerShutdown, err := initTracerProvider(ctx, envConfig)
	if err != nil {
		return fmt.Errorf("failed to init tracer provider: %w", err)
	}

	defer func() {
		if err := tracerShutdown(ctx); err != nil {
			log.Fatalf("Failed to shutdown TracerProvider: %s", err)
		}
	}()

	writeDB := db.New(db.Config{
		DriverName:         envConfig.WriteDBDriverName,
		Host:               envConfig.WriteDBHost,
		Port:               envConfig.WriteDBPort,
		Username:           envConfig.WriteDBUsername,
		Password:           envConfig.WriteDBPassword,
		DBName:             envConfig.WriteDBName,
		SSLMode:            envConfig.WriteDBSSLMode,
		MaxOpenConnections: envConfig.WriteDBMaxOpenConnections,
		MigrationPath:      "./storage/migrations",
	})

	if err := writeDB.Connect(ctx); err != nil {
		return fmt.Errorf("cannot connect to write db: %w", err)
	}

	slog.Info(fmt.Sprintf("Successfully connected to write db: %s", writeDB.GetDsn()))

	defer func() {
		if err := writeDB.Disconnect(); err != nil {
			log.Fatalf("Failed to disconnect from write db: %s", err)
		}
	}()

	if err := writeDB.Migrate(); err != nil {
		return fmt.Errorf("writeDB migration failed: %w", err)
	}

	readDB := db.New(db.Config{
		DriverName:         envConfig.ReadDBDriverName,
		Host:               envConfig.ReadDBHost,
		Port:               envConfig.ReadDBPort,
		Username:           envConfig.ReadDBUsername,
		Password:           envConfig.ReadDBPassword,
		DBName:             envConfig.ReadDBName,
		SSLMode:            envConfig.ReadDBSSLMode,
		MaxOpenConnections: envConfig.ReadDBMaxOpenConnections,
	})

	if err := readDB.Connect(ctx); err != nil {
		return fmt.Errorf("cannot connect to read db: %w", err)
	}

	slog.Info(fmt.Sprintf("Successfully connected to read db: %s", readDB.GetDsn()))

	defer func() {
		if err := readDB.Disconnect(); err != nil {
			log.Fatalf("Failed to disconnect from read db: %s", err)
		}
	}()

	rabbitMQ := rmq.New(&rmq.Config{
		Host:     envConfig.RMQHost,
		Port:     envConfig.RMQPort,
		Username: envConfig.RMQUsername,
		Password: envConfig.RMQPassword,
	}, []rmq.Exchange{
		{
			Name: "/post/feed/posted",
			Kind: "direct",
		},
	}, []rmq.Queue{
		{
			Name:    "/post/feed",
			Durable: true,
		},
	})

	if err := rabbitMQ.Connect(ctx); err != nil {
		return fmt.Errorf("cannot connect to rmq: %w", err)
	}

	slog.Info(fmt.Sprintf("Successfully connected to rmq on %s", net.JoinHostPort(envConfig.RMQHost, envConfig.RMQPort)))

	defer func() {
		if err := rabbitMQ.Disconnect(); err != nil {
			log.Fatalf("Failed to disconnect from rmq: %s", err)
		}
	}()

	connectionWatcher := connectionwatcher.New(&connectionwatcher.Config{
		PingInterval:     time.Duration(envConfig.ConnectionWatcherPingIntervalSeconds) * time.Second,
		PingTimeout:      time.Duration(envConfig.ConnectionWatcherPingTimeoutSeconds) * time.Second,
		ReconnectTimeout: time.Duration(envConfig.ConnectionWatcherReconnectTimeoutSeconds) * time.Second,
	})

	connectionWatcher.AddService("rmq", rabbitMQ)

	connectionWatcher.Start(ctx)
	defer connectionWatcher.Stop()

	httpClient := httpclient.New(&httpclient.Config{
		InsecureSkipVerify: true,
	})
	apiClient := apiclient.New(envConfig.MyfacebookDialogAPIBaseURL, httpClient)
	myfacebookDialogAPIClient := myfacebookdialogapiclient.New(apiClient)

	userRepository := sqlxrepo.NewUserRepository(writeDB, readDB)
	postRepository := sqlxrepo.NewPostRepository(writeDB, readDB)
	dialogRepository := rest.NewDialogRepository(myfacebookDialogAPIClient)

	redisDB := rdb.New(&rdb.Config{
		Host:     envConfig.RedisHost,
		Port:     envConfig.RedisPort,
		Password: envConfig.RedisPassword,
		DBNum:    envConfig.RedisDBNum,
	})

	if err := redisDB.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect to redis: %w", err)
	}

	defer func() {
		if err := redisDB.Disconnect(); err != nil {
			log.Fatalf("Failed to disconnect from redis: %s", err)
		}
	}()

	postFeedCache := postfeedcache.New(redisDB)

	postFanoutService := postfanoutservice.New(rabbitMQ, userRepository, postFeedCache, envConfig)

	err = postFanoutService.Start(ctx)
	if err != nil {
		return fmt.Errorf("failed to start post fanout service, %w", err)
	}

	defer postFanoutService.Stop()

	router := httprouter.New(httprouter.NewRegexRouteFactory())

	requestResponseMiddleware := httproutermiddleware.NewRequestResponseLog()

	apiv1ErrorResponseMiddleware := apiv1middleware.NewErrorResponse()
	apiv1ErrorLogMiddleware := apiv1middleware.NewErrorLog()
	apiv1AuthMiddleware := apiv1middleware.NewAuth(userRepository)

	router.Use(httproutermiddleware.NewRenameTraceRootSpan())
	router.Use(requestResponseMiddleware)

	router.Get("/health", &httphandler.Health{}, "")

	router.Group(func(router httprouter.Router) {
		router.Use(apiv1ErrorResponseMiddleware, apiv1ErrorLogMiddleware)

		router.Post("/user/register", &handler.Register{UserRepository: userRepository}, "")

		router.Get(`/user/{id:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}}`,
			&handler.GetUser{UserRepository: userRepository}, "/user/{id}")

		router.Post("/login", &handler.Login{
			UserRepository: userRepository,
		}, "")

		router.Get("/user/search", &handler.SearchUser{
			UserRepository: userRepository,
		}, "")

		router.Get(`/user/findByToken/{token:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}}`,
			&handler.FindUserByToken{UserRepository: userRepository}, "")

		router.Group(func(router httprouter.Router) {
			router.Use(apiv1AuthMiddleware)

			router.Put(`/friend/add/{id:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}}`, &handler.AddFriend{
				UserRepository: userRepository,
			}, "/friend/add/{id}")

			router.Put(`/friend/delete/{id:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}}`, &handler.DeleteFriend{
				UserRepository: userRepository,
				PostRepository: postRepository,
				PostFeedCache:  postFeedCache,
			}, "/friend/delete/{id}")

			router.Get("/post/get/{id:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}}", &handler.GetPost{
				PostRepository: postRepository,
			}, "/post/get/{id}")

			router.Post("/post/create", &handler.CreatePost{
				PostRepository: postRepository,
				RMQ:            rabbitMQ,
			}, "/post/create")

			router.Put("/post/update", &handler.UpdatePost{
				PostRepository: postRepository,
			}, "/post/update")

			router.Put("/post/delete/{id:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}}", &handler.DeletePost{
				PostRepository: postRepository,
				RMQ:            rabbitMQ,
			}, "/post/delete/{id}")

			router.Get("/post/feed", &handler.PostFeed{
				PostRepository: postRepository,
				UserRepository: userRepository,
				PostFeedCache:  postFeedCache,
				EnvConfig:      envConfig,
			}, "/post/feed")

			router.Post(`/dialog/{user_id:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}}/send`, &handler.SendDialog{
				DialogRepository: dialogRepository,
			}, "/dialog/{user_id}/send")

			router.Get(`/dialog/{user_id:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}}/list`, &handler.ListDialog{
				DialogRepository: dialogRepository,
			}, "/dialog/{user_id}/list")
		})
	})

	internalAPIErrorResponseMiddleware := internalapimiddleware.NewErrorResponse()
	internalAPIErrorLogMiddleware := internalapimiddleware.NewErrorLog()

	router.Group(func(router httprouter.Router) {
		router.WithPrefix("int") // add /int prefix to all group routes

		router.Use(internalAPIErrorResponseMiddleware, internalAPIErrorLogMiddleware)

		router.Get(`/user/findByToken/{token:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}}`,
			&internalapihandler.FindUserByToken{UserRepository: userRepository}, "/int/user/findByToken/{token}")

		router.Get(`/user/{id:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}}`,
			&internalapihandler.GetUser{UserRepository: userRepository}, "/int/user/{id}")
	})

	httpHandler := otelhttp.NewHandler(router, "")

	httpServer := httpserver.New(httpserver.Config{
		Port:                          envConfig.HTTPPort,
		RequestMaxHeaderBytes:         envConfig.RequestHeaderMaxSize,
		ReadHeaderTimeoutMilliseconds: envConfig.RequestReadHeaderTimeoutMilliseconds,
	}, httpHandler)

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

func initTracerProvider(ctx context.Context, envConfig *config.EnvConfig) (func(context.Context) error, error) {
	res, err := resource.Merge(resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL,
			semconv.ServiceName(envConfig.ServiceName),
			semconv.ServiceVersion(envConfig.Version),
		))
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	traceExporter, err := getTraceExporter(ctx, envConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create exporter: %w", err)
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tracerProvider)

	// set global propagator to tracecontext (the default is no-op).
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	// Shutdown will flush any remaining spans and shut down the exporter.
	return tracerProvider.Shutdown, nil
}

func getTraceExporter(ctx context.Context, envConfig *config.EnvConfig) (sdktrace.SpanExporter, error) { //nolint:ireturn
	switch envConfig.OTelExporterType { //nolint:gocritic
	case "otel_http":
		traceExporter, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpoint(envConfig.OTelExporterOTLPEndpoint), otlptracehttp.WithInsecure())
		if err != nil {
			return nil, fmt.Errorf("failed to create otlp exporter: %w", err)
		}

		return traceExporter, nil
	}

	traceExporter, err := stdouttrace.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout exporter: %w", err)
	}

	return traceExporter, nil
}
