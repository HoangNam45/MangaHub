package database

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func InitDB() {
	var err error

	DB, err = sql.Open("sqlite", "./data/mangahub.db")
	if err != nil {
		log.Fatal("Cannot open DB:", err)
	}

	err = DB.Ping()
	if err != nil {
		log.Fatal("Cannot connect DB:", err)
	}

	log.Println("Connected to SQLite (modernc)")
}