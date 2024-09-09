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

	// this needs to be different for the test for some reason
	pathToTemplates = "./../../templates"

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
	req, _ := http.NewRequest("GET", "/", nil)
	req = addContextAndSessionToRequest(req, app)
	resp := httptest.NewRecorder()
	// not sure what this does
	oldPath := pathToTemplates
	defer func() {
		pathToTemplates = oldPath
	}()
	pathToTemplates = "./../../templates/"

	// back to regular stuff
	handler := http.HandlerFunc(app.Home)

	handler.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Test case failed: expected %d, got %d", resp.Code, http.StatusOK)
	}

	body, _ := io.ReadAll(resp.Body)

	if !strings.Contains(string(body), "<small>From session:") {
		t.Error("Did not find correct text in HTML")
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
