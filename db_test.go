package tigerblood

import (
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
)

var testDB *DB

func TestMain(m *testing.M) {
	var err error
	testDB, err = NewDB("user=tigerblood dbname=tigerblood sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer testDB.Close()
	os.Exit(m.Run())
}

func TestCreateSchema(t *testing.T) {
	err := testDB.CreateTables()
	assert.Nil(t, err)
	err = testDB.CreateTables()
	assert.Nil(t, err, "Running CreateTables when the tables already exist shouldn't error")
}
