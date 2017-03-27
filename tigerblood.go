package tigerblood

var db *DB = nil

func SetDB(newDb *DB) {
	db = newDb
}
