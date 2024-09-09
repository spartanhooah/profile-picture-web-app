package main

import (
	"github.com/spartanhooah/profile-picture-web/cmd/web"
	"log"
	"net/http"
)

func main() {
	// set up an app config
	app := web.Application{}

	app.Session = web.GetSession()

	// print out a message
	log.Println("Starting server on port 8080")

	// start the server
	err := http.ListenAndServe(":8080", app.Routes())

	if err != nil {
		log.Fatal(err)
	}
}
