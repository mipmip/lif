package itemgraph

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/xRiErOS/lif/pkg/config"
	"github.com/xRiErOS/lif/pkg/item"
	"github.com/xRiErOS/lif/pkg/itemcore"
	"github.com/xRiErOS/lif/pkg/itemgraph/model"
)

// newTestResolver builds a resolver over a throwaway store.
func newTestResolver(t *testing.T) *CoreResolver {
	t.Helper()
	itemsDir := filepath.Join(t.TempDir(), ".beans")
	if err := os.MkdirAll(itemsDir, 0755); err != nil {
		t.Fatalf("creating test .beans dir: %v", err)
	}
	c := itemcore.New(itemsDir, config.Default())
	if err := c.Load(); err != nil {
		t.Fatalf("loading core: %v", err)
	}
	return &CoreResolver{Core: c}
}

// TestUpdateItemWritesTagsInDeterministicOrder pins the tag order down. The
// merge builds its result from a Go map, whose iteration order is randomized
// per run, so without an explicit sort the same input can produce different
// files on consecutive runs.
func TestUpdateItemWritesTagsInDeterministicOrder(t *testing.T) {
	r := newTestResolver(t)
	ctx := context.Background()

	b := &item.Item{
		ID:     "beans-tag1",
		Slug:   item.Slugify("Tag ordering"),
		Title:  "Tag ordering",
		Status: "todo",
		Type:   "task",
		Tags:   []string{"zeta", "mu"},
	}
	if err := r.Core.Create(b); err != nil {
		t.Fatalf("core.Create: %v", err)
	}

	got, err := r.UpdateItem(ctx, b.ID, model.UpdateBeanInput{
		AddTags:    []string{"alpha", "omega"},
		RemoveTags: []string{"mu"},
	})
	if err != nil {
		t.Fatalf("UpdateBean: %v", err)
	}

	want := []string{"alpha", "omega", "zeta"}
	if len(got.Tags) != len(want) {
		t.Fatalf("tags = %v, want %v", got.Tags, want)
	}
	for i := range want {
		if got.Tags[i] != want[i] {
			t.Fatalf("tags = %v, want %v", got.Tags, want)
		}
	}
}
