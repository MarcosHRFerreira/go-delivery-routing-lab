// Package internalsql provides database connection helpers.
package internalsql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/config"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

const (
	dbPingTimeout     = 5 * time.Second
	dbMaxOpenConns    = 25
	dbMaxIdleConns    = 25
	dbConnMaxLifetime = 5 * time.Minute
	dbConnMaxIdleTime = 2 * time.Minute
)

func ConnectMySQL(cfg *config.Config) (*sql.DB, error) {
	dataSourceName := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=UTC&charset=utf8mb4&collation=utf8mb4_unicode_ci&timeout=5s&readTimeout=5s&writeTimeout=5s",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)

	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("open mysql connection: %w", err)
	}

	db.SetMaxOpenConns(dbMaxOpenConns)
	db.SetMaxIdleConns(dbMaxIdleConns)
	db.SetConnMaxLifetime(dbConnMaxLifetime)
	db.SetConnMaxIdleTime(dbConnMaxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), dbPingTimeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping mysql connection: %w", err)
	}

	return db, nil
}

func ConectMySql(cfg *config.Config) (*sql.DB, error) {
	return ConnectMySQL(cfg)
}
