package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/sunriseex/test_wallet/internal/config"
	"github.com/sunriseex/test_wallet/internal/logger"
)

func InitDB(cfg *config.Config) *sql.DB {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPass, cfg.DBName)
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		logger.Log.Fatalf("Error connect to DB: %v", err)
	}

	db.SetMaxOpenConns(100)
	db.SetConnMaxIdleTime(20)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		logger.Log.Fatalf("Error ping DB: %v", err)
	}

	logger.Log.Info("Successfully connected to database")
	return db

}

func InitSchema(db *sql.DB) {
	query := `
	CREATE TABLE IF NOT EXISTS wallet_db (
		wallet_id UUID PRIMARY KEY,
		balance NUMERIC NOT NULL DEFAULT 0,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);`

	_, err := db.Exec(query)
	if err != nil {
		logger.Log.Fatalf("Error creating table: %v", err)
	}
	logger.Log.Info("Successfully created wallet_db table")

}
