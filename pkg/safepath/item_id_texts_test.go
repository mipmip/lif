package safepath

import "testing"

// AC3: the rejection of a malformed ID names the noun as item.
func TestValidateItemIDErrorsNameItem(t *testing.T) {
	cases := []struct {
		id   string
		want string
	}{
		{"", "item ID must not be empty"},
		{"a/b", `invalid item ID "a/b": must contain only alphanumeric characters, hyphens, and underscores`},
	}
	for _, tc := range cases {
		if err := ValidateItemID(tc.id); err == nil || err.Error() != tc.want {
			t.Errorf("ValidateItemID(%q) error = %v, want %q", tc.id, err, tc.want)
		}
	}
}
