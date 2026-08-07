package main

import (
	"log"

	"github.com/pressly/goose/v3"
)

func main() {
    err := goose.SetDialect("sqlite3")
    if err != nil {
        log.Fatal(err)
    }
}
