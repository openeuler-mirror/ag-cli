package pr

import "testing"

func TestBuildPRListPathIncludesLimit(t *testing.T) {
	got := buildPRListPath("openeuler", "ag-cli", "open", 5)
	want := "/repos/openeuler/ag-cli/pulls?state=open&per_page=5"

	if got != want {
		t.Fatalf("unexpected list path: got %q want %q", got, want)
	}
}
