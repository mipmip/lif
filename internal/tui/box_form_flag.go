package tui

// box_form_flag.go — env gate for the experimental jira-style box-form
// detail render (../docs/topics/jira-style-experiment/, S2b). Mirrors
// theme.asciiIcons()'s env pattern (internal/theme/theme.go) 1:1: read
// per-call (no caching), so tests can flip it via t.Setenv without any
// package-state reset dance. Default ON (since lif 0.13.0); set
// LIF_TUI_BOXFORM=0|false|no|off for the classic accordion detail.

import (
	"os"
	"strings"
)

// boxFormEnabled reports whether the experimental jira-style box-form detail
// render is active. Default on; LIF_TUI_BOXFORM=0|false|no|off disables it.
// Read per-call (no caching), like asciiIcons().
func boxFormEnabled() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("LIF_TUI_BOXFORM"))) {
	case "0", "false", "no", "off":
		return false
	}
	return true
}
