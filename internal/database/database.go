package database

import (
	"context"
	"fmt"
	"log"
	"time"
	"typetype/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(cfg *config.Config) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBSSLMode)
	if dsn == "" {
		log.Fatalln("No database url found.")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dbPool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf(
			"unable to create connection pool: %w",
			err,
		)
	}

	if err := dbPool.Ping(ctx); err != nil {
		dbPool.Close()

		return nil, fmt.Errorf(
			"database ping failed: %w",
			err,
		)
	}
	fmt.Println("Successfully connected to PostgreSQl with pgx1 ✅")
	return dbPool, err
}
func NewQueries(pool *pgxpool.Pool) *Queries {
	return New(pool) // New() is auto-provided by sqlc inside db.go
}
