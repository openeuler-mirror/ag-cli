package issue

import "testing"

func TestBuildIssueListPathIncludesLimit(t *testing.T) {
	got := buildIssueListPath("openeuler", "ag-cli", "open", 5)
	want := "/repos/openeuler/ag-cli/issues?state=open&per_page=5"

	if got != want {
		t.Fatalf("unexpected issue list path: got %q want %q", got, want)
	}
}
