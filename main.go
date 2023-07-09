package main

import (
	"database/sql"
	"log"

	"github.com/XiaozhouCui/go-bank/api"
	db "github.com/XiaozhouCui/go-bank/db/sqlc"
	"github.com/XiaozhouCui/go-bank/db/util"
	_ "github.com/lib/pq"
)

func main() {
	// load config from config file in the current path or from env variables
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}
	// connect to db
	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	store := db.NewStore(conn) // return a store interface
	server := api.NewServer(store)

	// start server
	err = server.Start(config.ServerAddress)

	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}
