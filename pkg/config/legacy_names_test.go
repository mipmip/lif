package config

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// legacy-name resolution (R-05..R-07): one rule per level -- only the new
// name present uses it, only the legacy name present uses it in place, both
// present is an error naming both, neither present uses the new name. Every
// test goes through the production resolver; the names come from the
// production constants.

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%s): %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%s): %v", path, err)
	}
}

// dirListing returns the sorted names directly inside dir.
func dirListing(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir(%s): %v", dir, err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	sort.Strings(names)
	return names
}

func requireNames(t *testing.T, err error, wants ...string) {
	t.Helper()
	if err == nil {
		t.Fatalf("error = nil, want one naming %q", wants)
	}
	for _, w := range wants {
		if !strings.Contains(err.Error(), w) {
			t.Errorf("error %q does not name %q", err.Error(), w)
		}
	}
}

func TestResolveConfigFile(t *testing.T) {
	t.Run("new name only is used", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, ConfigFileName), "lif:\n  prefix: a-\n")
		got, found, err := ResolveConfigFile(dir)
		if err != nil || !found || got != filepath.Join(dir, ConfigFileName) {
			t.Errorf("ResolveConfigFile() = %q, %v, %v; want the new-name file", got, found, err)
		}
	})
	t.Run("legacy name only is used in place", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, LegacyConfigFileName), "beans:\n  prefix: a-\n")
		got, found, err := ResolveConfigFile(dir)
		if err != nil || !found || got != filepath.Join(dir, LegacyConfigFileName) {
			t.Errorf("ResolveConfigFile() = %q, %v, %v; want the legacy file", got, found, err)
		}
	})
	t.Run("both names fail naming both and the directory, nothing written", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, ConfigFileName), "lif:\n")
		writeFile(t, filepath.Join(dir, LegacyConfigFileName), "beans:\n")
		before := dirListing(t, dir)
		_, _, err := ResolveConfigFile(dir)
		requireNames(t, err, ConfigFileName, LegacyConfigFileName, dir)
		if after := dirListing(t, dir); strings.Join(after, ",") != strings.Join(before, ",") {
			t.Errorf("directory changed: %v -> %v", before, after)
		}
	})
	t.Run("neither name yields the new name, not found", func(t *testing.T) {
		dir := t.TempDir()
		got, found, err := ResolveConfigFile(dir)
		if err != nil || found || got != filepath.Join(dir, ConfigFileName) {
			t.Errorf("ResolveConfigFile() = %q, %v, %v; want the new name, not found", got, found, err)
		}
	})
}

func TestFindConfigWithin_LegacyNames(t *testing.T) {
	t.Run("finds a legacy file in a parent directory", func(t *testing.T) {
		root := t.TempDir()
		sub := filepath.Join(root, "a", "b")
		if err := os.MkdirAll(sub, 0o755); err != nil {
			t.Fatal(err)
		}
		writeFile(t, filepath.Join(root, LegacyConfigFileName), "beans:\n")
		got, err := FindConfig(sub)
		if err != nil || got != filepath.Join(root, LegacyConfigFileName) {
			t.Errorf("FindConfig() = %q, %v; want the legacy file in the parent", got, err)
		}
	})
	t.Run("the nearest directory wins whatever its name", func(t *testing.T) {
		root := t.TempDir()
		sub := filepath.Join(root, "sub")
		writeFile(t, filepath.Join(root, ConfigFileName), "lif:\n")
		writeFile(t, filepath.Join(sub, LegacyConfigFileName), "beans:\n")
		got, err := FindConfig(sub)
		if err != nil || got != filepath.Join(sub, LegacyConfigFileName) {
			t.Errorf("FindConfig() = %q, %v; want the nearer legacy file", got, err)
		}
	})
	t.Run("both names in one directory of the search fail naming both", func(t *testing.T) {
		root := t.TempDir()
		sub := filepath.Join(root, "sub")
		writeFile(t, filepath.Join(root, ConfigFileName), "lif:\n")
		writeFile(t, filepath.Join(root, LegacyConfigFileName), "beans:\n")
		if err := os.MkdirAll(sub, 0o755); err != nil {
			t.Fatal(err)
		}
		_, err := FindConfig(sub)
		requireNames(t, err, ConfigFileName, LegacyConfigFileName, root)
	})
}

func TestLoad_SectionKey(t *testing.T) {
	load := func(t *testing.T, name, content string) (*Config, error) {
		t.Helper()
		path := filepath.Join(t.TempDir(), name)
		writeFile(t, path, content)
		return Load(path)
	}

	t.Run("lif key only", func(t *testing.T) {
		cfg, err := load(t, ConfigFileName, "lif:\n  prefix: new-\n  id_length: 6\n")
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if cfg.Items.Prefix != "new-" || cfg.Items.IDLength != 6 {
			t.Errorf("Items = %+v, want prefix new- and id_length 6", cfg.Items)
		}
		if got := cfg.SectionKey(); got != SectionKey {
			t.Errorf("SectionKey() = %q, want %q", got, SectionKey)
		}
	})
	t.Run("legacy key only is read in place", func(t *testing.T) {
		cfg, err := load(t, LegacyConfigFileName, "beans:\n  prefix: old-\n  id_length: 7\n")
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if cfg.Items.Prefix != "old-" || cfg.Items.IDLength != 7 {
			t.Errorf("Items = %+v, want prefix old- and id_length 7", cfg.Items)
		}
		if got := cfg.SectionKey(); got != LegacySectionKey {
			t.Errorf("SectionKey() = %q, want %q", got, LegacySectionKey)
		}
	})
	t.Run("both keys fail naming both keys and the file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, ConfigFileName)
		writeFile(t, path, "lif:\n  prefix: a-\nbeans:\n  prefix: b-\n")
		before := dirListing(t, dir)
		_, err := Load(path)
		requireNames(t, err, SectionKey+":", LegacySectionKey+":", path)
		if after := dirListing(t, dir); strings.Join(after, ",") != strings.Join(before, ",") {
			t.Errorf("directory changed: %v -> %v", before, after)
		}
	})
	t.Run("neither key uses the new name", func(t *testing.T) {
		cfg, err := load(t, LegacyConfigFileName, "project:\n  name: x\n")
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if got := cfg.SectionKey(); got != SectionKey {
			t.Errorf("SectionKey() = %q, want %q", got, SectionKey)
		}
	})
	t.Run("a config that was not loaded reports the new key", func(t *testing.T) {
		if got := Default().SectionKey(); got != SectionKey {
			t.Errorf("Default().SectionKey() = %q, want %q", got, SectionKey)
		}
	})
}

func TestResolveSectionKey(t *testing.T) {
	cases := []struct {
		name, content, want string
		wantErr             bool
	}{
		{"new only", "lif:\n", SectionKey, false},
		{"legacy only", "beans:\n", LegacySectionKey, false},
		{"both", "lif:\nbeans:\n", "", true},
		{"neither", "project:\n  name: x\n", SectionKey, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), LegacyConfigFileName)
			writeFile(t, path, tc.content)
			got, err := ResolveSectionKey(path)
			if tc.wantErr {
				requireNames(t, err, SectionKey+":", LegacySectionKey+":", path)
				return
			}
			if err != nil || got != tc.want {
				t.Errorf("ResolveSectionKey() = %q, %v; want %q", got, err, tc.want)
			}
		})
	}
}

func TestResolveDefaultStoreDir(t *testing.T) {
	t.Run("new directory only", func(t *testing.T) {
		dir := t.TempDir()
		mustMkdir(t, filepath.Join(dir, DefaultStoreDir))
		got, err := ResolveDefaultStoreDir(dir)
		if err != nil || got != DefaultStoreDir {
			t.Errorf("ResolveDefaultStoreDir() = %q, %v; want %q", got, err, DefaultStoreDir)
		}
	})
	t.Run("legacy directory only is used in place", func(t *testing.T) {
		dir := t.TempDir()
		mustMkdir(t, filepath.Join(dir, LegacyDefaultStoreDir))
		got, err := ResolveDefaultStoreDir(dir)
		if err != nil || got != LegacyDefaultStoreDir {
			t.Errorf("ResolveDefaultStoreDir() = %q, %v; want %q", got, err, LegacyDefaultStoreDir)
		}
	})
	t.Run("both directories fail naming both and the directory, nothing written", func(t *testing.T) {
		dir := t.TempDir()
		mustMkdir(t, filepath.Join(dir, DefaultStoreDir))
		mustMkdir(t, filepath.Join(dir, LegacyDefaultStoreDir))
		before := dirListing(t, dir)
		_, err := ResolveDefaultStoreDir(dir)
		requireNames(t, err, DefaultStoreDir, LegacyDefaultStoreDir, dir)
		if after := dirListing(t, dir); strings.Join(after, ",") != strings.Join(before, ",") {
			t.Errorf("directory changed: %v -> %v", before, after)
		}
	})
	t.Run("neither directory yields the new name", func(t *testing.T) {
		got, err := ResolveDefaultStoreDir(t.TempDir())
		if err != nil || got != DefaultStoreDir {
			t.Errorf("ResolveDefaultStoreDir() = %q, %v; want %q", got, err, DefaultStoreDir)
		}
	})
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
}

func TestResolveItemsPath_DefaultStoreDir(t *testing.T) {
	t.Run("an explicit path is honored verbatim even next to both default names", func(t *testing.T) {
		dir := t.TempDir()
		mustMkdir(t, filepath.Join(dir, DefaultStoreDir))
		mustMkdir(t, filepath.Join(dir, LegacyDefaultStoreDir))
		path := filepath.Join(dir, LegacyConfigFileName)
		writeFile(t, path, "beans:\n  path: "+LegacyDefaultStoreDir+"\n")
		cfg, err := Load(path)
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		got, err := cfg.ResolveItemsPath()
		if err != nil || got != filepath.Join(dir, LegacyDefaultStoreDir) {
			t.Errorf("ResolveItemsPath() = %q, %v; want the explicit path", got, err)
		}
	})
	t.Run("an explicit relative path outside the config directory is joined as is", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "sub", LegacyConfigFileName)
		writeFile(t, path, "beans:\n  path: ../"+LegacyDefaultStoreDir+"\n")
		cfg, err := Load(path)
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		got, err := cfg.ResolveItemsPath()
		if err != nil || got != filepath.Join(dir, LegacyDefaultStoreDir) {
			t.Errorf("ResolveItemsPath() = %q, %v; want %q", got, err, filepath.Join(dir, LegacyDefaultStoreDir))
		}
	})
	t.Run("no path: resolves the legacy default in the config directory", func(t *testing.T) {
		dir := t.TempDir()
		mustMkdir(t, filepath.Join(dir, LegacyDefaultStoreDir))
		path := filepath.Join(dir, ConfigFileName)
		writeFile(t, path, "lif:\n  prefix: a-\n")
		cfg, err := Load(path)
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		got, err := cfg.ResolveItemsPath()
		if err != nil || got != filepath.Join(dir, LegacyDefaultStoreDir) {
			t.Errorf("ResolveItemsPath() = %q, %v; want the legacy default", got, err)
		}
	})
	t.Run("no path: resolves the new default in the config directory", func(t *testing.T) {
		dir := t.TempDir()
		mustMkdir(t, filepath.Join(dir, DefaultStoreDir))
		path := filepath.Join(dir, LegacyConfigFileName)
		writeFile(t, path, "beans:\n  prefix: a-\n")
		cfg, err := Load(path)
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		got, err := cfg.ResolveItemsPath()
		if err != nil || got != filepath.Join(dir, DefaultStoreDir) {
			t.Errorf("ResolveItemsPath() = %q, %v; want the new default", got, err)
		}
	})
	t.Run("no path: with both default directories fails naming both", func(t *testing.T) {
		dir := t.TempDir()
		mustMkdir(t, filepath.Join(dir, DefaultStoreDir))
		mustMkdir(t, filepath.Join(dir, LegacyDefaultStoreDir))
		path := filepath.Join(dir, ConfigFileName)
		writeFile(t, path, "lif:\n  prefix: a-\n")
		cfg, err := Load(path)
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		_, err = cfg.ResolveItemsPath()
		requireNames(t, err, DefaultStoreDir, LegacyDefaultStoreDir, dir)
	})
	t.Run("no config file resolves in the start directory", func(t *testing.T) {
		dir := t.TempDir()
		mustMkdir(t, filepath.Join(dir, LegacyDefaultStoreDir))
		cfg, err := LoadFromDirectoryWithin(dir, dir)
		if err != nil {
			t.Fatalf("LoadFromDirectoryWithin() error = %v", err)
		}
		got, err := cfg.ResolveItemsPath()
		if err != nil || got != filepath.Join(dir, LegacyDefaultStoreDir) {
			t.Errorf("ResolveItemsPath() = %q, %v; want the legacy default in the start directory", got, err)
		}
	})
	t.Run("an anchor directory replaces the config directory for the default", func(t *testing.T) {
		cfgDir, anchor := t.TempDir(), t.TempDir()
		mustMkdir(t, filepath.Join(cfgDir, DefaultStoreDir))
		mustMkdir(t, filepath.Join(anchor, LegacyDefaultStoreDir))
		path := filepath.Join(cfgDir, ConfigFileName)
		writeFile(t, path, "lif:\n  anchor: repo-root\n")
		cfg, err := Load(path)
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		got, err := cfg.ResolveItemsPathIn(anchor)
		if err != nil || got != filepath.Join(anchor, LegacyDefaultStoreDir) {
			t.Errorf("ResolveItemsPathIn() = %q, %v; want the legacy default in the anchor", got, err)
		}
	})
}

func TestResolveStoreEnv(t *testing.T) {
	set := func(t *testing.T, lif, legacy string) {
		t.Helper()
		t.Setenv(StoreEnvVar, lif)
		t.Setenv(LegacyStoreEnvVar, legacy)
	}
	t.Run("new variable only", func(t *testing.T) {
		set(t, "/new", "")
		got, legacy, err := ResolveStoreEnv()
		if err != nil || got != "/new" || legacy {
			t.Errorf("ResolveStoreEnv() = %q, %v, %v; want /new, not legacy", got, legacy, err)
		}
	})
	t.Run("legacy variable only", func(t *testing.T) {
		set(t, "", "/old")
		got, legacy, err := ResolveStoreEnv()
		if err != nil || got != "/old" || !legacy {
			t.Errorf("ResolveStoreEnv() = %q, %v, %v; want /old, legacy", got, legacy, err)
		}
	})
	t.Run("both variables fail naming both", func(t *testing.T) {
		set(t, "/new", "/old")
		_, _, err := ResolveStoreEnv()
		requireNames(t, err, StoreEnvVar, LegacyStoreEnvVar)
	})
	t.Run("neither variable yields nothing", func(t *testing.T) {
		set(t, "", "")
		got, legacy, err := ResolveStoreEnv()
		if err != nil || got != "" || legacy {
			t.Errorf("ResolveStoreEnv() = %q, %v, %v; want nothing", got, legacy, err)
		}
	})
}

func TestSave_InPlace(t *testing.T) {
	t.Run("a legacy config is saved into the legacy file under its key", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, LegacyConfigFileName)
		writeFile(t, path, "beans:\n  path: .beans\n  prefix: old-\n")
		cfg, err := Load(path)
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		cfg.Items.Prefix = "renamed-"
		if err := cfg.Save(cfg.ConfigDir()); err != nil {
			t.Fatalf("Save() error = %v", err)
		}
		if got := dirListing(t, dir); strings.Join(got, ",") != LegacyConfigFileName {
			t.Errorf("directory = %v, want only %s", got, LegacyConfigFileName)
		}
		data, _ := os.ReadFile(path)
		if !strings.Contains(string(data), LegacySectionKey+":") || strings.Contains(string(data), SectionKey+":") {
			t.Errorf("saved file does not keep the legacy key only:\n%s", data)
		}
		reloaded, err := Load(path)
		if err != nil || reloaded.Items.Prefix != "renamed-" || reloaded.SectionKey() != LegacySectionKey {
			t.Errorf("reload = %+v, %v; want prefix renamed- under the legacy key", reloaded.Items, err)
		}
	})
	t.Run("a new config is saved into the new file under its key", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, ConfigFileName)
		writeFile(t, path, "lif:\n  path: .lif\n  prefix: new-\n")
		cfg, err := Load(path)
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		cfg.Items.Prefix = "renamed-"
		if err := cfg.Save(cfg.ConfigDir()); err != nil {
			t.Fatalf("Save() error = %v", err)
		}
		if got := dirListing(t, dir); strings.Join(got, ",") != ConfigFileName {
			t.Errorf("directory = %v, want only %s", got, ConfigFileName)
		}
		data, _ := os.ReadFile(path)
		if !strings.Contains(string(data), SectionKey+":") || strings.Contains(string(data), LegacySectionKey+":") {
			t.Errorf("saved file does not keep the new key only:\n%s", data)
		}
	})
	t.Run("a config that was not loaded writes into the legacy file already there", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, LegacyConfigFileName), "beans:\n  prefix: old-\n")
		cfg := DefaultWithPrefix("fresh-")
		cfg.SetConfigDir(dir)
		if err := cfg.Save(dir); err != nil {
			t.Fatalf("Save() error = %v", err)
		}
		if got := dirListing(t, dir); strings.Join(got, ",") != LegacyConfigFileName {
			t.Errorf("directory = %v, want only %s", got, LegacyConfigFileName)
		}
		reloaded, err := Load(filepath.Join(dir, LegacyConfigFileName))
		if err != nil || reloaded.Items.Prefix != "fresh-" || reloaded.SectionKey() != LegacySectionKey {
			t.Errorf("reload = %+v, %v; want prefix fresh- under the legacy key", reloaded.Items, err)
		}
	})
	t.Run("a config that was not loaded writes into the new file already there", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, ConfigFileName), "lif:\n  prefix: old-\n")
		cfg := DefaultWithPrefix("fresh-")
		if err := cfg.Save(dir); err != nil {
			t.Fatalf("Save() error = %v", err)
		}
		if got := dirListing(t, dir); strings.Join(got, ",") != ConfigFileName {
			t.Errorf("directory = %v, want only %s", got, ConfigFileName)
		}
	})
	t.Run("an existing file keeps its own key when it has the legacy key in the new file", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, ConfigFileName), "beans:\n  prefix: old-\n")
		cfg := DefaultWithPrefix("fresh-")
		if err := cfg.Save(dir); err != nil {
			t.Fatalf("Save() error = %v", err)
		}
		reloaded, err := Load(filepath.Join(dir, ConfigFileName))
		if err != nil || reloaded.SectionKey() != LegacySectionKey {
			t.Errorf("SectionKey() = %q, %v; want the file's own key %q", reloaded.SectionKey(), err, LegacySectionKey)
		}
	})
	t.Run("an empty directory gets the new file with the new key", func(t *testing.T) {
		dir := t.TempDir()
		cfg := DefaultWithPrefix("fresh-")
		if err := cfg.Save(dir); err != nil {
			t.Fatalf("Save() error = %v", err)
		}
		if got := dirListing(t, dir); strings.Join(got, ",") != ConfigFileName {
			t.Errorf("directory = %v, want only %s", got, ConfigFileName)
		}
		reloaded, err := Load(filepath.Join(dir, ConfigFileName))
		if err != nil || reloaded.SectionKey() != SectionKey || reloaded.Items.Prefix != "fresh-" {
			t.Errorf("reload = %+v (key %q), %v; want prefix fresh- under %q", reloaded.Items, reloaded.SectionKey(), err, SectionKey)
		}
	})
	t.Run("both config files fail naming both and write nothing", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, ConfigFileName), "lif:\n")
		writeFile(t, filepath.Join(dir, LegacyConfigFileName), "beans:\n")
		before := dirListing(t, dir)
		err := DefaultWithPrefix("fresh-").Save(dir)
		requireNames(t, err, ConfigFileName, LegacyConfigFileName, dir)
		if after := dirListing(t, dir); strings.Join(after, ",") != strings.Join(before, ",") {
			t.Errorf("directory changed: %v -> %v", before, after)
		}
	})
	t.Run("an omitted path stays omitted", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, LegacyConfigFileName)
		writeFile(t, path, "beans:\n  prefix: old-\n")
		cfg, err := Load(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := cfg.Save(dir); err != nil {
			t.Fatal(err)
		}
		reloaded, err := Load(path)
		if err != nil || reloaded.Items.Path != "" {
			t.Errorf("Items.Path = %q, %v; want it still omitted", reloaded.Items.Path, err)
		}
	})
}

func TestLegacyNamePredicates(t *testing.T) {
	if !IsLegacyConfigFile(filepath.Join("x", LegacyConfigFileName)) || IsLegacyConfigFile(filepath.Join("x", ConfigFileName)) {
		t.Error("IsLegacyConfigFile does not tell the two config file names apart")
	}
	if !IsLegacyStoreDir(filepath.Join("x", LegacyDefaultStoreDir)) || IsLegacyStoreDir(filepath.Join("x", DefaultStoreDir)) {
		t.Error("IsLegacyStoreDir does not tell the two store directory names apart")
	}
}
