package safepath

import (
	"testing"
)

func TestSafeJoin(t *testing.T) {
	tests := []struct {
		name      string
		root      string
		untrusted string
		want      string
		wantErr   bool
	}{
		{
			name:      "simple filename",
			root:      "/home/user/.beans",
			untrusted: "bean-abc.md",
			want:      "/home/user/.beans/bean-abc.md",
		},
		{
			name:      "subdirectory path",
			root:      "/home/user/.beans",
			untrusted: "archive/bean-abc.md",
			want:      "/home/user/.beans/archive/bean-abc.md",
		},
		{
			name:      "path traversal with ..",
			root:      "/home/user/.beans",
			untrusted: "../../../etc/passwd",
			wantErr:   true,
		},
		{
			name:      "path traversal with embedded ..",
			root:      "/home/user/.beans",
			untrusted: "archive/../../etc/passwd",
			wantErr:   true,
		},
		{
			name:      "dot-dot escaping root",
			root:      "/home/user/.beans",
			untrusted: "..",
			wantErr:   true,
		},
		{
			name:      "absolute path ignores root",
			root:      "/home/user/.beans",
			untrusted: "/etc/passwd",
			wantErr:   true,
		},
		{
			name:      "current directory",
			root:      "/home/user/.beans",
			untrusted: ".",
			want:      "/home/user/.beans",
		},
		{
			name:      "nested safe path",
			root:      "/home/user/.beans",
			untrusted: ".conversations/bean-abc.jsonl",
			want:      "/home/user/.beans/.conversations/bean-abc.jsonl",
		},
		{
			name:      "trailing slash on root",
			root:      "/home/user/.beans/",
			untrusted: "file.md",
			want:      "/home/user/.beans/file.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SafeJoin(tt.root, tt.untrusted)
			if tt.wantErr {
				if err == nil {
					t.Errorf("SafeJoin(%q, %q) = %q, want error", tt.root, tt.untrusted, got)
				}
				return
			}
			if err != nil {
				t.Errorf("SafeJoin(%q, %q) unexpected error: %v", tt.root, tt.untrusted, err)
				return
			}
			if got != tt.want {
				t.Errorf("SafeJoin(%q, %q) = %q, want %q", tt.root, tt.untrusted, got, tt.want)
			}
		})
	}
}

func TestValidateItemID(t *testing.T) {
	tests := []struct {
		name    string
		itemID  string
		wantErr bool
	}{
		{name: "valid simple", itemID: "bean-abc1", wantErr: false},
		{name: "valid with underscores", itemID: "__central__", wantErr: false},
		{name: "valid alphanumeric", itemID: "abc123", wantErr: false},
		{name: "valid with hyphens", itemID: "my-bean-id", wantErr: false},
		{name: "empty", itemID: "", wantErr: true},
		{name: "path traversal", itemID: "../../../etc/passwd", wantErr: true},
		{name: "contains slash", itemID: "bean/evil", wantErr: true},
		{name: "contains backslash", itemID: "bean\\evil", wantErr: true},
		{name: "contains dot-dot", itemID: "bean..evil", wantErr: true},
		{name: "contains space", itemID: "bean evil", wantErr: true},
		{name: "contains null byte", itemID: "bean\x00evil", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateItemID(tt.itemID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateBeanID(%q) error = %v, wantErr %v", tt.itemID, err, tt.wantErr)
			}
		})
	}
}
