package db

import (
	"context"
	"database/sql" // standard Go interface for SQL dbs
	"fmt"
	"time"

	_ "modernc.org/sqlite" // sqlite driver
)

// Creating a scheduler table if it does not already exist
const scheme = `CREATE TABLE IF NOT EXISTS scheduler (   
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	date TEXT NOT NULL DEFAULT '',
	title TEXT NOT NULL DEFAULT '',
	comment TEXT NOT NULL DEFAULT '',
	repeat TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS scheduler_date ON scheduler(date);` // Creating an index on the date column

var db *sql.DB //Creating db connection descriptor

func Init(dbFile string) error {
	var err error

	db, err = sql.Open("sqlite", dbFile) //Preparing db handle
	if err != nil {
		return fmt.Errorf("failed to open db: %w", err)
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	cont, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err = db.PingContext(cont); err != nil { //verifing that db file can be read and written
		_ = db.Close()
		return fmt.Errorf("db connection failed: %w", err)
	}

	if _, err = db.ExecContext(cont, scheme); err != nil { //creating tables and indexes
		_ = db.Close()
		return fmt.Errorf("failed to initialize schema: %w", err)
	}

	return nil
}

// closing connections before program exit
func Close() error {
	if db != nil {
		return db.Close()
	}
	return nil
}
