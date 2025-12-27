package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/codercollo/simple_bank/util"
	_ "github.com/lib/pq"
)

// Test database objects
var testQueries *Queries
var testDB *sql.DB

func TestMain(m *testing.M) {
	var err error

	//Load config
	config, err := util.LoadConfig("../..")
	if err != nil {
		log.Fatal("cannot load config", err)
	}

	//Connect to test database
	testDB, err = sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	//Initialize queries
	testQueries = New(testDB)

	//Run tests
	os.Exit(m.Run())
}
