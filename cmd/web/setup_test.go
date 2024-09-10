package web

import (
	"os"
	"testing"
)

var app Application

func TestMain(m *testing.M) {
	pathToTemplates = "./../../templates"

	app.Session = GetSession()

	os.Exit(m.Run())
}
