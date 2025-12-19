package database

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string

	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

func LoadConfig() Config {
	maxOpen, _ := strconv.Atoi(os.Getenv("DB_MAX_OPEN_CONNS"))
	maxIdle, _ := strconv.Atoi(os.Getenv("DB_MAX_IDLE_CONNS"))
	lifetime, _ := time.ParseDuration(os.Getenv("DB_CONN_MAX_LIFETIME"))

	return Config{
		Host:            os.Getenv("DB_HOST"),
		Port:            os.Getenv("DB_PORT"),
		User:            os.Getenv("DB_USER"),
		Password:        os.Getenv("DB_PASSWORD"),
		DBName:          os.Getenv("DB_NAME"),
		SSLMode:         os.Getenv("DB_SSLMODE"),
		MaxOpenConns:    maxOpen,
		MaxIdleConns:    maxIdle,
		ConnMaxLifetime: lifetime,
	}
}

func Open(cfg Config, dbName string) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		dbName,
		cfg.SSLMode,
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	return db, db.Ping()
}
