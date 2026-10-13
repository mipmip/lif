package itemcore

import (
	"strings"
	"testing"

	"github.com/xRiErOS/lif/pkg/config"
	"github.com/xRiErOS/lif/pkg/item"
)

// AC3: the error and status texts the core returns name the noun as item and
// the product as lif. Each subtest drives the production function.
func TestUserFacingTextsNameItems(t *testing.T) {
	t.Run("ErrNotFound", func(t *testing.T) {
		if got := ErrNotFound.Error(); got != "item not found" {
			t.Errorf("ErrNotFound = %q, want %q", got, "item not found")
		}
	})

	t.Run("policy violation quotes lif complete", func(t *testing.T) {
		err := &PolicyViolationError{ItemID: "x-1", Status: "completed", Missing: []string{"commit"}}
		if got := err.Error(); !strings.Contains(got, "`lif complete x-1 --commit HEAD`") {
			t.Errorf("Error() = %q, want it to quote `lif complete x-1 --commit HEAD`", got)
		}
	})

	t.Run("parent not found", func(t *testing.T) {
		c, _ := setupTestCore(t)
		err := c.ValidateParent(&item.Item{ID: "x-1", Type: "task"}, "nope")
		if err == nil || err.Error() != "parent item not found: nope" {
			t.Errorf("ValidateParent() error = %v, want %q", err, "parent item not found: nope")
		}
	})

	t.Run("parent of the wrong type", func(t *testing.T) {
		c, _ := setupTestCore(t)
		task := &item.Item{ID: "task-1", Slug: "t", Title: "T", Status: "todo", Type: "task"}
		if err := c.Create(task); err != nil {
			t.Fatal(err)
		}
		err := c.ValidateParent(&item.Item{ID: "x-1", Type: "milestone"}, "task-1")
		if err == nil || !strings.HasPrefix(err.Error(), "milestone items ") {
			t.Errorf("ValidateParent() error = %v, want it to start with %q", err, "milestone items ")
		}
	})

	t.Run("rename of an unknown ID", func(t *testing.T) {
		c := newTestCore(t, "t-", nil)
		_, err := c.PlanRenameID("t-nope", "t-other")
		if err == nil || err.Error() != `item "t-nope" not found` {
			t.Errorf("PlanRenameID() error = %v, want %q", err, `item "t-nope" not found`)
		}
	})

	t.Run("rebrand without a matching item", func(t *testing.T) {
		c := newTestCore(t, "t-", nil)
		_, err := c.PlanRebrand("u-")
		if err == nil || !strings.Contains(err.Error(), "no items matched prefix") {
			t.Errorf("PlanRebrand() error = %v, want it to contain %q", err, "no items matched prefix")
		}
	})

	t.Run("prefix consistency", func(t *testing.T) {
		c := New(t.TempDir(), &config.Config{Items: config.ItemsConfig{Prefix: "a-"}})
		c.items["a-1"] = &item.Item{ID: "a-1"}
		c.items["b-1"] = &item.Item{ID: "b-1"}
		if got := c.ValidatePrefixConsistency(); !strings.Contains(got, "items have prefixes") {
			t.Errorf("mixed-prefix message = %q, want it to say %q", got, "items have prefixes")
		}

		c = New(t.TempDir(), &config.Config{Items: config.ItemsConfig{Prefix: "a-"}})
		c.items["b-1"] = &item.Item{ID: "b-1"}
		if got := c.ValidatePrefixConsistency(); !strings.Contains(got, "items have prefix ") {
			t.Errorf("prefix-mismatch message = %q, want it to say %q", got, "items have prefix ")
		}
	})
}
