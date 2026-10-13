package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// A user coming from standalone beans-tui has only ~/.config/beans-tui: it is
// read, its repo_beans_paths key lands in RepoStorePaths, the first save goes
// to ~/.config/lif/tui and keeps the map, and the legacy file stays untouched.
func TestLegacyConfigIsReadAndSaveMigrates(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	legacy := filepath.Join(home, ".config", "beans-tui", "config.yaml")
	legacyYAML := "repos:\n  - /repo/a\n  - /repo/b\nrepo_beans_paths:\n  /repo/a: /store/a\n"
	writeFile(t, legacy, legacyYAML)

	s, err := LoadSettings()
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"/repo/a", "/repo/b"}; !reflect.DeepEqual(s.Repos, want) {
		t.Errorf("Repos = %v, want %v", s.Repos, want)
	}
	if want := map[string]string{"/repo/a": "/store/a"}; !reflect.DeepEqual(s.RepoStorePaths, want) {
		t.Errorf("RepoStorePaths = %v, want %v", s.RepoStorePaths, want)
	}

	if err := SaveUserSettings([]string{"/repo/a", "/repo/b"}, "vi", "#f5a97f", 40); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(home, ".config", "lif", "tui", "config.yaml"))
	if err != nil {
		t.Fatalf("new config not written: %v", err)
	}
	reloaded, err := parseSettings(got)
	if err != nil {
		t.Fatal(err)
	}
	if want := map[string]string{"/repo/a": "/store/a"}; !reflect.DeepEqual(reloaded.RepoStorePaths, want) {
		t.Errorf("saved RepoStorePaths = %v, want %v", reloaded.RepoStorePaths, want)
	}
	if strings.Contains(string(got), "repo_beans_paths") {
		t.Errorf("saved file still carries the legacy key: %q", got)
	}
	after, err := os.ReadFile(legacy)
	if err != nil || string(after) != legacyYAML {
		t.Errorf("legacy file changed: %q, %v", after, err)
	}
}

func TestLegacyStateIsRead(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeFile(t, filepath.Join(home, ".config", "beans-tui", "state.json"), `{"last_repo":"/repo/legacy"}`)

	st, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if st.LastRepo != "/repo/legacy" {
		t.Errorf("LastRepo = %q, want /repo/legacy", st.LastRepo)
	}
}
