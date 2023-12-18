package db

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file" // enable file migrations
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // enable postgres driver
)

type Config struct {
	DriverName         string
	Host               string
	Port               int
	User               string
	Password           string
	DBName             string
	SSLMode            string
	MigrationPath      string
	MaxOpenConnections int
}

type DB struct {
	config Config
	conn   *sqlx.DB
}

func New(config Config) *DB {
	return &DB{
		config: config,
	}
}

func (db *DB) Connect(ctx context.Context) error {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		db.config.Host, db.config.Port, db.config.User, db.config.Password, db.config.DBName, db.config.SSLMode)

	conn, err := sqlx.ConnectContext(ctx, db.config.DriverName, dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database %q on %s:%d: %w", db.config.DBName, db.config.Host, db.config.Port, err)
	}

	conn.SetMaxOpenConns(db.config.MaxOpenConnections)

	if err := conn.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping postgres on %s:%d: %w", db.config.Host, db.config.Port, err)
	}

	db.conn = conn

	return nil
}

func (db *DB) Migrate() error {
	driver, err := postgres.WithInstance(db.conn.DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration postgres driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://"+db.config.MigrationPath, db.config.DBName, driver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err = m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			slog.Info("No new db migrations to run")

			return nil
		}

		return fmt.Errorf("failed to up migration: %w", err)
	}

	slog.Info("Successfully run db migrations")

	return nil
}

func (db *DB) Disconnect() error {
	if err := db.conn.Close(); err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}

	return nil
}

func (db *DB) GetConnection() *sqlx.DB {
	return db.conn
}
