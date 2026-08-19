package pr

import (
	"bytes"
	"testing"
)

func TestBuildPRListPathIncludesLimit(t *testing.T) {
	got := buildPRListPath("openeuler", "ag-cli", "open", 5)
	want := "/repos/openeuler/ag-cli/pulls?state=open&per_page=5"

	if got != want {
		t.Fatalf("unexpected list path: got %q want %q", got, want)
	}
}

func TestBuildPRFilesPath(t *testing.T) {
	got := buildPRFilesPath("openeuler", "ag-cli", "19")
	want := "/repos/openeuler/ag-cli/pulls/19/files"

	if got != want {
		t.Fatalf("unexpected files path: got %q want %q", got, want)
	}
}

func TestRenderPRDiff(t *testing.T) {
	body := []byte(`[` +
		`{"filename":"pkg/cmd/pr/pr.go","additions":1,"deletions":1,"patch":{"diff":"@@ -1 +1 @@\n-a\n+b\n"}},` +
		`{"filename":"new.txt","additions":1,"deletions":0,"patch":{"diff":"@@ -0,0 +1 @@\n+hello"}}` +
		`]`)

	var buf bytes.Buffer
	if err := renderPRDiff(body, &buf); err != nil {
		t.Fatalf("renderPRDiff failed: %v", err)
	}

	want := "diff --git a/pkg/cmd/pr/pr.go b/pkg/cmd/pr/pr.go\n" +
		"--- a/pkg/cmd/pr/pr.go\n" +
		"+++ b/pkg/cmd/pr/pr.go\n" +
		"@@ -1 +1 @@\n-a\n+b\n" +
		"\n" +
		"diff --git a/new.txt b/new.txt\n" +
		"--- a/new.txt\n" +
		"+++ b/new.txt\n" +
		"@@ -0,0 +1 @@\n+hello\n"

	if buf.String() != want {
		t.Fatalf("unexpected diff output:\ngot:\n%s\nwant:\n%s", buf.String(), want)
	}
}

func TestRenderPRDiffUsesPatchPathsWhenPresent(t *testing.T) {
	body := []byte(`[{"filename":"new_name.txt","additions":0,"deletions":0,` +
		`"patch":{"old_path":"old_name.txt","new_path":"new_name.txt","diff":"@@ -1 +0,0 @@\n-gone\n"}}]`)

	var buf bytes.Buffer
	if err := renderPRDiff(body, &buf); err != nil {
		t.Fatalf("renderPRDiff failed: %v", err)
	}

	want := "diff --git a/old_name.txt b/new_name.txt\n" +
		"--- a/old_name.txt\n" +
		"+++ b/new_name.txt\n" +
		"@@ -1 +0,0 @@\n-gone\n"

	if buf.String() != want {
		t.Fatalf("unexpected diff output:\ngot:\n%s\nwant:\n%s", buf.String(), want)
	}
}

func TestRenderPRDiffRejectsNonArray(t *testing.T) {
	if err := renderPRDiff([]byte(`{"code":0,"diffs":[]}`), &bytes.Buffer{}); err == nil {
		t.Fatal("expected error for legacy object payload")
	}
}

func TestNormalizePRHeadKeepsRepoQualifiedHead(t *testing.T) {
	got := normalizePRHead("ag-cli", "papertager/ag-cli:feature/x")
	want := "papertager/ag-cli:feature/x"

	if got != want {
		t.Fatalf("unexpected normalized head: got %q want %q", got, want)
	}
}

func TestNormalizePRHeadExpandsOwnerBranchWithSlash(t *testing.T) {
	got := normalizePRHead("ag-cli", "papertager:feature/x")
	want := "papertager/ag-cli:feature/x"

	if got != want {
		t.Fatalf("unexpected normalized head: got %q want %q", got, want)
	}
}

func TestValidatePRCreateOptionsRequiresTitle(t *testing.T) {
	err := validatePRCreateOptions("", "papertager:feature")
	if err == nil || err.Error() != "title is required" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidatePRCreateOptionsRequiresHead(t *testing.T) {
	err := validatePRCreateOptions("test", "")
	if err == nil || err.Error() != "head is required" {
		t.Fatalf("unexpected error: %v", err)
	}
}
