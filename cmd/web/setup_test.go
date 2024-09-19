package web

import (
	"github.com/spartanhooah/profile-picture-web/db/repository/dbrepo"
	"os"
	"testing"
)

var app Application

func TestMain(m *testing.M) {
	pathToTemplates = "./../../templates"

	app.Session = GetSession()

	app.DB = &dbrepo.TestDBRepo{}

	os.Exit(m.Run())
}
