package web

import (
	"fmt"
	"github.com/spartanhooah/profile-picture-web/data"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"
)

var pathToTemplates = "./templates/"
var uploadPath = "./static/img"

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

	if app.Session.Exists(req.Context(), "user") {
		td.User = app.Session.Get(req.Context(), "user").(data.User)
	}

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

func (app *Application) UploadProfilePicture(resp http.ResponseWriter, req *http.Request) {
	// call a function to extract a file from an upload (request)
	files, err := UploadFiles(req, uploadPath)

	if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)

		return
	}

	// get the user from the session
	user := app.Session.Get(req.Context(), "user").(data.User)

	// create a var of type data.UserImage
	var img = data.UserImage{
		UserID:   user.ID,
		FileName: files[0].OriginalFileName,
	}

	// insert user image into user_images
	_, err = app.DB.InsertUserImage(img)

	if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)

		return
	}

	// refresh user in session
	updatedUser, err := app.DB.GetUser(user.ID)

	if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)

		return
	}

	app.Session.Put(req.Context(), "user", updatedUser)

	// redirect back to profile page
	http.Redirect(resp, req, "/user/profile", http.StatusSeeOther)
}

type UploadedFile struct {
	OriginalFileName string
	FileSize         int64
}

func UploadFiles(req *http.Request, uploadDirectory string) ([]*UploadedFile, error) {
	var uploadedFiles []*UploadedFile

	fiveMb := 1024 * 1024 * 5
	err := req.ParseMultipartForm(int64(fiveMb))

	if err != nil {
		return nil, fmt.Errorf("The uploaded file is too big; must be less than %d bytes", fiveMb)
	}

	for _, fHeaders := range req.MultipartForm.File {
		for _, header := range fHeaders {
			uploadedFiles, err = func(uploadedFiles []*UploadedFile) ([]*UploadedFile, error) {
				var uploadedFile UploadedFile
				infile, err := header.Open()

				if err != nil {
					return nil, err
				}

				defer infile.Close()

				uploadedFile.OriginalFileName = header.Filename

				var outfile *os.File

				if outfile, err = os.Create(filepath.Join(uploadDirectory, uploadedFile.OriginalFileName)); err != nil {
					return nil, err
				}

				defer outfile.Close()
				fileSize, err := io.Copy(outfile, infile)

				if err != nil {
					return nil, err
				}

				uploadedFile.FileSize = fileSize

				return append(uploadedFiles, &uploadedFile), nil
			}(uploadedFiles)

			if err != nil {
				return uploadedFiles, err
			}
		}
	}

	return uploadedFiles, nil
}
