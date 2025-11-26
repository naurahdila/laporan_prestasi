package database

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var PG *pgxpool.Pool

func InitPostgres() error {
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	user := os.Getenv("POSTGRES_USER")
	pass := os.Getenv("POSTGRES_PASSWORD")
	dbname := os.Getenv("POSTGRES_DB")
	if port == "" { port = "5432" }

	connStr := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", user, pass, host, port, dbname)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return err
	}
	// ping
	if err := pool.Ping(ctx); err != nil {
		return err
	}
	PG = pool
	return nil
}
