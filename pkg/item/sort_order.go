package item

import (
	"sort"
	"strings"
)

// SortByOrder sorts items in place by their fractional-index Order key alone.
//
// This is the shared "sibling order" used by the `order` command (placement)
// and `list --sort order` (display): items with a set Order come first,
// compared lexicographically by the Order string; items without an Order
// sort after all of those. Any remaining tie (most commonly two items that
// both have no Order) breaks on Title, case-insensitively, for determinism.
func SortByOrder(items []*Item) {
	sort.SliceStable(items, func(i, j int) bool {
		oi, oj := items[i].Order, items[j].Order
		iHas, jHas := oi != "", oj != ""

		if iHas != jHas {
			// Items with an explicit order come before those without.
			return iHas
		}
		if iHas && jHas && oi != oj {
			return oi < oj
		}

		return strings.ToLower(items[i].Title) < strings.ToLower(items[j].Title)
	})
}
