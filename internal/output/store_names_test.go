package output

import (
	"encoding/json"
	"testing"
)

// The missing-store code is NO_LIF_DIR; there is no NO_BEANS_DIR alias.
func TestErrNoLifDir_IsNoLifDir(t *testing.T) {
	if ErrNoLifDir != "NO_LIF_DIR" {
		t.Errorf("ErrNoLifDir = %q, want %q", ErrNoLifDir, "NO_LIF_DIR")
	}
}

// SuccessInit names the directory init initialized, not a fixed one.
func TestSuccessInit_NamesTheInitializedDirectory(t *testing.T) {
	for _, dir := range []string{".lif", ".beans", "custom-store"} {
		out := captureStdout(t, func() {
			if err := SuccessInit("/some/project/" + dir); err != nil {
				t.Fatalf("SuccessInit() error = %v", err)
			}
		})
		var resp Response
		if err := json.Unmarshal(out, &resp); err != nil {
			t.Fatalf("bad JSON %q: %v", out, err)
		}
		if want := "Initialized " + dir + " directory"; resp.Message != want {
			t.Errorf("message = %q, want %q", resp.Message, want)
		}
	}
}
