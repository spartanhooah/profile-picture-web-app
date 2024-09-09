package web

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func Test_Form_Has(t *testing.T) {
	form := NewForm(nil)

	has := form.Has("something")

	if has {
		t.Error("Form says it has a field when it shouldn't")
	}

	postedData := url.Values{}
	postedData.Add("a", "a")

	form = NewForm(postedData)

	has = form.Has("a")

	if !has {
		t.Error("Form should have a field called 'a'")
	}
}

func Test_Form_Required(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	form := NewForm(req.PostForm)

	form.Required("a", "b", "c")

	if form.Valid() {
		t.Error("Form shows valid when required fields are missing")
	}

	postedData := url.Values{}
	postedData.Add("a", "a")
	postedData.Add("b", "b")
	postedData.Add("c", "c")

	req = httptest.NewRequest(http.MethodPost, "/", nil)
	req.PostForm = postedData

	form = NewForm(req.PostForm)

	form.Required("a", "b", "c")

	if !form.Valid() {
		t.Error("Form shows valid when required fields are missing")
	}
}

func Test_Form_Check(t *testing.T) {
	form := NewForm(nil)

	form.Check(false, "password", "password is required")

	if form.Valid() {
		t.Error("Form shows valid when field is required")
	}
}

func Test_Form_ErrorGet(t *testing.T) {
	form := NewForm(nil)

	form.Check(false, "password", "password is required")

	s := form.Errors.Get("password")

	if len(s) == 0 {
		t.Error("Should have an error returned from get but do not")
	}

	s = form.Errors.Get("something")

	if len(s) != 0 {
		t.Error("Should not have any error returned from get but got one")
	}
}
