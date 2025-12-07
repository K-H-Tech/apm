package tools

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"

	"github.com/K-H-Tech/apm/internal/config"
)

// ConnectToPostgreSQL will initialize a PostgreSQL connection
func ConnectToPostgreSQL(pgCfg config.PostgreSQL) (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		pgCfg.Host,
		pgCfg.Port,
		pgCfg.Username,
		pgCfg.Password,
		pgCfg.DBName,
	)

	d, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Panicln(err)
	}

	d.SetMaxOpenConns(pgCfg.MaxOpenConnections)
	d.SetMaxIdleConns(pgCfg.MaxIdleConnections)
	d.SetConnMaxLifetime(pgCfg.ConnMaxLifetime)

	if err := d.Ping(); err != nil {
		return nil, err
	}

	return d, nil
}
