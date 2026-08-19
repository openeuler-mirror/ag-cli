package release

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"atomgit.com/openeuler/ag-cli/internal/api"
)

func TestBuildReleaseListPath(t *testing.T) {
	if got := buildReleaseListPath("openeuler", "ag-cli"); got != "/repos/openeuler/ag-cli/releases" {
		t.Fatalf("unexpected list path: %s", got)
	}
}

func TestBuildReleaseTagPath(t *testing.T) {
	if got := buildReleaseTagPath("openeuler", "ag-cli", "v1.0.0"); got != "/repos/openeuler/ag-cli/releases/v1.0.0" {
		t.Fatalf("unexpected tag path: %s", got)
	}
}

func TestDownloadFileStreamsWithAuth(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer tok" {
			t.Errorf("unexpected auth header: %q", got)
		}
		w.Write([]byte("file-content"))
	}))
	defer srv.Close()

	var buf bytes.Buffer
	if err := downloadFile(srv.URL, "tok", &buf); err != nil {
		t.Fatalf("downloadFile failed: %v", err)
	}
	if buf.String() != "file-content" {
		t.Fatalf("unexpected content: %q", buf.String())
	}
}

func TestDownloadFileErrorsOnNon200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "forbidden", http.StatusForbidden)
	}))
	defer srv.Close()

	if err := downloadFile(srv.URL, "tok", &bytes.Buffer{}); err == nil {
		t.Fatal("expected error for 403")
	}
}

func TestValidateReleaseCreateOptionsRequiresTag(t *testing.T) {
	if err := validateReleaseCreateOptions("", "title", "body"); err == nil || err.Error() != "tag is required" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateReleaseCreateOptionsAcceptsMinimal(t *testing.T) {
	if err := validateReleaseCreateOptions("v1.0.0", "", ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRenderReleaseList(t *testing.T) {
	releases := []api.Release{
		{TagName: "v1.0.0", Name: "First", Prerelease: false},
		{TagName: "v0.9.0", Name: "RC", Prerelease: true},
	}

	var buf bytes.Buffer
	renderReleaseList(releases, &buf)

	want := "v1.0.0 - First\nv0.9.0 - RC (pre-release)\n"
	if buf.String() != want {
		t.Fatalf("unexpected output:\ngot:\n%s\nwant:\n%s", buf.String(), want)
	}
}
