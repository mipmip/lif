package item

import (
	"errors"
	"testing"
)

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, errors.New("disk gone") }

// AC3: a failed read names the noun as item and keeps the cause.
func TestParseReadErrorNamesItem(t *testing.T) {
	_, err := Parse(failingReader{})
	if err == nil || err.Error() != "reading item: disk gone" {
		t.Errorf("Parse() error = %v, want %q", err, "reading item: disk gone")
	}
}
