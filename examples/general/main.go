package main

import (
	"database/sql"
	"flag"
	"fmt"

	_ "github.com/lib/pq"

	"github.com/rustedturnip/goscanql"
)

var (
	dbUser string
	dbPass string
	dbHost string
	dbName string
	dbPort int
)

func init() {
	flag.StringVar(&dbUser, "db-user", "postgres", "database user name")
	flag.StringVar(&dbPass, "db-pass", "postgres", "database user password")
	flag.StringVar(&dbHost, "db-host", "localhost", "database host")
	flag.StringVar(&dbName, "db-name", "goscanql", "database name")
	flag.IntVar(&dbPort, "db-port", 5432, "database port")
}

func main() {

	flag.Parse()

	db, err := sql.Open("postgres",
		fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPass, dbName))
	if err != nil {
		panic(err)
	}

	rows, err := db.Query(`SELECT account.id            AS id,
       account.email         AS email,
       account.date_of_birth AS date_of_birth,
       account.created_at    AS created_at,
       pet.id                AS pets_id,
       pet.name              AS pets_name,
       pet.animal            AS pets_animal,
       pet.breed             AS pets_breed,
       colour.name           AS pets_colour_name,
       colour.red            AS pets_colour_red,
       colour.green          AS pets_colour_green,
       colour.blue           AS pets_colour_blue
  FROM account
           LEFT JOIN pet ON account.id = pet.account_id
           LEFT JOIN colour ON pet.colour_name = colour.name;`)

	if err != nil {
		panic(err)
	}

	accounts, err := goscanql.RowsToStructs[*Account](rows)
	if err != nil {
		panic(err)
	}

	for _, account := range accounts {
		fmt.Printf("%#v\n", *account)
	}
}
