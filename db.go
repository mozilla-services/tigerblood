package tigerblood

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

// DB is a DB instance for running queries against the tigerblood database
type DB struct {
	*sql.DB
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

	return &DB{
		DB: db,
	}, nil
}

const createReputationTableSQL = `
CREATE TABLE IF NOT EXISTS reputation (
ip ip4r PRIMARY KEY NOT NULL,
reputation int NOT NULL
);
`

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
