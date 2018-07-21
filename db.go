package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

var db *sql.DB

func init() {
	conninfo := "user=postgres host=db dbname=friends_management sslmode=disable"
	dbconn, err := sql.Open("postgres", conninfo)
	if err != nil {
		log.Fatalf("error in db connection info %+v", err)
	}
	if err := dbconn.Ping(); err != nil {
		log.Fatal("error in pinging db %+v", err)
	}
	db = dbconn
}
