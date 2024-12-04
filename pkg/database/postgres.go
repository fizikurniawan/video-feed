package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

type DatabaseManager struct {
	db *sql.DB
}

// NewDatabaseManager initializes the DatabaseManager with a database connection.
func NewDatabaseManager(dsn string) (*DatabaseManager, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	// Check the connection
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	return &DatabaseManager{db: db}, nil
}

// GetConnection retrieves the current DB connection.
func (dm *DatabaseManager) GetConnection() (*sql.DB, error) {
	if dm.db == nil || dm.db.Ping() != nil {
		// Attempt reconnect if DB connection is invalid
		err := dm.Reconnect()
		if err != nil {
			return nil, fmt.Errorf("failed to reconnect to database: %v", err)
		}
	}
	return dm.db, nil
}

// Reconnect attempts to reconnect to the database.
func (dm *DatabaseManager) Reconnect() error {
	log.Println("Reconnecting to the database...")
	// Implement the reconnect logic here (e.g., sleep, retry, etc.)
	var err error
	for i := 0; i < 3; i++ {
		dm.db, err = sql.Open("postgres", "your_connection_string_here") // Update with actual DSN
		if err == nil {
			err = dm.db.Ping()
			if err == nil {
				log.Println("Reconnected successfully.")
				return nil
			}
		}
		log.Printf("Reconnect attempt %d failed: %v", i+1, err)
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("failed to reconnect to the database after multiple attempts: %v", err)
}

// Query executes a SQL query and returns rows.
func (dm *DatabaseManager) Query(query string, args ...interface{}) (*sql.Rows, error) {
	log.Printf("Executing query: %s", query)
	db, err := dm.GetConnection()
	if err != nil {
		return nil, err
	}
	return db.Query(query, args...)
}

// Exec executes a SQL statement (INSERT, UPDATE, DELETE).
func (dm *DatabaseManager) Exec(query string, args ...interface{}) (sql.Result, error) {
	log.Printf("Executing query: %s", query)
	db, err := dm.GetConnection()
	if err != nil {
		return nil, err
	}
	return db.Exec(query, args...)
}

// BeginTransaction starts a new transaction.
func (dm *DatabaseManager) BeginTransaction() (*sql.Tx, error) {
	log.Println("Starting a new transaction...")
	db, err := dm.GetConnection()
	if err != nil {
		return nil, err
	}
	return db.Begin()
}

// CommitTransaction commits an ongoing transaction.
func (dm *DatabaseManager) CommitTransaction(tx *sql.Tx) error {
	log.Println("Committing transaction...")
	return tx.Commit()
}

// RollbackTransaction rolls back an ongoing transaction.
func (dm *DatabaseManager) RollbackTransaction(tx *sql.Tx) error {
	log.Println("Rolling back transaction...")
	return tx.Rollback()
}

// QueryRow executes a query and returns a single row.
func (dm *DatabaseManager) QueryRow(query string, args ...interface{}) *sql.Row {
	log.Printf("Executing query: %s", query)
	db, err := dm.GetConnection()
	if err != nil {
		return nil
	}
	return db.QueryRow(query, args...)
}

// Close closes the DB connection.
func (dm *DatabaseManager) Close() error {
	log.Println("Closing database connection...")
	return dm.db.Close()
}
