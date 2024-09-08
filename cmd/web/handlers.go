package web

import (
	"html/template"
	"net/http"
	"path"
)

var pathToTemplates = "./templates/"

func (app *Application) Home(response http.ResponseWriter, req *http.Request) {
	_ = app.Render(response, req, "home.page.gohtml", &TemplateData{})
}

type TemplateData struct {
	IP   string
	Data map[string]any
}

func (app *Application) Render(resp http.ResponseWriter, req *http.Request, t string, data *TemplateData) error {
	// parse the template from disk
	parsedTemplate, err := template.ParseFiles(path.Join(pathToTemplates, t))

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
