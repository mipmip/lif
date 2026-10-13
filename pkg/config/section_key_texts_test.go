package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Config errors name the section key of the file they come from.

func loadWithKey(t *testing.T, file, key, body string) (*Config, error) {
	t.Helper()
	path := filepath.Join(t.TempDir(), file)
	if err := os.WriteFile(path, []byte(key+":\n"+body), 0o644); err != nil {
		t.Fatal(err)
	}
	return Load(path)
}

func TestConfigErrors_NameTheSectionKeyInUse(t *testing.T) {
	for _, tc := range []struct{ file, key string }{
		{ConfigFileName, SectionKey},
		{LegacyConfigFileName, LegacySectionKey},
	} {
		t.Run(tc.key, func(t *testing.T) {
			_, err := loadWithKey(t, tc.file, tc.key, "    anchor: sideways\n")
			if err == nil || !strings.Contains(err.Error(), "unknown "+tc.key+".anchor") {
				t.Errorf("anchor error = %v, want it to name %q", err, tc.key+".anchor")
			}

			_, err = loadWithKey(t, tc.file, tc.key, "    require_fields_on:\n        nonsense: [x]\n")
			if err == nil || !strings.Contains(err.Error(), tc.key+".require_fields_on") {
				t.Errorf("unknown-status error = %v, want it to name %q", err, tc.key+".require_fields_on")
			}

			_, err = loadWithKey(t, tc.file, tc.key, "    require_fields_on:\n        todo: [\"\"]\n")
			if err == nil || !strings.Contains(err.Error(), tc.key+".require_fields_on[") {
				t.Errorf("empty-field error = %v, want it to name %q", err, tc.key+".require_fields_on[")
			}

			_, err = loadWithKey(t, tc.file, tc.key, "    require_fields_on:\n        todo: [title]\n")
			if err == nil || !strings.Contains(err.Error(), tc.key+".require_fields_on[") {
				t.Errorf("reserved-field error = %v, want it to name %q", err, tc.key+".require_fields_on[")
			}
		})
	}
}

func TestSave_WorktreePathCommentNamesTheLifRoot(t *testing.T) {
	cfg := Default()
	cfg.Worktree.Path = "/somewhere"
	dir := t.TempDir()
	cfg.SetConfigDir(dir)
	if err := cfg.Save(dir); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, ConfigFileName))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "default: ~/.lif/worktrees/<project>/") {
		t.Errorf("saved config does not name ~/.lif/worktrees/<project>/ in the worktree.path comment:\n%s", data)
	}
}

func TestTargetFileName_IsTheFileSaveWrites(t *testing.T) {
	legacyDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(legacyDir, LegacyConfigFileName), []byte("beans:\n    prefix: x-\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := Default().TargetFileName(legacyDir); got != LegacyConfigFileName {
		t.Errorf("TargetFileName(legacy dir) = %q, want %q", got, LegacyConfigFileName)
	}
	if got := Default().TargetFileName(t.TempDir()); got != ConfigFileName {
		t.Errorf("TargetFileName(empty dir) = %q, want %q", got, ConfigFileName)
	}
}

// With both config files present SaveTarget cannot decide; the message helper
// still names a file: the new one.
func TestTargetFileName_FallsBackToTheNewName(t *testing.T) {
	dir := t.TempDir()
	for _, f := range []string{ConfigFileName, LegacyConfigFileName} {
		if err := os.WriteFile(filepath.Join(dir, f), []byte("x: 1\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if got := Default().TargetFileName(dir); got != ConfigFileName {
		t.Errorf("TargetFileName(ambiguous dir) = %q, want %q", got, ConfigFileName)
	}
}
