package tigerblood

import (
	"log"
	"database/sql"
	"fmt"
	"github.com/lib/pq"
	"time"
)

type CheckViolationError struct {
	Inner *pq.Error
}

func (e CheckViolationError) Error() string {
	return e.Inner.Error()
}

type DuplicateKeyError struct {
	Inner *pq.Error
}

func (e DuplicateKeyError) Error() string {
	return e.Inner.Error()
}

const pgDuplicateKeyErrorCode = "23505"
const pgCheckViolationErrorCode = "23514"

var ErrNoRowsAffected = fmt.Errorf("No rows affected")

// DB is a DB instance for running queries against the tigerblood database
type DB struct {
	*sql.DB
	reputationSelectStmt *sql.Stmt
	violationReputationWeightSelectStmt *sql.Stmt
}

type ReputationEntry struct {
	IP         string
	Reputation uint
}

type ViolationReputationWeightEntry struct {
	ViolationType string
	ReputationPenalty uint
}

func checkConnection(db *DB) {
	for {
		var one uint
		err := db.QueryRow("SELECT 1").Scan(&one)
		if err != nil {
			log.Fatal("Database connection failed:", err)
		}
		if one != 1 {
			log.Fatal("Apparently the database doesn't know the meaning of one anymore. Crashing.")
		}
		time.Sleep(10 * time.Second)
	}
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

	violationReputationWeightSelectStmt, err := db.Prepare("SELECT violation_type, reputation FROM violation_reputation_weights WHERE violation_type = $1 LIMIT 1;")
	if err != nil {
		return nil, fmt.Errorf("Could not create prepared statement: %s", err)
	}
	newDB.violationReputationWeightSelectStmt = violationReputationWeightSelectStmt

	// DB watchdog, crashes the process if connection dies
	go checkConnection(newDB)

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

const createViolationReputationWeightsTableSQL = `
CREATE TABLE IF NOT EXISTS violation_reputation_weights (
violation_type varchar(128) PRIMARY KEY NOT NULL,
reputation int NOT NULL CHECK (reputation >= 0 AND reputation <= 100),
UNIQUE (violation_type, reputation)
);
`

const emptyViolationReputationWeightsTableSQL = `
TRUNCATE TABLE violation_reputation_weights;
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
	err = db.createViolationReputationWeightsTable()
	if err != nil {
		return fmt.Errorf("Could not create violation reputation weights table: %s", err)
	}
	return nil
}

// EmptyTables truncates the tigerblood tables
func (db DB) EmptyTables() error {
	err := db.emptyReputationTable()
	if err != nil {
		return fmt.Errorf("Could not truncate reputation table: %s", err)
	}
	err = db.emptyViolationReputationWeightsTable()
	if err != nil {
		return fmt.Errorf("Could not truncate violation reputation weights table: %s", err)
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

func (db DB) createViolationReputationWeightsTable() error {
	_, err := db.Query(createViolationReputationWeightsTableSQL)
	return err
}

func (db DB) emptyViolationReputationWeightsTable() error {
	_, err := db.Query(emptyViolationReputationWeightsTableSQL)
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

// InsertOrUpdateReputationPenalty applies a reputationPenalty to the
// default reputation (100) and inserts a reputationEntry or updates
// an reputationEntry with the penalty
func (db DB) InsertOrUpdateReputationPenalty(tx *sql.Tx, ip string, reputationPenalty uint) error {
	exec := db.Exec
	if tx != nil {
		exec = tx.Exec
	}
	_, err := exec("INSERT INTO reputation (ip, reputation) VALUES ($1, 100 - $2) ON CONFLICT (ip) DO UPDATE SET reputation = GREATEST(0, LEAST(excluded.reputation, reputation.reputation - $2));", ip, reputationPenalty)
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
		switch pqErr.Code {
		case pgCheckViolationErrorCode:
			return CheckViolationError{pqErr}
		case pgDuplicateKeyErrorCode:
			return DuplicateKeyError{pqErr}
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

// inserts a single ViolationReputationWeightEntry into the database
func (db DB) InsertViolationReputationWeightEntry(tx *sql.Tx, entry ViolationReputationWeightEntry) error {
	exec := db.Exec
	if tx != nil {
		exec = tx.Exec
	}
	_, err := exec("INSERT INTO violation_reputation_weights (violation_type, reputation) VALUES ($1, $2);", entry.ViolationType, entry.ReputationPenalty)
	if pqErr, ok := err.(*pq.Error); ok {
		if pqErr.Code == pgCheckViolationErrorCode {
			return CheckViolationError{pqErr}
		}
	}
	return err
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

// unknown violations have no effect on reputation
const unknownViolationPenalty = 0

// Find the reputation to set an IP's reputation to for the given violation type
func (db DB) SelectViolationReputationWeightEntry(violationType string) (ViolationReputationWeightEntry, error) {
	var entry ViolationReputationWeightEntry
	err := db.violationReputationWeightSelectStmt.QueryRow(violationType).Scan(&entry.ViolationType, &entry.ReputationPenalty)
	if err == sql.ErrNoRows {
		log.Printf("Could not find reputation to set for violation type: %s", violationType)
		entry = ViolationReputationWeightEntry{ViolationType: "Unknown", ReputationPenalty: unknownViolationPenalty}
	}
	return entry, err
}

// Insert or update the IP's reputation after subtracting a penalty for the violation type
func (db DB) InsertOrUpdateReputationEntryByViolationType(tx *sql.Tx, ip string, violation_type string) error {
	violation_entry, err := db.SelectViolationReputationWeightEntry(violation_type)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("Error finding reputation for violation type: %s", err)
	}
	err = db.InsertOrUpdateReputationPenalty(nil, ip, violation_entry.ReputationPenalty)
	return err
}
