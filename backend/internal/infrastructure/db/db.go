package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// InitDatabase opens a PostgreSQL connection and configures pool settings.
func InitDatabase(dsn, maxIdleTime, maxLifetime string, maxOpenConns, maxIdleConns int) (*sql.DB, error) {
	db, err := openDB(dsn, "postgres")
	if err != nil {
		return nil, fmt.Errorf("db: failed to open database: %w", err)
	}

	duration, er := time.ParseDuration(maxIdleTime)
	if er != nil {
		return nil, fmt.Errorf("db: failed to parse max idle time: %w", er)
	}
	db.SetConnMaxIdleTime(duration)

	lifetime, err := time.ParseDuration(maxLifetime)
	if err != nil {
		return nil, fmt.Errorf("db: failed to parse max lifetime: %w", err)
	}

	db.SetConnMaxLifetime(lifetime)
	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(lifetime)

	return db, nil
}

func openDB(dsn, dbType string) (*sql.DB, error) {
	db, err := sql.Open(dbType, dsn)
	if err != nil {
		return nil, fmt.Errorf("db: failed to open database: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("db: failed to ping database: %w", err)
	}

	return db, nil
}
