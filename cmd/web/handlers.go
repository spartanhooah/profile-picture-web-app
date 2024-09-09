package web

import (
	"fmt"
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

type TemplateData struct {
	IP   string
	Data map[string]any
}

func (app *Application) Render(resp http.ResponseWriter, req *http.Request, t string, data *TemplateData) error {
	// parse the template from disk
	parsedTemplate, err := template.ParseFiles(path.Join(pathToTemplates, t), path.Join(pathToTemplates, "base.layout.gohtml"))

	if err != nil {
		http.Error(resp, "bad request", http.StatusBadRequest)
		return err
	}

	data.IP = app.ipFromContext(req.Context())

	// execute the template, passing data if any
	err = parsedTemplate.Execute(resp, data)

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
		fmt.Fprintln(resp, "failed validation")
	}

	email := req.Form.Get("email")
	password := req.Form.Get("password")

	log.Println(email, password)

	fmt.Fprint(resp, email)
}
