package data

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverFindsConfigUpward(t *testing.T) {
	repo := newTestRepo(t)

	sub := filepath.Join(repo, "a", "b", "c")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatalf("mkdir sub: %v", err)
	}

	found, err := FindRepo(sub)
	if err != nil {
		t.Fatalf("FindRepo() error = %v", err)
	}
	if found != repo {
		t.Errorf("FindRepo() = %q, want %q", found, repo)
	}
}

func TestDiscoverErrorsWhenNoConfigFound(t *testing.T) {
	tmp := t.TempDir()

	if _, err := FindRepo(tmp); err == nil {
		t.Fatal("FindRepo() error = nil, want error")
	}
}

func TestDiscoverFindsLegacyConfigName(t *testing.T) {
	repo := t.TempDir()
	if err := os.WriteFile(filepath.Join(repo, ".beans.yml"), []byte("beans:\n    path: .beans\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got, err := FindRepo(repo); err != nil || got != repo {
		t.Fatalf("FindRepo() = (%q, %v), want %q", got, err, repo)
	}
}
