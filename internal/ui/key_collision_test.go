// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
)

// TestAsKeyCtrlShiftDistinct guards the invariant that was broken when tcell
// (v2.10+) relocated KeyCtrlA..KeyCtrlZ onto the ASCII codepoints of 'A'-'Z'
// (65-90): Ctrl+<letter> and Shift+<letter> must dispatch to distinct keys.
func TestAsKeyCtrlShiftDistinct(t *testing.T) {
	for r := rune('A'); r <= 'Z'; r++ {
		ctrlKey := tcell.KeyCtrlA + tcell.Key(r-'A')
		ctrl := AsKey(tcell.NewEventKey(ctrlKey, 0, tcell.ModCtrl))
		shift := AsKey(tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone))
		lower := AsKey(tcell.NewEventKey(tcell.KeyRune, r+('a'-'A'), tcell.ModNone))

		assert.Equal(t, ctrlKey, ctrl, "Ctrl+%c must keep its tcell key value", r)
		assert.Equal(t, KeyShiftA+tcell.Key(r-'A'), shift, "Shift+%c must map into the private range", r)
		assert.NotEqual(t, ctrl, shift, "Ctrl+%c and Shift+%c must be distinct", r, r)
		assert.NotEqual(t, shift, lower, "Shift+%c and plain %c must be distinct", r, r+('a'-'A'))
	}
}

// TestKeyNamesNotClobbered verifies that k9s no longer overwrites tcell's own
// Ctrl-letter mnemonics — which it did while Shift+<letter> shared their integer.
// The mnemonics back both the menu hints and the hotkey/plugin config round-trip.
func TestKeyNamesNotClobbered(t *testing.T) {
	assert.Equal(t, "Ctrl-A", tcell.KeyNames[tcell.KeyCtrlA])
	assert.Equal(t, "Ctrl-S", tcell.KeyNames[tcell.KeyCtrlS])
	assert.Equal(t, "Shift-A", tcell.KeyNames[KeyShiftA])
	assert.Equal(t, "Shift-S", tcell.KeyNames[KeyShiftS])
	assert.NotEqual(t, tcell.KeyCtrlS, KeyShiftS)
}
