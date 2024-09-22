package web

import (
	"bytes"
	"context"
	"fmt"
	"github.com/spartanhooah/profile-picture-web/data"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"testing"
)

func Test_application_handlers(t *testing.T) {
	var tests = []struct {
		name                    string
		url                     string
		expectedStatusCode      int
		expectedUrl             string
		expectedFirstStatusCode int
	}{
		{"home", "/", http.StatusOK, "/", http.StatusOK},
		{"404", "/fish", http.StatusNotFound, "/fish", http.StatusNotFound},
		{"profile", "/user/profile", http.StatusOK, "/", http.StatusTemporaryRedirect},
	}

	routes := app.Routes()

	// create a test server
	ts := httptest.NewTLSServer(routes)
	defer ts.Close()

	client := ts.Client()

	for _, test := range tests {
		client.CheckRedirect = nil
		resp, err := ts.Client().Get(ts.URL + test.url)

		if err != nil {
			t.Log(err)
			t.Fatal(err)
		}

		if resp.StatusCode != test.expectedStatusCode {
			t.Errorf("Test case %s failed: expected status code %d, got %d", test.name, test.expectedStatusCode, resp.StatusCode)
		}

		if resp.Request.URL.Path != test.expectedUrl {
			t.Errorf("Test case %s failed: expected final URL of  %s, got %s", test.name, test.expectedUrl, resp.Request.URL.Path)
		}

		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}

		resp2, _ := client.Get(ts.URL + test.url)

		if resp2.StatusCode != test.expectedFirstStatusCode {
			t.Errorf("Test case %s failed: expected first status code %d, got %d", test.name, test.expectedFirstStatusCode, resp2.StatusCode)
		}
	}
}

func Test_application_Home(t *testing.T) {
	var tests = []struct {
		name         string
		putInSession string
		expectedHTML string
	}{
		{"first visit", "", "<small>From session:"},
		{"second visit", "hello, world!", "<small>From session: hello, world!"},
	}

	for _, test := range tests {
		req, _ := http.NewRequest("GET", "/", nil)
		req = addContextAndSessionToRequest(req, app)
		// clear out session
		_ = app.Session.Destroy(req.Context())

		if test.putInSession != "" {
			app.Session.Put(req.Context(), "test", test.putInSession)
		}

		resp := httptest.NewRecorder()

		// back to regular stuff
		handler := http.HandlerFunc(app.Home)

		handler.ServeHTTP(resp, req)

		if resp.Code != http.StatusOK {
			t.Errorf("Test case failed: expected %d, got %d", resp.Code, http.StatusOK)
		}

		body, _ := io.ReadAll(resp.Body)

		if !strings.Contains(string(body), test.expectedHTML) {
			t.Errorf("%s did not find %s in response body", test.name, test.expectedHTML)
		}
	}
}

func Test_Application_renderWithBadTemplate(t *testing.T) {
	// set template path to a location with a bad template
	pathToTemplates = "./testdata/"

	req, _ := http.NewRequest("GET", "/", nil)
	req = addContextAndSessionToRequest(req, app)
	resp := httptest.NewRecorder()

	err := app.Render(resp, req, "bad.page.gohtml", &TemplateData{})

	if err == nil {
		t.Error("Expected an error")
	}

	pathToTemplates = "./../../templates/"
}

func Test_Application_Login(t *testing.T) {
	var tests = []struct {
		name               string
		postedData         url.Values
		expectedStatusCode int
		expectedUrl        string
	}{
		{
			"valid login",
			url.Values{
				"email":    {"admin@example.com"},
				"password": {"secret"},
			},
			http.StatusSeeOther,
			"/user/profile",
		},
		{
			"missing form data",
			url.Values{
				"email":    {""},
				"password": {""},
			},
			http.StatusSeeOther,
			"/",
		},
		{
			"bad credentials",
			url.Values{
				"email":    {"admin@example.com"},
				"password": {"wrong"},
			},
			http.StatusSeeOther,
			"/",
		},
		{
			"user not found",
			url.Values{
				"email":    {"user@test.com"},
				"password": {"password"},
			},
			http.StatusSeeOther,
			"/",
		},
	}

	for _, test := range tests {
		req, _ := http.NewRequest("POST", "/login", strings.NewReader(test.postedData.Encode()))
		req = addContextAndSessionToRequest(req, app)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		resp := httptest.NewRecorder()
		handler := http.HandlerFunc(app.Login)

		handler.ServeHTTP(resp, req)

		if resp.Code != test.expectedStatusCode {
			t.Errorf("Test case %s failed: expected status %d, got %d", test.name, test.expectedStatusCode, resp.Code)
		}

		actualUrl, err := resp.Result().Location()

		if err == nil {
			if actualUrl.String() != test.expectedUrl {
				t.Errorf("Test case %s failed: expected url %s, got %s", test.name, test.expectedUrl, actualUrl.String())
			}
		} else {
			t.Errorf("%s: no location header set", test.name)
		}
	}
}

func Test_Application_auth(t *testing.T) {
	nextHandler := http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {

	})

	var tests = []struct {
		name   string
		isAuth bool
	}{
		{"logged in", true},
		{"not logged in", false},
	}

	for _, test := range tests {
		handlerToTest := app.auth(nextHandler)
		req := httptest.NewRequest(http.MethodGet, "http://testing", nil)
		req = addContextAndSessionToRequest(req, app)

		if test.isAuth {
			app.Session.Put(req.Context(), "user", data.User{ID: 1})
		}

		response := httptest.NewRecorder()
		handlerToTest.ServeHTTP(response, req)

		if test.isAuth && response.Code != http.StatusOK {
			t.Errorf("Test case %s failed: expected status %d, got %d", test.name, http.StatusOK, response.Code)
		}

		if !test.isAuth && response.Code != http.StatusTemporaryRedirect {
			t.Errorf("Test case %s failed: expected status %d, got %d", test.name, http.StatusTemporaryRedirect, response.Code)
		}
	}
}

func Test_Application_UploadFiles(t *testing.T) {
	// set up pipes
	pr, pw := io.Pipe()

	// create a new writer of type *io.Writer
	writer := multipart.NewWriter(pw)

	// create a waitgroup and add 1 to it; only necessary when doing a table test
	wg := &sync.WaitGroup{}
	wg.Add(1)

	// simulate uploading a file using a goroutine and the writer
	go simulatePNGUpload("./testdata/img.png", writer, t, wg)

	// read from the pipe which receives data
	request := httptest.NewRequest(http.MethodPost, "/", pr)
	request.Header.Add("Content-Type", writer.FormDataContentType())

	// call UploadFiles
	uploadedFiles, err := UploadFiles(request, "./testdata/uploads/")

	if err != nil {
		t.Error(err)
	}

	// assertions
	imagePath := fmt.Sprintf("./testdata/uploads/%s", uploadedFiles[0].OriginalFileName)
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		t.Errorf("Expected file to exist: %s", err.Error())
	}

	// cleanup
	_ = os.Remove(imagePath)

	wg.Wait()
}

func Test_Application_UploadProfilePicture(t *testing.T) {
	uploadPath = "./testdata/uploads"
	filePath := "./testdata/img.png"

	// specify a field name for the form
	fieldName := "file"

	// create a bytes.Buffer to act as the request body
	body := new(bytes.Buffer)

	// create a new writer
	writer := multipart.NewWriter(body)

	file, err := os.Open(filePath)

	if err != nil {
		t.Fatal(err)
	}

	w, err := writer.CreateFormFile(fieldName, filePath)

	if _, err := io.Copy(w, file); err != nil {
		t.Fatal(err)
	}

	writer.Close()

	request := httptest.NewRequest(http.MethodPost, "/", body)

	request = addContextAndSessionToRequest(request, app)

	app.Session.Put(request.Context(), "user", data.User{ID: 1})

	request.Header.Add("Content-Type", writer.FormDataContentType())

	response := httptest.NewRecorder()

	handler := http.HandlerFunc(app.UploadProfilePicture)

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusSeeOther {
		t.Errorf("Expected status %d, got %d", http.StatusSeeOther, response.Code)
	}

	_ = os.Remove(uploadPath + "/img.png")
}

func getCtx(req *http.Request) context.Context {
	return context.WithValue(req.Context(), contextUserKey, "unknown")
}

func addContextAndSessionToRequest(req *http.Request, app Application) *http.Request {
	req = req.WithContext(getCtx(req))

	ctx, _ := app.Session.Load(req.Context(), req.Header.Get("X-Session"))

	return req.WithContext(ctx)
}

func simulatePNGUpload(fileToUpload string, writer *multipart.Writer, t *testing.T, wg *sync.WaitGroup) {
	defer writer.Close()
	defer wg.Done()

	// create form data field "file" with filename being the value
	part, err := writer.CreateFormFile("file", path.Base(fileToUpload))

	if err != nil {
		t.Error(err)
	}

	// open the actual file
	f, err := os.Open(fileToUpload)

	if err != nil {
		t.Error(err)
	}

	defer f.Close()

	// decode the image
	img, _, err := image.Decode(f)

	if err != nil {
		t.Error("error decoding image:", err)
	}

	// write the image to the io.Writer
	err = png.Encode(part, img)

	if err != nil {
		t.Error(err)
	}
}
