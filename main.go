package main

import (
	"database/sql"
	"log"

	"github.com/codercollo/simple_bank/api"
	db "github.com/codercollo/simple_bank/db/sqlc"
	"github.com/codercollo/simple_bank/util"
	_ "github.com/lib/pq"
)

func main() {
	//Load config
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	//Initialize database connection
	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	//Initialize application dependecies
	store := db.NewStore(conn)
	server := api.NewServer(store)

	//Start HTTP server
	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("cannot start server")
	}

}
