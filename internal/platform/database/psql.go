package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"hippo/internal/platform/config"
)

const (
	sslModeVerifyFull = "verify-full"
	sslModeVerifyCA   = "verify-ca"
	driverName        = "postgres"

	timeoutPing = 3 * time.Second
)

func NewPostgresConnection(conn config.DBConn) (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		conn.Host, conn.Port, conn.User, conn.Password, conn.Name, conn.SSLMode,
	)

	if conn.SSLMode == sslModeVerifyFull || conn.SSLMode == sslModeVerifyCA {
		if conn.SSLRootCert == "" {
			return nil, fmt.Errorf("sslrootcert is required for sslmode=%s", conn.SSLMode)
		}
		connStr += fmt.Sprintf(" sslrootcert=%s", conn.SSLRootCert)
	}

	db, err := sql.Open(driverName, connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open DB connection (host=%s): %w", conn.Host, err)
	}

	if conn.MaxOpenConns > 0 {
		db.SetMaxOpenConns(conn.MaxOpenConns)
	}
	if conn.MaxIdleConns > 0 {
		db.SetMaxIdleConns(conn.MaxIdleConns)
	}
	if conn.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(conn.ConnMaxLifetime)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeoutPing)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("db ping failed: %w", err)
	}

	return db, nil
}

//Things TODO
//Dynamic change SetMaxOpenConns & SetMaxIdleConns in accordance wth load
