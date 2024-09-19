package main

import (
	"database/sql"
	"encoding/gob"
	"flag"
	"github.com/spartanhooah/profile-picture-web/cmd/web"
	"github.com/spartanhooah/profile-picture-web/data"
	"github.com/spartanhooah/profile-picture-web/db/repository/dbrepo"
	"log"
	"net/http"
)

func main() {
	gob.Register(data.User{})
	// set up an app config
	app := web.Application{}

	flag.StringVar(&app.Datasource, "datasource", "host=localhost port=5432 user=postgres password=postgres dbname=users sslmode=disable timezone=UTC connect_timeout=5", "Postgres connection")
	flag.Parse()

	conn, err := app.ConnectToDB()

	if err != nil {
		log.Fatal(err)
	}

	defer func(conn *sql.DB) {
		err := conn.Close()
		if err != nil {
			log.Fatal("Error closing connection", err)
		}
	}(conn)

	app.DB = &dbrepo.PostgresDBRepo{DB: conn}

	app.Session = web.GetSession()

	// print out a message
	log.Println("Starting server on port 8080")

	// start the server
	err = http.ListenAndServe(":8080", app.Routes())

	if err != nil {
		log.Fatal(err)
	}
}
