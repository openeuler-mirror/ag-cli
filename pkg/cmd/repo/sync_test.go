package repo

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

type fakeRunner struct {
	outputs     map[string]string
	errors      map[string]error
	commands    []string
	interactive []string
}

func (f *fakeRunner) Run(name string, args ...string) (string, error) {
	key := commandKey(name, args...)
	f.commands = append(f.commands, key)
	if err, ok := f.errors[key]; ok {
		return "", err
	}
	return f.outputs[key], nil
}

func (f *fakeRunner) RunInteractive(name string, args ...string) error {
	key := commandKey(name, args...)
	f.interactive = append(f.interactive, key)
	if err, ok := f.errors[key]; ok {
		return err
	}
	return nil
}

func commandKey(name string, args ...string) string {
	return strings.TrimSpace(name + " " + strings.Join(args, " "))
}

func TestBuildAtomGitCloneURL(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "owner repo",
			in:   "openeuler/ag-cli",
			want: "https://atomgit.com/openeuler/ag-cli.git",
		},
		{
			name: "https url without suffix",
			in:   "https://atomgit.com/openeuler/ag-cli",
			want: "https://atomgit.com/openeuler/ag-cli.git",
		},
		{
			name: "https url with suffix",
			in:   "https://atomgit.com/openeuler/ag-cli.git",
			want: "https://atomgit.com/openeuler/ag-cli.git",
		},
		{
			name: "ssh url",
			in:   "git@atomgit.com:openeuler/ag-cli.git",
			want: "git@atomgit.com:openeuler/ag-cli.git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildAtomGitCloneURL(tt.in)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("buildAtomGitCloneURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildAtomGitCloneURLRejectsInvalidRepo(t *testing.T) {
	_, err := buildAtomGitCloneURL("ag-cli")
	if err == nil {
		t.Fatal("expected invalid repository format error")
	}
}

func TestValidateSyncOptionsRejectsUnsupportedStrategy(t *testing.T) {
	err := validateSyncOptions(&SyncOptions{
		Remote:         "origin",
		UpstreamRemote: "upstream",
		Strategy:       "squash",
	})
	if err == nil {
		t.Fatal("expected unsupported strategy error")
	}
}

func TestEnsureUpstreamRemoteAddsMissingRemote(t *testing.T) {
	r := &fakeRunner{
		errors: map[string]error{
			"git remote get-url upstream": fmt.Errorf("missing remote"),
		},
	}

	err := ensureUpstreamRemote(r, "upstream", "https://atomgit.com/openeuler/ag-cli.git", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{"git remote add upstream https://atomgit.com/openeuler/ag-cli.git"}
	if !reflect.DeepEqual(r.interactive, want) {
		t.Fatalf("interactive commands = %#v, want %#v", r.interactive, want)
	}
}

func TestEnsureUpstreamRemoteRequiresRepoWhenRemoteMissing(t *testing.T) {
	r := &fakeRunner{
		errors: map[string]error{
			"git remote get-url upstream": fmt.Errorf("missing remote"),
		},
	}

	err := ensureUpstreamRemote(r, "upstream", "", false)
	if err == nil {
		t.Fatal("expected missing upstream repository error")
	}
}

func TestRunSyncDryRunDoesNotMutateRepository(t *testing.T) {
	r := &fakeRunner{
		outputs: map[string]string{
			"git rev-parse --git-dir":     ".git",
			"git branch --show-current":   "master",
			"git status --porcelain":      "",
			"git remote get-url upstream": "https://atomgit.com/openeuler/ag-cli.git",
			"git log --oneline --decorate --max-count=20 HEAD..upstream/master": "abc123 update docs",
		},
	}

	err := runSync(r, &SyncOptions{
		Remote:         "origin",
		UpstreamRemote: "upstream",
		Strategy:       "ff-only",
		DryRun:         true,
	}, "openeuler/ag-cli")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(r.interactive) != 0 {
		t.Fatalf("dry-run should not run interactive commands: %#v", r.interactive)
	}
}

func TestRunSyncFetchesMergesAndPushes(t *testing.T) {
	r := &fakeRunner{
		outputs: map[string]string{
			"git rev-parse --git-dir":     ".git",
			"git status --porcelain":      "",
			"git remote get-url upstream": "https://atomgit.com/openeuler/ag-cli.git",
		},
	}

	err := runSync(r, &SyncOptions{
		Branch:         "develop",
		Remote:         "origin",
		UpstreamRemote: "upstream",
		Strategy:       "rebase",
		Push:           true,
		SetUpstream:    true,
	}, "openeuler/ag-cli")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{
		"git fetch upstream develop",
		"git rebase upstream/develop",
		"git push --set-upstream origin develop",
	}
	if !reflect.DeepEqual(r.interactive, want) {
		t.Fatalf("interactive commands = %#v, want %#v", r.interactive, want)
	}
}
