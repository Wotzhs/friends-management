package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var db *sql.DB

func init() {
	dbname := "friends_management"

	if os.Getenv("GO_ENV") == "test" {
		dbname += "_test"
	}

	conninfo := "user=postgres host=db sslmode=disable dbname=" + dbname
	dbconn, err := sql.Open("postgres", conninfo)
	if err != nil {
		log.Fatalf("error in db connection info %+v", err)
	}
	if err := dbconn.Ping(); err != nil {
		log.Fatal("error in pinging db %+v", err)
	}
	db = dbconn
}
