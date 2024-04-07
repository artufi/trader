package database

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	_ "github.com/jackc/pgx/v5/stdlib"
	"time"
)

func MustOpen(db *sql.DB, err error) *sql.DB {
	if err != nil {
		panic(err)
	}
	return db
}

// OpenPool will open a SQL Pool connection with the provided Postgres database.
// *sql.DB should be closed
func OpenPool(ctx context.Context, config PostgresConfig) (*sql.DB, error) {
	pool, err := pgxpool.New(ctx, config.String())
	if err != nil {
		return nil, fmt.Errorf("postgres open pool: %w", err)
	}
	return stdlib.OpenDBFromPool(pool), nil
}

func Connect(db *sql.DB, maxRetries int, delay time.Duration) error {
	var err error
	for i := 1; i <= maxRetries; i++ {
		err = db.Ping()
		if err == nil {
			return nil
		}
		time.Sleep(time.Second * delay)
	}
	return fmt.Errorf("postgres connection failed attempts=%d: %w", maxRetries, err)
}

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
}

func (pc PostgresConfig) String() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		pc.Host, pc.Port, pc.User, pc.Password, pc.Database, pc.SSLMode)
}
