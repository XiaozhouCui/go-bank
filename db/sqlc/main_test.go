package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/XiaozhouCui/go-bank/db/util"
	_ "github.com/lib/pq"
)

var testQueries *Queries

var testDB *sql.DB

// TestMain is the entry point for all tests.
func TestMain(m *testing.M) {
	config, err := util.LoadConfig("../..")
	if err != nil {
		log.Fatal("cannot load fonfig:", err)
	}
	testDB, err = sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	testQueries = New(testDB)

	os.Exit(m.Run())
}
