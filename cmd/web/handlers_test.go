package web

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func Test_application_handlers(t *testing.T) {
	var tests = []struct {
		name               string
		url                string
		expectedStatusCode int
	}{
		{"home", "/", http.StatusOK},
		{"404", "/fish", http.StatusNotFound},
	}

	routes := app.Routes()

	// create a test server
	ts := httptest.NewTLSServer(routes)
	defer ts.Close()

	for _, test := range tests {
		resp, err := ts.Client().Get(ts.URL + test.url)

		if err != nil {
			t.Log(err)
			t.Fatal(err)
		}

		if resp.StatusCode != test.expectedStatusCode {
			t.Errorf("Test case %s failed: expected %d, got %d", test.name, test.expectedStatusCode, resp.StatusCode)
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
}

func getCtx(req *http.Request) context.Context {
	return context.WithValue(req.Context(), contextUserKey, "unknown")
}

func addContextAndSessionToRequest(req *http.Request, app Application) *http.Request {
	req = req.WithContext(getCtx(req))

	ctx, _ := app.Session.Load(req.Context(), req.Header.Get("X-Session"))

	return req.WithContext(ctx)
}
