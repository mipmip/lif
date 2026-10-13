---
paths:
  - "pkg/item/**"
  - "pkg/itemcore/**"
  - "pkg/itemgraph/**"
  - "internal/graph/**"
  - "internal/commands/**"
---

# Item Data Model

## Sorting Consistency

Item sorting must be consistent across the entire system. All places that return lists of items should use `item.SortByStatusPriorityAndType()` from `pkg/item/sort.go`. The frontend mirrors this in `sortItems()` in `frontend/src/lib/items.svelte.ts`. If you change the sort order, update both.

The sort order is: status → manual order (fractional index) → priority → type → title (case-insensitive).

## Relationship Validation

When modifying item relationships (parent, blocking, blocked-by):

- Always check for **cycles** using `Core.DetectCycle()` before adding a relationship
- Validate **type hierarchy** for parent relationships (e.g., a task can't be parent of a milestone)
- Normalize short IDs to full IDs via `Core.NormalizeID()` before storing
- Use the `itemgraph.CoreResolver` methods (`SetParent`, `AddBlocking`, etc., in `pkg/itemgraph/mutations.go`) — don't inline this logic

## ETag Concurrency Control

Items use content-hash ETags for optimistic concurrency:

- `item.ETag()` computes a hash of the item's content
- The `ifMatch` parameter on mutations validates the caller has seen the latest version
- `require_if_match` in config can make ETags mandatory for all mutations
- The frontend doesn't currently use ETags, but the CLI does (especially for body modifications)

## Dirty Items

`itemcore.Core` tracks items modified at runtime but not yet written to disk:

- `dirty` map in Core tracks which item IDs have unsaved changes
- Worktree item changes use `WithPersist(false)` — they're runtime-only until the PR merges to main
- The `saveDirtyItems` and `saveItem` mutations allow explicit persistence
- The `hasDirtyItems` query lets the frontend show unsaved state indicators
