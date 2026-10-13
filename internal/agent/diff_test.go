package agent

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestComputeUnifiedDiff_IdenticalContent(t *testing.T) {
	content := "hello\nworld\n"
	diff := computeUnifiedDiff(content, content, "test.txt")
	if diff != "" {
		t.Errorf("expected empty diff for identical content, got: %s", diff)
	}
}

func TestComputeUnifiedDiff_NewFile(t *testing.T) {
	diff := computeUnifiedDiff("", "line1\nline2\n", "new.txt")
	if diff == "" {
		t.Fatal("expected non-empty diff for new file")
	}
	if !strings.Contains(diff, "+line1") {
		t.Errorf("expected diff to contain +line1, got: %s", diff)
	}
	if !strings.Contains(diff, "+line2") {
		t.Errorf("expected diff to contain +line2, got: %s", diff)
	}
	if !strings.Contains(diff, "a/new.txt") {
		t.Errorf("expected diff to contain file label, got: %s", diff)
	}
}

func TestComputeUnifiedDiff_ModifiedContent(t *testing.T) {
	old := "line1\nline2\nline3\n"
	new := "line1\nmodified\nline3\n"
	diff := computeUnifiedDiff(old, new, "src/main.go")
	if diff == "" {
		t.Fatal("expected non-empty diff for modified content")
	}
	if !strings.Contains(diff, "-line2") {
		t.Errorf("expected diff to contain -line2, got: %s", diff)
	}
	if !strings.Contains(diff, "+modified") {
		t.Errorf("expected diff to contain +modified, got: %s", diff)
	}
	if !strings.Contains(diff, "a/src/main.go") {
		t.Errorf("expected diff to contain file label, got: %s", diff)
	}
}

// Q06: the two temp files computeUnifiedDiff hands to diff carry the lif
// prefix. A recording stand-in for diff on PATH sees the file names, which
// the --label arguments otherwise keep out of the real output.
func TestComputeUnifiedDiff_TempFilesUseLifPrefix(t *testing.T) {
	stubDir := t.TempDir()
	record := filepath.Join(stubDir, "args")
	script := "#!/bin/sh\nfor a in \"$@\"; do echo \"$a\" >> " + record + "; done\nexit 1\n"
	if err := os.WriteFile(filepath.Join(stubDir, "diff"), []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", stubDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	computeUnifiedDiff("a\n", "b\n", "f.txt")

	raw, err := os.ReadFile(record)
	if err != nil {
		t.Fatalf("stand-in diff was not run: %v", err)
	}
	args := strings.Split(strings.TrimSpace(string(raw)), "\n")
	oldName, newName := filepath.Base(args[len(args)-2]), filepath.Base(args[len(args)-1])
	if !strings.HasPrefix(oldName, "lif-diff-old-") {
		t.Errorf("old temp file = %q, want prefix lif-diff-old-", oldName)
	}
	if !strings.HasPrefix(newName, "lif-diff-new-") {
		t.Errorf("new temp file = %q, want prefix lif-diff-new-", newName)
	}
}
