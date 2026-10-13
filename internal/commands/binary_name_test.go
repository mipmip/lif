package commands

import (
	"os/exec"
	"strings"
	"testing"
)

// The compiled root binary is named lif: `version` leads with the product
// name and the deprecated serve/tui stubs name the lif-* replacements.
func TestVersionLineStartsWithLif(t *testing.T) {
	out, err := exec.Command(completionTestBinary, "version").Output()
	if err != nil {
		t.Fatalf("lif version: %v", err)
	}
	first := strings.SplitN(string(out), "\n", 2)[0]
	if !strings.HasPrefix(first, "lif ") {
		t.Errorf("first version line = %q, want prefix %q", first, "lif ")
	}
}

func TestDeprecatedStubsNameLifBinaries(t *testing.T) {
	dir := t.TempDir()
	if out, err := exec.Command(completionTestBinary, "init", "--beans-path", dir+"/.beans").CombinedOutput(); err != nil {
		t.Fatalf("lif init: %v\n%s", err, out)
	}
	for verb, want := range map[string]string{"serve": "lif-serve", "tui": "lif-tui"} {
		cmd := exec.Command(completionTestBinary, "--beans-path", dir+"/.beans", verb)
		out, err := cmd.CombinedOutput()
		ee, ok := err.(*exec.ExitError)
		if !ok || ee.ExitCode() != 1 {
			t.Errorf("%s: err = %v, want exit status 1", verb, err)
		}
		if !strings.Contains(string(out), `"`+want+`"`) {
			t.Errorf("%s stub output = %q, want it to name %q", verb, out, want)
		}
	}
}
