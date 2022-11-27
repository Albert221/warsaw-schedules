package main

import (
	"log"

	"github.com/jmoiron/sqlx"
	"warsaw-schedules.dev/cmd"
	"warsaw-schedules.dev/db/sql"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	db, err := sqlx.Connect("mysql", "root:password@/warsaw_schedules")
	if err != nil {
		log.Fatalln(err)
	}

	stopRepo := sql.NewSqlStopRepository(db)

	cmd.Execute(stopRepo)
}
