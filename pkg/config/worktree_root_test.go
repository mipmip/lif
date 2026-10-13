package config

import (
	"os"
	"path/filepath"
	"testing"
)

// worktree root (R-10): without worktree.path the home root is the legacy
// ~/.beans/worktrees/<project> when that directory exists, else
// ~/.lif/worktrees/<project>. No ambiguity error: the legacy root wins when
// both exist. Every test goes through the production resolver, with HOME set
// to a temp directory.

func homeWithRoots(t *testing.T, project string, legacy, current bool) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	if legacy {
		if err := os.MkdirAll(filepath.Join(home, ".beans", "worktrees", project), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if current {
		if err := os.MkdirAll(filepath.Join(home, ".lif", "worktrees", project), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	return home
}

func TestResolveWorktreePath(t *testing.T) {
	t.Run("legacy root present is used and reported", func(t *testing.T) {
		home := homeWithRoots(t, "proj", true, false)
		got, legacy, err := (&Config{}).ResolveWorktreePath("proj")
		want := filepath.Join(home, ".beans", "worktrees", "proj")
		if err != nil || got != want || !legacy {
			t.Errorf("ResolveWorktreePath() = %q, %v, %v; want %q, true, nil", got, legacy, err, want)
		}
	})

	t.Run("legacy root absent uses the lif root without creating it", func(t *testing.T) {
		home := homeWithRoots(t, "proj", false, false)
		got, legacy, err := (&Config{}).ResolveWorktreePath("proj")
		want := filepath.Join(home, ".lif", "worktrees", "proj")
		if err != nil || got != want || legacy {
			t.Errorf("ResolveWorktreePath() = %q, %v, %v; want %q, false, nil", got, legacy, err, want)
		}
		if _, statErr := os.Stat(want); !os.IsNotExist(statErr) {
			t.Errorf("resolver created %s (stat err = %v); it must only resolve", want, statErr)
		}
	})

	t.Run("both roots present: legacy wins, no error", func(t *testing.T) {
		home := homeWithRoots(t, "proj", true, true)
		got, legacy, err := (&Config{}).ResolveWorktreePath("proj")
		want := filepath.Join(home, ".beans", "worktrees", "proj")
		if err != nil || got != want || !legacy {
			t.Errorf("ResolveWorktreePath() = %q, %v, %v; want %q, true, nil", got, legacy, err, want)
		}
	})

	t.Run("legacy root of another project does not count", func(t *testing.T) {
		home := homeWithRoots(t, "other", true, false)
		got, legacy, err := (&Config{}).ResolveWorktreePath("proj")
		want := filepath.Join(home, ".lif", "worktrees", "proj")
		if err != nil || got != want || legacy {
			t.Errorf("ResolveWorktreePath() = %q, %v, %v; want %q, false, nil", got, legacy, err, want)
		}
	})

	t.Run("a legacy file instead of a directory does not count", func(t *testing.T) {
		home := homeWithRoots(t, "proj", false, false)
		writeFile(t, filepath.Join(home, ".beans", "worktrees", "proj"), "not a directory")
		got, legacy, err := (&Config{}).ResolveWorktreePath("proj")
		want := filepath.Join(home, ".lif", "worktrees", "proj")
		if err != nil || got != want || legacy {
			t.Errorf("ResolveWorktreePath() = %q, %v, %v; want %q, false, nil", got, legacy, err, want)
		}
	})

	t.Run("empty project name resolves under default", func(t *testing.T) {
		home := homeWithRoots(t, "default", false, false)
		got, legacy, err := (&Config{}).ResolveWorktreePath("")
		want := filepath.Join(home, ".lif", "worktrees", "default")
		if err != nil || got != want || legacy {
			t.Errorf("ResolveWorktreePath() = %q, %v, %v; want %q, false, nil", got, legacy, err, want)
		}
	})

	t.Run("worktree.path wins and is never reported legacy", func(t *testing.T) {
		home := homeWithRoots(t, "proj", true, true)
		custom := t.TempDir()
		cfg := &Config{Worktree: WorktreeConfig{Path: custom}}
		got, legacy, err := cfg.ResolveWorktreePath("proj")
		if err != nil || got != custom || legacy {
			t.Errorf("ResolveWorktreePath() = %q, %v, %v; want %q, false, nil", got, legacy, err, custom)
		}

		cfg = &Config{Worktree: WorktreeConfig{Path: "~/custom-root"}}
		got, legacy, err = cfg.ResolveWorktreePath("proj")
		want := filepath.Join(home, "custom-root")
		if err != nil || got != want || legacy {
			t.Errorf("ResolveWorktreePath(~ path) = %q, %v, %v; want %q, false, nil", got, legacy, err, want)
		}
	})
}
