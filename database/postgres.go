package database

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// InitPostgres mengembalikan connection pool, bukan error saja
func InitPostgres() (*pgxpool.Pool, error) {
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	user := os.Getenv("POSTGRES_USER")
	pass := os.Getenv("POSTGRES_PASSWORD")
	dbname := os.Getenv("POSTGRES_DB")

	if port == "" {
		port = "5432"
	}

	// Format DSN (Data Source Name)
	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", user, pass, host, port, dbname)

	// Context timeout untuk inisialisasi awal
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Parse config untuk validasi DSN
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("error parsing postgres config: %w", err)
	}

	// Konfigurasi Pool (Opsional, untuk performa)
	config.MaxConns = 20
	config.MinConns = 5

	// Buat Connection Pool
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Cek koneksi (Ping)
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	return pool, nil
}