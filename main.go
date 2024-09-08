package main

import (
	"github.com/spartanhooah/profile-picture-web/cmd/web"
	"log"
	"net/http"
)

func main() {
	// set up an app config
	app := web.Application{}

	// get application routes
	mux := app.Routes()

	// print out a message
	log.Println("Starting server on port 8080")

	// start the server
	err := http.ListenAndServe(":8080", mux)

	if err != nil {
		log.Fatal(err)
	}
}
