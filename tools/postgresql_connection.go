package tools

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"

	"github.com/K-H-Tech/apm/internal/config"
)

// ConnectToPostgreSQL will initialize a PostgreSQL connection
func ConnectToPostgreSQL(pgCfg config.PostgreSQL) (*sql.DB, error) {
	sslMode := pgCfg.SSLMode
	if sslMode == "" {
		sslMode = "require" // Secure default
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		pgCfg.Host,
		pgCfg.Port,
		pgCfg.Username,
		pgCfg.Password,
		pgCfg.DBName,
		sslMode,
	)

	d, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	d.SetMaxOpenConns(pgCfg.MaxOpenConnections)
	d.SetMaxIdleConns(pgCfg.MaxIdleConnections)
	d.SetConnMaxLifetime(pgCfg.ConnMaxLifetime)

	if err := d.Ping(); err != nil {
		return nil, err
	}

	return d, nil
}
