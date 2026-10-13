package tui

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// runCmd runs cmd and returns the message it yields.
func runCmd(t *testing.T, cmd tea.Cmd) tea.Msg {
	t.Helper()
	if cmd == nil {
		t.Fatal("expected a command, got nil")
	}
	return cmd()
}

// AC4: the TUI labels its modals, help shortcuts and list with the item
// vocabulary. Every case renders the production model.
func TestTUILabelsNameItems(t *testing.T) {
	t.Run("help shortcuts", func(t *testing.T) {
		view := newHelpOverlayModel(100, 40).View()
		for _, want := range []string{"View item details", "Create new item", "Copy item ID"} {
			if !strings.Contains(view, want) {
				t.Errorf("help overlay lacks %q:\n%s", want, view)
			}
		}
	})

	t.Run("create modal", func(t *testing.T) {
		view := newCreateModalModel(100, 40).View()
		for _, want := range []string{"Create New Item", "Enter item title..."} {
			if !strings.Contains(view, want) {
				t.Errorf("create modal lacks %q:\n%s", want, view)
			}
		}
	})

	t.Run("list title", func(t *testing.T) {
		resolver, cfg := candidateFixture(t)
		m := newListModel(resolver, cfg)
		m.width, m.height = 100, 30
		m.list.SetSize(98, 26)

		if view := m.View(); !strings.Contains(view, "Items") {
			t.Errorf("list view lacks the title %q:\n%s", "Items", view)
		}
		m.tagFilter = "echo"
		if view := m.View(); !strings.Contains(view, "Items [tag: echo]") {
			t.Errorf("filtered list view lacks the title %q:\n%s", "Items [tag: echo]", view)
		}
	})

	t.Run("constrained list title", func(t *testing.T) {
		resolver, cfg := candidateFixture(t)
		m := newListModel(resolver, cfg)

		if view := m.ViewConstrained(100, 30); !strings.Contains(view, "Items") {
			t.Errorf("constrained list view lacks the title %q:\n%s", "Items", view)
		}
		m.tagFilter = "echo"
		if view := m.ViewConstrained(100, 30); !strings.Contains(view, "Items [tag: echo]") {
			t.Errorf("filtered constrained list view lacks the title %q:\n%s", "Items [tag: echo]", view)
		}
	})

	t.Run("linked items", func(t *testing.T) {
		resolver, cfg := candidateFixture(t)
		feature, err := resolver.Core.Get("beans-f1")
		if err != nil {
			t.Fatal(err)
		}
		if view := newDetailModel(feature, resolver, cfg, 100, 40).View(); !strings.Contains(view, "Linked Items") {
			t.Errorf("detail view lacks the title %q:\n%s", "Linked Items", view)
		}
	})
}

// AC4: the multi-select pickers title themselves "<n> selected items".
func TestMultiSelectPickersNameItems(t *testing.T) {
	resolver, cfg := candidateFixture(t)

	title := func(msg tea.Msg) string {
		switch m := msg.(type) {
		case openParentPickerMsg:
			return m.itemTitle
		case openStatusPickerMsg:
			return m.itemTitle
		case openTypePickerMsg:
			return m.itemTitle
		case openPriorityPickerMsg:
			return m.itemTitle
		}
		return "unexpected message"
	}

	for _, key := range []string{"p", "s", "t", "P"} {
		t.Run("key "+key, func(t *testing.T) {
			m := newListModel(resolver, cfg)
			m.selectedItems["beans-f1"] = true
			m.selectedItems["beans-f2"] = true

			_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
			if got := title(runCmd(t, cmd)); got != "2 selected items" {
				t.Errorf("picker title = %q, want %q", got, "2 selected items")
			}
		})
	}
}

// AC4: an item path that leaves the store is refused with a status message
// that names the noun as item.
func TestUnsafeItemPathStatusNamesItem(t *testing.T) {
	resolver, cfg := candidateFixture(t)
	app := New(resolver.Core, cfg)

	app.Update(openEditorMsg{itemID: "beans-f1", itemPath: "../../outside.md"})

	if got := app.list.statusMessage; !strings.HasPrefix(got, "unsafe item path: ") {
		t.Errorf("status message = %q, want it to start with %q", got, "unsafe item path: ")
	}
}

// AC4: copying several IDs reports "Copied <n> item IDs to clipboard". The
// clipboard command is a stand-in on PATH, so the real clipboard stays
// untouched; the library calls pbcopy by name on macOS only.
func TestCopyItemIDsStatusNamesItems(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("the clipboard library resolves its command at import time off macOS")
	}
	stubDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(stubDir, "pbcopy"), []byte("#!/bin/sh\ncat >/dev/null\n"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", stubDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	resolver, cfg := candidateFixture(t)
	app := New(resolver.Core, cfg)

	app.Update(copyItemIDMsg{ids: []string{"beans-f1", "beans-f2"}})

	if got := app.list.statusMessage; got != "Copied 2 item IDs to clipboard" {
		t.Errorf("status message = %q, want %q", got, "Copied 2 item IDs to clipboard")
	}
}
