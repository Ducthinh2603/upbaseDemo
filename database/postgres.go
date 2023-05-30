package database

import(
	"bufio"
	"database/sql"
	"os"
	"log"

	_ "github.com/lib/pq"
)

var PgDb *sql.DB
var err error

func init() {
	PgDb, err = sql.Open("postgres", GetDatabaseConfig())
	if err != nil {
		log.Fatal(err)
	}
	statement := ""
	f, err := os.Open("database/init.sql")
	if err != nil {
		log.Fatal("Can't open entry point: ", err)
	}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		statement += scanner.Text() + "\n"
	}
	_, err = PgDb.Exec(statement)
	if err != nil {
		log.Fatal("Can't initiate database: ", err)
	}
}