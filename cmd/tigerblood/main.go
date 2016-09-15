package main

import (
	"fmt"
	"go.mozilla.org/tigerblood"
	"log"
	"net/http"
	"os"
)

func main() {
	dsn, found := os.LookupEnv("TIGERBLOOD_DSN")
	if !found {
		log.Println("No database DSN found")
	}
	db, err := tigerblood.NewDB(dsn)
	if err != nil {
		panic(fmt.Errorf("Could not connect to the database: %s", err))
	}
	db.SetMaxOpenConns(80)
	var handler http.Handler = tigerblood.NewTigerbloodHandler(db)
	if _, found := os.LookupEnv("TIGERBLOOD_NO_HAWK"); !found {
		handler = tigerblood.NewHawkHandler(handler, nil)
	}
	http.HandleFunc("/", handler.ServeHTTP)
	http.ListenAndServe("127.0.0.1:8080", nil)
}
