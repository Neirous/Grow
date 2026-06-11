package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func Init(dbPath string) error {
	var err error
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}

	// Single connection for SQLite to avoid BUSY errors
	DB.SetMaxOpenConns(1)

	// Enable WAL mode and foreign keys
	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA foreign_keys=ON",
		"PRAGMA busy_timeout=5000",
	}

	for _, p := range pragmas {
		if _, err := DB.Exec(p); err != nil {
			return fmt.Errorf("pragma %s: %w", p, err)
		}
	}

	// Run schema migration
	if _, err := DB.Exec(Schema); err != nil {
		return fmt.Errorf("run schema: %w", err)
	}

	log.Println("Database initialized successfully")
	return nil
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}
