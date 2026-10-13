// Package safepath provides path traversal guards for user-influenced path components.
package safepath

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// itemIDPattern matches valid item IDs: alphanumeric characters, hyphens, and underscores.
// The special ID "__central__" (used for central agent chat) is also valid.
var itemIDPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// SafeJoin joins a trusted root directory with an untrusted relative path component,
// returning the resulting absolute path only if it stays within root.
// Returns an error if the result would escape the root directory.
func SafeJoin(root, untrusted string) (string, error) {
	// Reject absolute paths outright — the untrusted component must be relative.
	if filepath.IsAbs(untrusted) {
		return "", fmt.Errorf("path traversal detected: %q is absolute", untrusted)
	}

	// Clean the root to get a canonical form
	cleanRoot := filepath.Clean(root)

	// Join and clean the full path
	joined := filepath.Clean(filepath.Join(cleanRoot, untrusted))

	// Verify the result is within root. We check that the joined path either
	// equals root exactly or has root as a proper prefix followed by a separator.
	if joined != cleanRoot && !strings.HasPrefix(joined, cleanRoot+string(filepath.Separator)) {
		return "", fmt.Errorf("path traversal detected: %q escapes root %q", untrusted, root)
	}

	return joined, nil
}

// ValidateItemID checks that an item ID contains only safe characters
// (alphanumeric, hyphens, underscores). Returns an error if the ID is
// empty or contains unexpected characters that could be used for path
// traversal or command injection.
func ValidateItemID(itemID string) error {
	if itemID == "" {
		return fmt.Errorf("item ID must not be empty")
	}
	if !itemIDPattern.MatchString(itemID) {
		return fmt.Errorf("invalid item ID %q: must contain only alphanumeric characters, hyphens, and underscores", itemID)
	}
	return nil
}
