package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/sunriseex/test_wallet/internal/config"
)

func InitDB(cfg *config.Config) *sql.DB {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPass, cfg.DBName)
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatalf("Error connect to DB: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("Error ping DB: %v", err)
	}
	log.Println("Successfully connected to database")
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
		log.Fatalf("Error creating table: %v", err)
	}
	log.Println("Successfully created wallet_db table")

}
