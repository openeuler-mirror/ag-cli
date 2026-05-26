package repo

import "testing"

func TestParseRepoArgWithOwner(t *testing.T) {
	cloneURL, repoName := parseRepoArg("openeuler/ag-cli", "")
	if cloneURL != "https://atomgit.com/openeuler/ag-cli.git" {
		t.Fatalf("unexpected clone url: %s", cloneURL)
	}
	if repoName != "ag-cli" {
		t.Fatalf("unexpected repo name: %s", repoName)
	}
}

func TestParseRepoArgWithBareRepoName(t *testing.T) {
	cloneURL, repoName := parseRepoArg("ag-cli", "papertager")
	if cloneURL != "https://atomgit.com/papertager/ag-cli.git" {
		t.Fatalf("unexpected clone url: %s", cloneURL)
	}
	if repoName != "ag-cli" {
		t.Fatalf("unexpected repo name: %s", repoName)
	}
}
