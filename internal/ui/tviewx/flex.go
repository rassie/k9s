// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package tviewx

import "github.com/rivo/tview"

// Flex wraps tview.Flex and restores the index-based item operations that the
// derailed/tview fork exposed (ItemAt/AddItemAtIndex/RemoveItemAtIndex).
// rivo/tview only offers append/remove-by-primitive and keeps per-item sizing
// in unexported fields, so this wrapper tracks the items (and their sizes)
// itself and rebuilds the embedded flex on every structural change.
type Flex struct {
	*tview.Flex

	items []flexItem
}

type flexItem struct {
	item       tview.Primitive
	fixedSize  int
	proportion int
	focus      bool
}

// NewFlex returns a new index-aware flex.
func NewFlex() *Flex {
	return &Flex{Flex: tview.NewFlex()}
}

// SetDirection sets the flex direction.
func (f *Flex) SetDirection(direction int) *Flex {
	f.Flex.SetDirection(direction)
	return f
}

// AddItem appends an item to the flex.
func (f *Flex) AddItem(item tview.Primitive, fixedSize, proportion int, focus bool) *Flex {
	f.items = append(f.items, flexItem{item, fixedSize, proportion, focus})
	f.Flex.AddItem(item, fixedSize, proportion, focus)
	return f
}

// AddItemAtIndex inserts an item at the given index.
func (f *Flex) AddItemAtIndex(index int, item tview.Primitive, fixedSize, proportion int, focus bool) *Flex {
	it := flexItem{item, fixedSize, proportion, focus}
	if index < 0 {
		index = 0
	}
	if index >= len(f.items) {
		f.items = append(f.items, it)
	} else {
		f.items = append(f.items[:index], append([]flexItem{it}, f.items[index:]...)...)
	}
	f.rebuild()
	return f
}

// RemoveItem removes the given item from the flex.
func (f *Flex) RemoveItem(p tview.Primitive) *Flex {
	dst := f.items[:0]
	for _, it := range f.items {
		if it.item != p {
			dst = append(dst, it)
		}
	}
	f.items = dst
	f.rebuild()
	return f
}

// RemoveItemAtIndex removes the item at the given index.
func (f *Flex) RemoveItemAtIndex(index int) *Flex {
	if index < 0 || index >= len(f.items) {
		return f
	}
	f.items = append(f.items[:index], f.items[index+1:]...)
	f.rebuild()
	return f
}

// ItemAt returns the item at the given index, or nil if out of range.
func (f *Flex) ItemAt(index int) tview.Primitive {
	if index < 0 || index >= len(f.items) {
		return nil
	}
	return f.items[index].item
}

// GetItemCount returns the number of items.
func (f *Flex) GetItemCount() int {
	return len(f.items)
}

// Clear removes all items.
func (f *Flex) Clear() *Flex {
	f.items = nil
	f.Flex.Clear()
	return f
}

func (f *Flex) rebuild() {
	f.Flex.Clear()
	for _, it := range f.items {
		f.Flex.AddItem(it.item, it.fixedSize, it.proportion, it.focus)
	}
}
