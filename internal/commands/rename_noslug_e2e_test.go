package commands

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// A prefixed item renamed with --no-slug must keep resolving under its ID.
func TestRenameNoSlugKeepsPrefixedID(t *testing.T) {
	dir := t.TempDir()
	store := filepath.Join(dir, "store")
	if err := os.Mkdir(store, 0o755); err != nil {
		t.Fatal(err)
	}
	run := func(args ...string) string {
		t.Helper()
		cmd := exec.Command(completionTestBinary, append([]string{"--lif-path", store}, args...)...)
		cmd.Dir = dir
		cmd.Env = append(cmd.Environ(), "HOME="+dir)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("lif %v: %v\n%s", args, err, out)
		}
		return string(out)
	}

	run("create", "x", "--prefix", "tp-", "--json")
	list := run("list", "--json")
	i := strings.Index(list, `"id": "tp-`)
	if i < 0 {
		t.Fatalf("no tp- id in list output:\n%s", list)
	}
	id := strings.SplitN(list[i+len(`"id": "`):], `"`, 2)[0]

	run("rename", id, "--no-slug")
	if got := run("show", id); !strings.Contains(got, "title: x") {
		t.Fatalf("show %s after --no-slug:\n%s", id, got)
	}
	if got := run("list", "--json"); !strings.Contains(got, `"id": "`+id+`"`) {
		t.Fatalf("list lost id %s:\n%s", id, got)
	}
}
