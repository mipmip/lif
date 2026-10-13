package tui

import (
	"os"
	"testing"
)

// The box-form detail is the product default; the pre-existing suite asserts
// the classic accordion, so pin it here. Tests for box-form opt in via
// t.Setenv("LIF_TUI_BOXFORM", "1"), which overrides this.
func TestMain(m *testing.M) {
	os.Setenv("LIF_TUI_BOXFORM", "0")
	os.Exit(m.Run())
}
