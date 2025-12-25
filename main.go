package main

import (
	"database/sql"
	"log"

	"github.com/codercollo/simple_bank/api"
	db "github.com/codercollo/simple_bank/db/sqlc"
	_ "github.com/lib/pq"
)

// Database & server configuraton
const (
	dbDriver      = "postgres"
	dbSource      = "postgres://root:secret@localhost:5433/simple_bank?sslmode=disable"
	serverAddress = "0.0.0.0:8080"
)

func main() {

	//Initialize databse connection
	conn, err := sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	//Initialize application dependecies
	store := db.NewStore(conn)
	server := api.NewServer(store)

	//Start HTTP server
	err = server.Start(serverAddress)
	if err != nil {
		log.Fatal("cannot start server")
	}

}
