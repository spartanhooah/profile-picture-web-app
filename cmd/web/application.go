package web

import (
	"github.com/alexedwards/scs/v2"
	"github.com/spartanhooah/profile-picture-web/db/repository"
)

type Application struct {
	Session    *scs.SessionManager
	Datasource string
	DB         repository.DatabaseRepo
}
