package web

import (
	"github.com/spartanhooah/profile-picture-web/data"
	"html/template"
	"log"
	"net/http"
	"path"
	"time"
)

var pathToTemplates = "./templates/"

func (app *Application) Home(resp http.ResponseWriter, req *http.Request) {
	var td = make(map[string]any)

	if app.Session.Exists(req.Context(), "test") {
		msg := app.Session.GetString(req.Context(), "test")
		td["test"] = msg
	} else {
		app.Session.Put(req.Context(), "test", "Hit this page at "+time.Now().UTC().String())
	}

	_ = app.Render(resp, req, "home.page.gohtml", &TemplateData{Data: td})
}

func (app *Application) Profile(resp http.ResponseWriter, req *http.Request) {
	_ = app.Render(resp, req, "profile.page.gohtml", &TemplateData{})
}

type TemplateData struct {
	IP    string
	Data  map[string]any
	Error string
	Flash string
	User  data.User
}

func (app *Application) Render(resp http.ResponseWriter, req *http.Request, t string, td *TemplateData) error {
	parsedTemplate, err := template.ParseFiles(path.Join(pathToTemplates, t), path.Join(pathToTemplates, "base.layout.gohtml"))

	if err != nil {
		http.Error(resp, "bad request", http.StatusBadRequest)
		return err
	}

	td.IP = app.ipFromContext(req.Context())

	td.Error = app.Session.PopString(req.Context(), "error")
	td.Flash = app.Session.PopString(req.Context(), "flash")

	// execute the template, passing template data if any
	err = parsedTemplate.Execute(resp, td)

	if err != nil {
		return err
	}

	return nil
}

func (app *Application) Login(resp http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()

	if err != nil {
		log.Println(err)
		http.Error(resp, "bad request", http.StatusBadRequest)
	}

	// validate data
	form := NewForm(req.PostForm)
	form.Required("email", "password")

	if !form.Valid() {
		// redirect to login page
		app.Session.Put(req.Context(), "error", "Invalid login credentials")
		http.Redirect(resp, req, "/", http.StatusSeeOther)
		return
	}

	email := req.Form.Get("email")
	password := req.Form.Get("password")

	user, err := app.DB.GetUserByEmail(email)

	if err != nil {
		app.Session.Put(req.Context(), "error", "Invalid login!")
		http.Redirect(resp, req, "/", http.StatusSeeOther)
		return
	}

	if !app.authenticate(req, user, password) {
		app.Session.Put(req.Context(), "error", "Invalid login!")
		http.Redirect(resp, req, "/", http.StatusSeeOther)
	}

	// prevent fixation attack
	_ = app.Session.RenewToken(req.Context())

	app.Session.Put(req.Context(), "flash", "Successfully logged in")
	http.Redirect(resp, req, "/user/profile", http.StatusSeeOther)
}

func (app *Application) authenticate(req *http.Request, user *data.User, password string) bool {
	if valid, err := user.PasswordMatches(password); err != nil || !valid {
		return false
	}

	app.Session.Put(req.Context(), "user", user)

	return true
}
