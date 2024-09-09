package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_Application_AddIPToContext(t *testing.T) {
	tests := []struct {
		headerName   string
		headerValue  string
		addr         string
		expectedAddr string
		emptyAddr    bool
	}{
		{"", "", "", "unknown", false},
		{"", "", "", "unknown", true},
		{"X-Forwarded-For", "192.3.2.1", "", "192.3.2.1", false},
		{"", "", "hello:world", "unknown", false},
		{"", "", "127.0.0.1:8080", "127.0.0.1", false},
	}

	// create a dummy handler
	nextHandler := http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		// ensure the value exists in the context
		val := req.Context().Value(contextUserKey)

		if val == nil {
			t.Error(contextUserKey, "not present")
		}

		// ensure we get a string
		_, ok := val.(string)

		if !ok {
			t.Error("not a string")
		}
	})

	for _, test := range tests {
		// create the handler to test
		handlerToTest := app.AddIPToContext(nextHandler)

		// create a request and populate it with data from the table
		req := httptest.NewRequest(http.MethodGet, "http://testing", nil)

		if test.emptyAddr {
			req.RemoteAddr = ""
		}

		if len(test.headerName) > 0 {
			req.Header.Add(test.headerName, test.headerValue)
		}

		if len(test.addr) > 0 {
			req.RemoteAddr = test.addr
		}

		// execute the test
		handlerToTest.ServeHTTP(nil, req)

		actualAddr := req.Context().Value(contextUserKey)
		if test.expectedAddr != actualAddr {
			t.Errorf("Expected %s but got %s", test.expectedAddr, actualAddr)
		}
	}
}

func Test_Application_ipFromContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), contextUserKey, "127.0.0.1")

	ip := app.ipFromContext(ctx)

	if ip != "127.0.0.1" {
		t.Error("Wrong value from context")
	}
}
