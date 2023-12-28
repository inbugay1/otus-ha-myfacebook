package rdb

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"github.com/redis/go-redis/v9"
)

type Config struct {
	Host     string
	Port     string
	Password string
	DBNum    int
}

type RedisDB struct {
	config *Config
	client *redis.Client
}

func New(config *Config) *RedisDB {
	return &RedisDB{
		config: config,
	}
}

func (rdb *RedisDB) Connect(ctx context.Context) error {
	rdb.client = redis.NewClient(&redis.Options{
		Addr:     net.JoinHostPort(rdb.config.Host, rdb.config.Port),
		Password: rdb.config.Password,
		DB:       rdb.config.DBNum,
	})

	if err := rdb.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to ping redis: %w", err)
	}

	slog.Info(fmt.Sprintf("Connected successfully to redis db num: %d on host: %s:%s", rdb.config.DBNum, rdb.config.Host, rdb.config.Port))

	return nil
}

func (rdb *RedisDB) Disconnect() error {
	err := rdb.client.Close()
	if err != nil {
		return fmt.Errorf("failed to close redis client connection: %w", err)
	}

	slog.Info("Successfully close redis client connection")

	return nil
}

func (rdb *RedisDB) GetClient() *redis.Client {
	return rdb.client
}
