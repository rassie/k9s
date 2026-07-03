// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package tviewx

import "github.com/rivo/tview"

// TextView wraps tview.TextView and restores the cursor API that the
// derailed/tview fork carried (ShowCursor/SetCursorIndex). In that fork the
// cursor state was write-only — it was never rendered — so these are faithfully
// reproduced here as state-tracking no-ops, preserving k9s' existing behavior
// exactly. The fields are retained so a real cursor can be drawn later without
// changing the call sites.
type TextView struct {
	*tview.TextView

	cursorIndex int
	showCursor  bool
}

// NewTextView returns a new cursor-aware text view.
func NewTextView() *TextView {
	return &TextView{TextView: tview.NewTextView(), cursorIndex: 4}
}

// ShowCursor toggles cursor visibility.
func (t *TextView) ShowCursor(f bool) {
	t.showCursor = f
}

// SetCursorIndex tracks the cursor position.
func (t *TextView) SetCursorIndex(i int) {
	t.cursorIndex = i
}
