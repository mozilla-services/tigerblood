package tigerblood

import (
	"database/sql"
	"fmt"
	"github.com/lib/pq"
)

type CheckViolationError struct {
	Inner *pq.Error
}

func (e CheckViolationError) Error() string {
	return e.Inner.Error()
}

const pgCheckViolationErrorCode = "23514"

var ErrNoRowsAffected = fmt.Errorf("No rows affected")

// DB is a DB instance for running queries against the tigerblood database
type DB struct {
	*sql.DB
	reputationSelectStmt *sql.Stmt
}

type ReputationEntry struct {
	IP         string
	Reputation uint
}

// NewDB creates a new DB instance from a DSN.
func NewDB(dsn string) (*DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(100)
	newDB := &DB{
		DB: db,
	}
	err = newDB.CreateTables()
	if err != nil {
		return nil, fmt.Errorf("Could not create tables: %s", err)
	}
	reputationSelectStmt, err := db.Prepare("SELECT ip, reputation FROM reputation WHERE ip >>= $1 ORDER BY @ ip LIMIT 1;")
	if err != nil {
		return nil, fmt.Errorf("Could not create prepared statement: %s", err)
	}
	newDB.reputationSelectStmt = reputationSelectStmt
	return newDB, nil
}

const createReputationTableSQL = `
CREATE TABLE IF NOT EXISTS reputation (
ip ip4r PRIMARY KEY NOT NULL,
reputation int NOT NULL CHECK (reputation >= 0 AND reputation <= 100)
);
CREATE INDEX IF NOT EXISTS reputation_ip_idx ON reputation USING gist (ip);
`

const emptyReputationTableSQL = `
TRUNCATE TABLE reputation;
`

// Close closes the database
func (db DB) Close() error {
	err := db.reputationSelectStmt.Close()
	if err != nil {
		return err
	}
	return db.DB.Close()
}

// CreateTables creates all the tables tigerblood needs, if they don't exist already
func (db DB) CreateTables() error {
	err := db.createReputationTable()
	if err != nil {
		return fmt.Errorf("Could not create reputation table: %s", err)
	}
	return nil
}

func (db DB) createReputationTable() error {
	_, err := db.Query(createReputationTableSQL)
	return err
}

func (db DB) emptyReputationTable() error {
	_, err := db.Query(emptyReputationTableSQL)
	return err
}

// InsertOrUpdateReputationEntry inserts a single ReputationEntry into the database, and if it already exists, it updates it
func (db DB) InsertOrUpdateReputationEntry(tx *sql.Tx, entry ReputationEntry) error {
	exec := db.Exec
	if tx != nil {
		exec = tx.Exec
	}
	_, err := exec("INSERT INTO reputation (ip, reputation) VALUES ($1, $2) ON CONFLICT (ip) DO UPDATE SET reputation = $2;", entry.IP, entry.Reputation)
	return err
}

// InsertReputationEntry inserts a single ReputationEntry into the database
func (db DB) InsertReputationEntry(tx *sql.Tx, entry ReputationEntry) error {
	exec := db.Exec
	if tx != nil {
		exec = tx.Exec
	}
	_, err := exec("INSERT INTO reputation (ip, reputation) VALUES ($1, $2);", entry.IP, entry.Reputation)
	if pqErr, ok := err.(*pq.Error); ok {
		if pqErr.Code == pgCheckViolationErrorCode {
			return CheckViolationError{pqErr}
		}
	}
	return err
}

// UpdateReputationEntry updated a single ReputationEntry on the database
func (db DB) UpdateReputationEntry(tx *sql.Tx, entry ReputationEntry) error {
	exec := db.Exec
	if tx != nil {
		exec = tx.Exec
	}
	result, err := exec("UPDATE reputation SET reputation = $2 WHERE ip = $1 RETURNING ip;", entry.IP, entry.Reputation)
	if pqErr, ok := err.(*pq.Error); ok {
		if pqErr.Code == pgCheckViolationErrorCode {
			return CheckViolationError{pqErr}
		}
	}
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNoRowsAffected
	}
	return nil
}

// SelectSmallestMatchingSubnet returns the smallest subnet in the database that contains the IP passed as a parameter.
func (db DB) SelectSmallestMatchingSubnet(ip string) (ReputationEntry, error) {
	var entry ReputationEntry
	err := db.reputationSelectStmt.QueryRow(ip).Scan(&entry.IP, &entry.Reputation)
	return entry, err
}

// DeleteReputationEntry deletes an entry from the database based on the entry's IP address
func (db DB) DeleteReputationEntry(tx *sql.Tx, entry ReputationEntry) error {
	exec := db.Exec
	if tx != nil {
		exec = tx.Exec
	}
	_, err := exec("DELETE FROM reputation WHERE ip = $1;", entry.IP)
	return err
}
