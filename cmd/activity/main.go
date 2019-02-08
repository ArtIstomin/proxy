package main

import (
	"flag"
	"log"

	"github.com/artistomin/proxy/internal/app/activity/infrastructure/postgresql"
)

var (
	dsn = flag.String("dsn", "postgres://root:root@localhost:5432/activity?sslmode=disable", "Database URI")
)

func main() {
	flag.Parse()

	db, err := postgresql.New(*dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

}
