package tigerblood

import (
	"database/sql"
	"fmt"
	"github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"go.mozilla.org/mozlogrus"
	"strings"
	"sync"
	"time"
)

func init() {
	mozlogrus.Enable("tigerblood")
}

// CheckViolationError postgres violation error
type CheckViolationError struct {
	Inner *pq.Error
}

func (e CheckViolationError) Error() string {
	return e.Inner.Error()
}

// DuplicateKeyError postgres duplicate key error
type DuplicateKeyError struct {
	Inner *pq.Error
}

func (e DuplicateKeyError) Error() string {
	return e.Inner.Error()
}

const pgDuplicateKeyErrorCode = "23505"
const pgCheckViolationErrorCode = "23514"

// ErrNoRowsAffected error to detect when an update doesn't occur
var ErrNoRowsAffected = fmt.Errorf("No rows affected")

// DB is a DB instance for running queries against the tigerblood database
type DB struct {
	*sql.DB
	reputationSelectStmt *sql.Stmt
	closeNotify          chan bool
	wait                 *sync.WaitGroup
}

// ReputationEntry an (IP, Reputation) entry
type ReputationEntry struct {
	IP         string // The IP address for the entry
	Reputation uint   // The reputation score
	Reviewed   bool   // True if the entry has the reviewed flag set
}

// IPViolationEntry an (IP, Violation) where Violation is the violation type name
type IPViolationEntry struct {
	IP        string
	Violation string
}

// ExceptionEntry describes an IP address exception
type ExceptionEntry struct {
	IP       string    // IP subnet exception applies to
	Creator  string    // Entity that created exception
	Modified time.Time // Entry modification date
	Expires  time.Time // Entry expiry date
}

func checkConnection(db *DB) {
	db.wait.Add(1)
	defer db.wait.Done()
	for {
		var one uint
		err := db.QueryRow("SELECT 1").Scan(&one)
		if err != nil {
			log.Fatal("Database connection failed:", err)
		}
		if one != 1 {
			log.Fatal("Apparently the database doesn't know the meaning of one anymore. Crashing.")
		}
		select {
		case <-db.closeNotify:
			return
		case <-time.After(10 * time.Second):
		}
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
	db.SetMaxOpenConns(75)
	db.SetMaxIdleConns(75)   // default is 2: https://golang.org/src/database/sql/sql.go#L652
	db.SetConnMaxLifetime(0) // don't timeout

	newDB := &DB{
		DB:          db,
		closeNotify: make(chan bool, 1),
		wait:        &sync.WaitGroup{},
	}
	err = newDB.CreateTables()
	if err != nil {
		return nil, fmt.Errorf("Could not create tables: %s", err)
	}
	reputationSelectStmt, err := db.Prepare("SELECT ip, reputation, reviewed FROM " +
		"reputation WHERE ip >>= $1 " +
		"AND NOT EXISTS (SELECT 1 FROM exception WHERE $1 <<= ip) " +
		"ORDER BY @ ip LIMIT 1;")
	if err != nil {
		return nil, fmt.Errorf("Could not create prepared statement: %s", err)
	}
	newDB.reputationSelectStmt = reputationSelectStmt

	// DB watchdog, crashes the process if connection dies
	go checkConnection(newDB)

	return newDB, nil
}

// Creates the reputation table (or modifies it to match the schema we want). Done in a few
// steps here to support migration of the schema from older to newer versions.
const createReputationTableSQL = `
CREATE TABLE IF NOT EXISTS reputation (
ip ip4r PRIMARY KEY NOT NULL,
reputation int NOT NULL CHECK (reputation >= 0 AND reputation <= 100)
);
CREATE INDEX IF NOT EXISTS reputation_ip_idx ON reputation USING gist (ip);

DO $$
	BEGIN
		ALTER TABLE reputation ADD COLUMN reviewed boolean DEFAULT false;
	EXCEPTION
		WHEN duplicate_column THEN -- ignore error
	END;
$$;

CREATE OR REPLACE FUNCTION reviewed_reset() RETURNS TRIGGER AS $$
	BEGIN
		NEW.reviewed = false;
		RETURN NEW;
	END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS check_reviewed ON reputation;
CREATE TRIGGER check_reviewed BEFORE UPDATE ON reputation
	FOR EACH ROW WHEN (NEW.reputation = 100)
	EXECUTE PROCEDURE reviewed_reset();
`

const createExceptionTableSQL = `
CREATE TABLE IF NOT EXISTS exception (
ip ip4r NOT NULL,
modified timestamp with time zone NOT NULL,
expires timestamp with time zone,
creator text NOT NULL,
UNIQUE(ip, creator)
);
CREATE INDEX IF NOT EXISTS exception_ip_idx ON exception USING gist (ip);
`

const emptyReputationTableSQL = `
TRUNCATE TABLE reputation;
`

const emptyExceptionTableSQL = `
TRUNCATE TABLE exception;
`

// Close closes the database
func (db DB) Close() error {
	db.closeNotify <- true
	db.wait.Wait()
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
	err = db.createExceptionTable()
	if err != nil {
		return fmt.Errorf("Could not create exception table: %s", err)
	}
	return nil
}

// EmptyTables truncates the tigerblood tables
func (db DB) EmptyTables() error {
	err := db.emptyReputationTable()
	if err != nil {
		return fmt.Errorf("Could not truncate reputation table: %s", err)
	}
	err = db.emptyExceptionTable()
	if err != nil {
		return fmt.Errorf("Could not truncate exception table: %s", err)
	}
	return nil
}

func (db DB) createReputationTable() error {
	_, err := db.Exec(createReputationTableSQL)
	return err
}

func (db DB) emptyReputationTable() error {
	_, err := db.Exec(emptyReputationTableSQL)
	return err
}

func (db DB) createExceptionTable() error {
	_, err := db.Exec(createExceptionTableSQL)
	return err
}

func (db DB) emptyExceptionTable() error {
	_, err := db.Exec(emptyExceptionTableSQL)
	return err
}

// InsertOrUpdateReputationEntry inserts a single ReputationEntry into the database, and if it already exists, it updates it
func (db DB) InsertOrUpdateReputationEntry(tx *sql.Tx, entry ReputationEntry) error {
	exec := db.Exec
	if tx != nil {
		exec = tx.Exec
	}
	_, err := exec("INSERT INTO reputation (ip, reputation, reviewed) "+
		"SELECT $1, $2, $3 WHERE NOT EXISTS (SELECT 1 FROM exception WHERE $1 <<= ip)"+
		"ON CONFLICT (ip) DO UPDATE SET reputation = $2, reviewed = $3;", entry.IP,
		entry.Reputation, entry.Reviewed)
	return err
}

// InsertOrUpdateReputationPenalties applies a reputationPenalty to the
// default reputation (100) and inserts a reputationEntry or updates
// an reputationEntry with the penalty
func (db DB) InsertOrUpdateReputationPenalties(tx *sql.Tx, ips []string, reputationPenalties []uint) error {
	exec := db.Exec
	if tx != nil {
		exec = tx.Exec
	}

	vals := []interface{}{}
	index := 0

	sqlStr := "WITH x AS (SELECT * FROM unnest(array["

	for i, ip := range ips {
		sqlStr += fmt.Sprintf("$%d,", i+1)
		vals = append(vals, ip)
		index = i + 1
	}
	sqlStr = strings.TrimSuffix(sqlStr, ",")

	sqlStr += "]::ip4r[], array["
	for i, p := range reputationPenalties {
		sqlStr += fmt.Sprintf("100 - $%d,", index+i+1)
		vals = append(vals, p)
	}
	sqlStr = strings.TrimSuffix(sqlStr, ",")
	sqlStr += "]) AS x(t0, t1) WHERE NOT EXISTS (SELECT 1 FROM exception WHERE t0 <<= ip)) " +
		"INSERT INTO reputation (ip, reputation) SELECT x.t0, x.t1 FROM x " +
		"ON CONFLICT (ip) DO UPDATE SET reputation = " +
		"GREATEST(0, LEAST(excluded.reputation, reputation.reputation - (100 - excluded.reputation)));"

	log.Debugf("sql: %s %s", sqlStr, vals)
	_, err := exec(sqlStr, vals...)
	return err
}

// InsertReputationEntry inserts a single ReputationEntry into the database
func (db DB) InsertReputationEntry(tx *sql.Tx, entry ReputationEntry) error {
	exec := db.Exec
	if tx != nil {
		exec = tx.Exec
	}
	_, err := exec("INSERT INTO reputation (ip, reputation, reviewed) "+
		"SELECT $1, $2, $3 WHERE NOT EXISTS (SELECT 1 FROM exception WHERE $1 <<= ip);",
		entry.IP, entry.Reputation, entry.Reviewed)

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

// UpdateReputationEntry updates a single ReputationEntry on the database
func (db DB) UpdateReputationEntry(tx *sql.Tx, entry ReputationEntry) error {
	exec := db.Exec
	if tx != nil {
		exec = tx.Exec
	}
	result, err := exec("UPDATE reputation SET reputation = $2, reviewed = $3 WHERE ip = $1 RETURNING ip;",
		entry.IP, entry.Reputation, entry.Reviewed)
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
	err := db.reputationSelectStmt.QueryRow(ip).Scan(&entry.IP, &entry.Reputation, &entry.Reviewed)
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

// InsertOrUpdateExceptionEntry inserts a single ExceptionEntry into the database, and if it already exists, it updates it
func (db DB) InsertOrUpdateExceptionEntry(tx *sql.Tx, entry ExceptionEntry) error {
	exec := db.Exec
	if tx != nil {
		exec = tx.Exec
	}
	var nt pq.NullTime
	// If the entry has no expiry set, insert it as a NULL (no expiry)
	if !entry.Expires.IsZero() {
		nt.Valid = true
		nt.Time = entry.Expires
	}
	_, err := exec("INSERT INTO exception (ip, modified, expires, creator) VALUES ($1, now(), $2, $3) "+
		"ON CONFLICT (ip, creator) DO UPDATE SET expires = $2, modified = now();",
		entry.IP, nt, entry.Creator)
	return err
}

// DeleteExceptionCreatorType removes all exceptions in the exception table that have been created
// by a specific exception source type. For example, if creatorType is file then the function will
// remove any exception created by any file based exception source.
func (db DB) DeleteExceptionCreatorType(tx *sql.Tx, creatorType string) error {
	exec := db.Exec
	if tx != nil {
		exec = tx.Exec
	}
	creatorType = creatorType + "%"
	_, err := exec("DELETE FROM exception WHERE creator LIKE $1", creatorType)
	return err
}

// DeleteExpiredExceptions removes any exception from the exception table that has expired
func (db DB) DeleteExpiredExceptions(tx *sql.Tx) error {
	exec := db.Exec
	if tx != nil {
		exec = tx.Exec
	}
	_, err := exec("DELETE FROM exception WHERE expires IS NOT NULL AND expires < now();")
	return err
}

// SelectExceptionsContaining returns any exceptions that apply to IP, or an empty slice if
// none were found.
func (db DB) SelectExceptionsContaining(ip string) (ret []ExceptionEntry, err error) {
	rows, err := db.Query("SELECT ip, modified, expires, creator FROM exception "+
		"WHERE $1 <<= ip", ip)
	if err != nil {
		return
	}
	for rows.Next() {
		var (
			nt  pq.NullTime
			ent ExceptionEntry
		)
		err = rows.Scan(&ent.IP, &ent.Modified, &nt, &ent.Creator)
		if err != nil {
			rows.Close()
			return
		}
		if nt.Valid {
			ent.Expires = nt.Time
		}
		ret = append(ret, ent)
	}
	err = rows.Err()
	return
}

// SelectExceptionsContainedBy returns any exceptions contained within subnet
func (db DB) SelectExceptionsContainedBy(subnet string) (ret []ExceptionEntry, err error) {
	rows, err := db.Query("SELECT ip, modified, expires, creator FROM exception "+
		"WHERE (expires > now() OR expires IS NULL) AND $1 >>= ip", subnet)
	if err != nil {
		return
	}
	for rows.Next() {
		var (
			nt  pq.NullTime
			ent ExceptionEntry
		)
		err = rows.Scan(&ent.IP, &ent.Modified, &nt, &ent.Creator)
		if err != nil {
			rows.Close()
			return
		}
		if nt.Valid {
			ent.Expires = nt.Time
		}
		ret = append(ret, ent)
	}
	err = rows.Err()
	return
}

// SelectAllExceptions returns all active exceptions
func (db DB) SelectAllExceptions() (ret []ExceptionEntry, err error) {
	return db.SelectExceptionsContainedBy("0.0.0.0/0")
}

// SetReviewedFlag sets the reviewed boolean flag on a reputation entry in the database
func (db DB) SetReviewedFlag(tx *sql.Tx, entry ReputationEntry, f bool) error {
	exec := db.Exec
	if tx != nil {
		exec = tx.Exec
	}
	res, err := exec("UPDATE reputation SET reviewed = $1 WHERE ip = $2;", f, entry.IP)
	if err != nil {
		return err
	}
	c, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if c == 0 {
		return fmt.Errorf("Request affected 0 reputation entries")
	}
	return nil
}
