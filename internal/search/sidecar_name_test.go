package search

import (
	"os"
	"path/filepath"
	"testing"
)

// The sidecar of a persisted index is lif-etags.json.
func TestPersistedIndex_WritesTheLifEtagsSidecar(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "idx")
	idx, err := Open(dir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if err := idx.IndexItem(itemWith("a1", "Alpha", "x")); err != nil {
		t.Fatalf("IndexItem() error = %v", err)
	}
	if err := idx.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "lif-etags.json")); err != nil {
		t.Errorf("sidecar lif-etags.json missing: %v", err)
	}
}
