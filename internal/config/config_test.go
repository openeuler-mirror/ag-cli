package config

import "testing"

func TestLoadTokenFromEnvPrefersAtomGitToken(t *testing.T) {
	t.Setenv("ATOMGIT_TOKEN", "atomgit-token")
	t.Setenv("GITCODE_TOKEN", "gitcode-token")

	if got := loadTokenFromEnv(); got != "atomgit-token" {
		t.Fatalf("expected ATOMGIT_TOKEN to win, got %q", got)
	}
}

func TestLoadTokenFromEnvFallsBackToGitCodeToken(t *testing.T) {
	t.Setenv("ATOMGIT_TOKEN", "")
	t.Setenv("GITCODE_TOKEN", "gitcode-token")

	if got := loadTokenFromEnv(); got != "gitcode-token" {
		t.Fatalf("expected GITCODE_TOKEN fallback, got %q", got)
	}
}

func TestLoadTokenFromEnvEmptyWhenUnset(t *testing.T) {
	t.Setenv("ATOMGIT_TOKEN", "")
	t.Setenv("GITCODE_TOKEN", "")

	if got := loadTokenFromEnv(); got != "" {
		t.Fatalf("expected empty token, got %q", got)
	}
}
