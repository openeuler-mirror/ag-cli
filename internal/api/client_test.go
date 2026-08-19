package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetRawReturnsBodyWithAuth(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v5/repos/openeuler/ag-cli/pulls/19/files" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer tok" {
			t.Errorf("unexpected auth header: %q", got)
		}
		w.Write([]byte(`{"raw":true}`))
	}))
	defer srv.Close()

	c := newClientWithBaseURL(srv.URL+"/api/v5", "tok")
	got, err := c.GetRaw("/repos/openeuler/ag-cli/pulls/19/files")
	if err != nil {
		t.Fatalf("GetRaw failed: %v", err)
	}
	if string(got) != `{"raw":true}` {
		t.Fatalf("unexpected body: %q", got)
	}
}

func TestGetRawErrorsOnNon200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := newClientWithBaseURL(srv.URL, "tok")
	if _, err := c.GetRaw("/anything"); err == nil || !strings.Contains(err.Error(), "500") {
		t.Fatalf("expected 500 error, got %v", err)
	}
}
