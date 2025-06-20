package main

import (
	"database/sql"
	"log"
)

// checkDatabase attempts to connect and ping the configured database.
func checkDatabase() *UserError {
	db, err := sql.Open("mysql", "a4web:a4web@tcp(localhost:3306)/a4web?parseTime=true")
	if err != nil {
		return &UserError{Err: err, ErrorMessage: "failed to open database connection"}
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		return &UserError{Err: err, ErrorMessage: "failed to communicate with database"}
	}
	return nil
}

func performStartupChecks() {
	if ue := checkDatabase(); ue != nil {
		log.Fatalf("%s: %v", ue.ErrorMessage, ue.Err)
	}
}
