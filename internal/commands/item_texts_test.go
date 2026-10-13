package commands

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/xRiErOS/lif/pkg/config"
	"github.com/xRiErOS/lif/pkg/item"
)

// AC3: the tag completion describes each tag's usage count in items.
func TestTagCompletionCountsItems(t *testing.T) {
	setupListTest(t)
	for _, b := range []*item.Item{
		{ID: "t-1", Slug: "a", Title: "A", Status: "todo", Type: "task", Tags: []string{"echo"}},
		{ID: "t-2", Slug: "b", Title: "B", Status: "todo", Type: "task", Tags: []string{"echo"}},
	} {
		if err := core.Create(b); err != nil {
			t.Fatal(err)
		}
	}

	got, _ := tagFlagCompletion(nil, nil, "")
	if len(got) != 1 || got[0] != "echo\t2 item(s)" {
		t.Errorf("tagFlagCompletion() = %q, want [%q]", got, "echo\t2 item(s)")
	}
}

// AC3: update refuses --if-match for several IDs and names the noun as item.
func TestUpdateIfMatchWithSeveralIDsNamesItem(t *testing.T) {
	first := setupUpdateTest(t)
	oldIfMatch := updateIfMatch
	updateIfMatch = "deadbeefdeadbeef"
	t.Cleanup(func() { updateIfMatch = oldIfMatch })

	err := updateCmd.RunE(updateCmd, []string{first.ID, "beans-other"})
	want := "--if-match takes one item's etag, got 2 ids"
	if err == nil || err.Error() != want {
		t.Errorf("error = %v, want %q", err, want)
	}
}

// AC3: archive reports its outcome in items, in plain and in JSON mode.
func TestArchiveTextsNameItems(t *testing.T) {
	run := func(t *testing.T, asJSON, withCompleted bool) string {
		t.Helper()
		setupListTest(t)
		if withCompleted {
			done := &item.Item{ID: "t-1", Slug: "a", Title: "A", Status: "completed", Type: "task"}
			if err := core.Create(done); err != nil {
				t.Fatal(err)
			}
		}
		old := archiveJSON
		archiveJSON = asJSON
		t.Cleanup(func() { archiveJSON = old })

		return string(captureListStdout(t, func() {
			if err := archiveCmd.RunE(archiveCmd, nil); err != nil {
				t.Fatalf("archive error = %v", err)
			}
		}))
	}

	for _, tc := range []struct {
		name                  string
		asJSON, withCompleted bool
		want                  string
	}{
		{"nothing to archive", false, false, "No items with archive status to archive."},
		{"nothing to archive as json", true, false, "No items to archive"},
		{"archived", false, true, "Archived 1 item(s) to "},
		{"archived as json", true, true, "Archived 1 item(s) to "},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if out := run(t, tc.asJSON, tc.withCompleted); !strings.Contains(out, tc.want) {
				t.Errorf("output = %q, want it to contain %q", out, tc.want)
			}
		})
	}
}

// AC1/AC3: create's failure, its flag help and its completion hint name the
// noun as item and the product as lif.
func TestCreateTextsNameLifAndItems(t *testing.T) {
	t.Run("failure", func(t *testing.T) {
		setupListTest(t)
		t.Cleanup(func() { resetFlags(sharedTestRoot(t)) })

		_, _, err := runRootInDir(t, core.Root(), "create", "Some title", "--blocking", "nope")
		want := "failed to create item: target item not found: nope"
		if err == nil || err.Error() != want {
			t.Errorf("error = %v, want %q", err, want)
		}
	})

	t.Run("hint", func(t *testing.T) {
		if want := `lif create "Fix the login bug"`; !strings.Contains(createHint(), want) {
			t.Errorf("createHint() = %q, want it to contain %q", createHint(), want)
		}
	})

	t.Run("type flag", func(t *testing.T) {
		sharedTestRoot(t)
		if got := createCmd.Flags().Lookup("type").Usage; !strings.HasPrefix(got, "Item type (") {
			t.Errorf("--type usage = %q, want it to start with %q", got, "Item type (")
		}
	})
}

// AC3: a plain `lif init` confirms in the product's name.
func TestInitConfirmationNamesLif(t *testing.T) {
	var err error
	out := captureListStdout(t, func() { _, err = runInitCmd(t, "", false) })
	if err != nil {
		t.Fatalf("init error = %v", err)
	}
	if want := "Initialized Lif project\n"; string(out) != want {
		t.Errorf("init output = %q, want %q", out, want)
	}
}

// AC3: the shared batch resolution (start/complete/scrap/tag/update/...)
// names the noun as item when an ID repeats or does not resolve.
func TestBatchTargetErrorsNameItems(t *testing.T) {
	first := setupStartTest(t)
	resetStartFlags(t)

	dup := startCmd.RunE(startCmd, []string{first.ID, first.ID})
	if want := "item " + first.ID + " named more than once"; dup == nil || !strings.Contains(dup.Error(), want) {
		t.Errorf("duplicate-ID error = %v, want it to contain %q", dup, want)
	}

	unknown := startCmd.RunE(startCmd, []string{"beans-notfound"})
	if want := "item not found: beans-notfound"; unknown == nil || unknown.Error() != want {
		t.Errorf("unknown-ID error = %v, want %q", unknown, want)
	}
}

// AC3: an explicit store path that does not exist, and a missing store
// found by search, are reported without the former product name.
func TestStoreResolutionErrorsNameLif(t *testing.T) {
	neutralizeStoreEnv(t)
	c := config.Default()
	c.SetConfigDir(t.TempDir())

	missing := filepath.Join(t.TempDir(), "gone")
	_, _, err := resolveStorePath(missing, c)
	if want := "store path does not exist or is not a directory: " + missing; err == nil || err.Error() != want {
		t.Errorf("explicit path error = %v, want %q", err, want)
	}

	_, _, err = resolveStorePath("", c)
	if err == nil || !strings.Contains(err.Error(), "(run 'lif init' to create one)") {
		t.Errorf("missing store error = %v, want it to point at 'lif init'", err)
	}
}
