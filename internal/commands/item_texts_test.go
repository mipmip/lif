package commands

import (
	"strings"
	"testing"

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
