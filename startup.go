package main

import (
	"database/sql"
	"fmt"
	"log"
)

var dbPool *sql.DB

func InitDB() *UserError {
	cfg := loadDBConfig()
	if cfg.User == "" {
		cfg.User = "a4web"
	}
	if cfg.Pass == "" {
		cfg.Pass = "a4web"
	}
	if cfg.Host == "" {
		cfg.Host = "localhost"
	}
	if cfg.Port == "" {
		cfg.Port = "3306"
	}
	if cfg.Name == "" {
		cfg.Name = "a4web"
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", cfg.User, cfg.Pass, cfg.Host, cfg.Port, cfg.Name)
	var err error
	dbPool, err = sql.Open("mysql", dsn)
	if err != nil {
		return &UserError{Err: err, ErrorMessage: "failed to open database connection"}
	}
	if err := dbPool.Ping(); err != nil {
		return &UserError{Err: err, ErrorMessage: "failed to communicate with database"}
	}
	return nil
}

// checkDatabase attempts to connect and ping the configured database.
func checkDatabase() *UserError {
	return InitDB()
}

func performStartupChecks() {
	if ue := checkDatabase(); ue != nil {
		log.Fatalf("%s: %v", ue.ErrorMessage, ue.Err)
	}
}
