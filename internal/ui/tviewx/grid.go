// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package tviewx

import "github.com/rivo/tview"

// GridItem mirrors the grid item that derailed/tview exposed via Grid.GetItem.
// rivo/tview keeps its grid items unexported, so this wrapper tracks them.
type GridItem struct {
	Item                        tview.Primitive
	Row, Column                 int
	Width, Height               int
	MinGridWidth, MinGridHeight int
	Focus                       bool
}

// Grid wraps tview.Grid and restores Grid.GetItem, which k9s' Pulse view uses to
// track per-cell focus. Actual focus is driven by the application (SetFocus); the
// GridItem.Focus flag is k9s' own bookkeeping, faithful to the fork's behavior.
type Grid struct {
	*tview.Grid

	items []*GridItem
}

// NewGrid returns a new item-tracking grid.
func NewGrid() *Grid {
	return &Grid{Grid: tview.NewGrid()}
}

// AddItem adds a primitive to the grid and records it.
func (g *Grid) AddItem(p tview.Primitive, row, column, rowSpan, colSpan, minGridHeight, minGridWidth int, focus bool) *Grid {
	g.items = append(g.items, &GridItem{
		Item:          p,
		Row:           row,
		Column:        column,
		Height:        rowSpan,
		Width:         colSpan,
		MinGridHeight: minGridHeight,
		MinGridWidth:  minGridWidth,
		Focus:         focus,
	})
	g.Grid.AddItem(p, row, column, rowSpan, colSpan, minGridHeight, minGridWidth, focus)
	return g
}

// GetItem returns the tracked grid item at index i.
func (g *Grid) GetItem(i int) *GridItem {
	return g.items[i]
}
