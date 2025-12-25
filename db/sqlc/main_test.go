package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

// Database configuraton
const (
	dbDriver = "postgres"
	dbSource = "postgres://root:secret@localhost:5433/simple_bank?sslmode=disable"
)

// Test database objects
var testQueries *Queries
var testDB *sql.DB

func TestMain(m *testing.M) {
	var err error

	//Open databse connection
	testDB, err = sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	//Initialize queries
	testQueries = New(testDB)

	//Run tests
	os.Exit(m.Run())
}
