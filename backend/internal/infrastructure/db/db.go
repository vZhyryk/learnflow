package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

func InitDatabase(dsn, maxIdleTime string, maxOpenConns, maxIdleConns int) (*sql.DB, error) {
	db, err := openDB(dsn, "pgx")
	if err != nil {
		return nil, fmt.Errorf("db: failed to open database: %w", err)
	}

	duration, er := time.ParseDuration(maxIdleTime)
	if er != nil {
		return nil, fmt.Errorf("db: failed to parse max idle time: %w", er)
	}
	db.SetConnMaxIdleTime(duration)

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)

	return db, nil
}

func openDB(dsn, dbType string) (*sql.DB, error) {
	db, err := sql.Open(dbType, dsn)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		return nil, err
	}

	return db, nil
}
