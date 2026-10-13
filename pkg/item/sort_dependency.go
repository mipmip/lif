package item

import "sort"

// SortRoadmapContainers sorts type-homogeneous items (e.g. all milestones, or
// all epics within one milestone) by status, then by dependency order
// (Blocking/BlockedBy edges within the slice), then by manual order
// (fractional index), then priority, then created_at (oldest first).
func SortRoadmapContainers(items []*Item, statusNames, priorityNames []string) {
	statusOrder := make(map[string]int, len(statusNames))
	for i, s := range statusNames {
		statusOrder[s] = i
	}
	getStatusOrder := func(status string) int {
		if order, ok := statusOrder[status]; ok {
			return order
		}
		return len(statusNames)
	}

	sortRoadmapBuckets(items, priorityNames, getStatusOrder)
}

// SortRoadmapLeaves sorts type-mixed leaf items (e.g. tasks and bugs directly
// under a milestone) by type first, then applies the same status →
// dependency → order → priority → created_at chain as SortRoadmapContainers
// within each type bucket.
func SortRoadmapLeaves(items []*Item, statusNames, priorityNames, typeNames []string) {
	typeOrder := make(map[string]int, len(typeNames))
	for i, t := range typeNames {
		typeOrder[t] = i
	}
	getTypeOrder := func(typ string) int {
		if order, ok := typeOrder[typ]; ok {
			return order
		}
		return len(typeNames)
	}

	buckets := bucketBy(items, func(b *Item) int { return getTypeOrder(b.Type) })
	for _, bucket := range buckets {
		SortRoadmapContainers(bucket, statusNames, priorityNames)
	}
	copy(items, flatten(buckets))
}

// sortRoadmapBuckets groups items by status rank and runs the
// dependency-aware ordering within each bucket.
func sortRoadmapBuckets(items []*Item, priorityNames []string, getStatusOrder func(string) int) {
	buckets := bucketBy(items, func(b *Item) int { return getStatusOrder(b.Status) })
	for _, bucket := range buckets {
		sortByDependencyOrderPriorityAndCreated(bucket, priorityNames)
	}
	copy(items, flatten(buckets))
}

// bucketBy groups items into buckets keyed by rank(b), ordered ascending by
// rank; items within a bucket keep their original relative order.
func bucketBy(items []*Item, rank func(*Item) int) [][]*Item {
	order := []int{}
	buckets := map[int][]*Item{}
	for _, b := range items {
		r := rank(b)
		if _, ok := buckets[r]; !ok {
			order = append(order, r)
		}
		buckets[r] = append(buckets[r], b)
	}
	sort.Ints(order)
	result := make([][]*Item, len(order))
	for i, r := range order {
		result[i] = buckets[r]
	}
	return result
}

func flatten(buckets [][]*Item) []*Item {
	total := 0
	for _, bucket := range buckets {
		total += len(bucket)
	}
	result := make([]*Item, 0, total)
	for _, bucket := range buckets {
		result = append(result, bucket...)
	}
	return result
}

// sortByDependencyOrderPriorityAndCreated reorders items within a single
// status (and, for leaves, type) bucket: dependency edges (Blocking /
// BlockedBy, restricted to items within this bucket) take precedence over
// the manual order key, which takes precedence over priority, which takes
// precedence over created_at. Items left over after a cycle can't be
// resolved further are appended in the same tie-break order instead of
// hanging or panicking.
func sortByDependencyOrderPriorityAndCreated(items []*Item, priorityNames []string) {
	if len(items) < 2 {
		return
	}

	priorityOrder := make(map[string]int, len(priorityNames))
	for i, p := range priorityNames {
		priorityOrder[p] = i
	}
	normalPriorityOrder := len(priorityNames)
	for i, p := range priorityNames {
		if p == "normal" {
			normalPriorityOrder = i
			break
		}
	}
	getPriorityOrder := func(priority string) int {
		if priority == "" {
			return normalPriorityOrder
		}
		if order, ok := priorityOrder[priority]; ok {
			return order
		}
		return len(priorityNames)
	}

	less := func(a, b *Item) bool {
		aHas, bHas := a.Order != "", b.Order != ""
		if aHas && bHas {
			if a.Order != b.Order {
				return a.Order < b.Order
			}
		} else if aHas != bHas {
			return aHas
		}
		pa, pb := getPriorityOrder(a.Priority), getPriorityOrder(b.Priority)
		if pa != pb {
			return pa < pb
		}
		if a.CreatedAt != nil && b.CreatedAt != nil {
			if !a.CreatedAt.Equal(*b.CreatedAt) {
				return a.CreatedAt.Before(*b.CreatedAt)
			}
		} else if (a.CreatedAt != nil) != (b.CreatedAt != nil) {
			return a.CreatedAt != nil
		}
		return a.ID < b.ID
	}

	present := make(map[string]bool, len(items))
	byID := make(map[string]*Item, len(items))
	for _, b := range items {
		present[b.ID] = true
		byID[b.ID] = b
	}

	// blockedBy[x] = set of items that must come before x.
	blockedBy := make(map[string]map[string]bool, len(items))
	addEdge := func(beforeID, afterID string) {
		if beforeID == afterID || !present[beforeID] || !present[afterID] {
			return
		}
		if blockedBy[afterID] == nil {
			blockedBy[afterID] = map[string]bool{}
		}
		blockedBy[afterID][beforeID] = true
	}
	for _, b := range items {
		for _, blockedID := range b.Blocking {
			addEdge(b.ID, blockedID)
		}
		for _, blockerID := range b.BlockedBy {
			addEdge(blockerID, b.ID)
		}
	}

	remaining := make(map[string]bool, len(items))
	for _, b := range items {
		remaining[b.ID] = true
	}

	result := make([]*Item, 0, len(items))
	for len(remaining) > 0 {
		var ready []*Item
		for id := range remaining {
			blocked := false
			for depID := range blockedBy[id] {
				if remaining[depID] {
					blocked = true
					break
				}
			}
			if !blocked {
				ready = append(ready, byID[id])
			}
		}
		if len(ready) == 0 {
			// Cycle: no item is free of an unresolved dependency. Break the
			// tie deterministically instead of hanging.
			for id := range remaining {
				ready = append(ready, byID[id])
			}
		}
		sort.Slice(ready, func(i, j int) bool { return less(ready[i], ready[j]) })
		next := ready[0]
		result = append(result, next)
		delete(remaining, next.ID)
	}

	copy(items, result)
}
