// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui

import (
	"fmt"
	"testing"

	"github.com/derailed/k9s/internal/ui/uitest"
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

// TestAsKeyControlBandPrivate verifies that no printable rune dispatches into
// tcell's control-key band (KeyCtrlSpace..KeyCtrlUnderscore, i.e. 64-95 since
// tcell v2.10). The band's ASCII characters — '@', 'A'-'Z', '[', '\', ']', '^'
// and '_' — are mirrored into k9s' private range instead; otherwise typing '@'
// would fire Ctrl-Space or typing '\' would fire Ctrl-\.
func TestAsKeyControlBandPrivate(t *testing.T) {
	seen := map[tcell.Key]rune{}
	for r := rune('@'); r <= '_'; r++ {
		k := AsKey(tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone))
		assert.False(t, k >= tcell.KeyCtrlSpace && k <= tcell.KeyCtrlUnderscore,
			"%q dispatches into tcell's control band as %d (%s)", r, k, tcell.KeyNames[k])
		if prev, ok := seen[k]; ok {
			t.Errorf("%q and %q dispatch to the same key %d", prev, r, k)
		}
		seen[k] = r
	}

	assert.Equal(t, KeyShift2, AsKey(tcell.NewEventKey(tcell.KeyRune, '@', tcell.ModNone)))
	assert.Equal(t, KeyShift6, AsKey(tcell.NewEventKey(tcell.KeyRune, '^', tcell.ModNone)))
	assert.Equal(t, KeyLeftBracket, AsKey(tcell.NewEventKey(tcell.KeyRune, '[', tcell.ModNone)))
	assert.Equal(t, KeyRightBracket, AsKey(tcell.NewEventKey(tcell.KeyRune, ']', tcell.ModNone)))
}

// TestAsKeyCtrlPunctuationProtocols verifies that Ctrl+<punctuation> dispatches
// to the same key regardless of the keyboard protocol the terminal speaks.
// tcell enables kitty CSI-u and xterm modifyOtherKeys on xterm-like terminals;
// both report these chords as a rune with ModCtrl, which tcell only normalizes
// for letters. Legacy terminals send the raw control byte instead.
func TestAsKeyCtrlPunctuationProtocols(t *testing.T) {
	uu := map[string]struct {
		r      rune
		legacy byte
		key    tcell.Key
	}{
		"space":     {' ', 0x00, tcell.KeyCtrlSpace},
		"at":        {'@', 0x00, tcell.KeyCtrlSpace},
		"leftSq":    {'[', 0x1b, tcell.KeyEscape},
		"backslash": {'\\', 0x1c, tcell.KeyCtrlBackslash},
		"rightSq":   {']', 0x1d, tcell.KeyCtrlRightSq},
		"carat":     {'^', 0x1e, tcell.KeyCtrlCarat},
		"underline": {'_', 0x1f, tcell.KeyCtrlUnderscore},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			for proto, seq := range map[string]string{
				"legacy":          string([]byte{u.legacy}),
				"kitty":           fmt.Sprintf("\x1b[%d;5u", u.r),
				"modifyOtherKeys": fmt.Sprintf("\x1b[27;5;%d~", u.r),
			} {
				evt := uitest.ScanKey(t, seq)
				assert.Equal(t, u.key, AsKey(evt), "%s: Ctrl+%q", proto, u.r)
			}
		})
	}

	assert.Equal(t, tcell.Key(KeySpace), AsKey(tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone)))
}

// TestAsKeyCtrlModifierCombos verifies that extra modifiers on a Ctrl chord do
// not demote it to the bare letter. A legacy terminal cannot tell Ctrl+Shift+D
// or Ctrl+Alt+D from Ctrl+D; kitty and modifyOtherKeys report them as a rune
// with several modifiers, which must not dispatch as 'd' (describe) or
// Shift-D instead of Ctrl-D (delete).
func TestAsKeyCtrlModifierCombos(t *testing.T) {
	uu := map[string]string{
		"legacy ctrl":           "\x04",
		"kitty ctrl":            "\x1b[100;5u",
		"kitty ctrl+shift":      "\x1b[100;6u",
		"kitty ctrl+alt":        "\x1b[100;7u",
		"modifyOtherKeys shift": "\x1b[27;6;68~",
	}

	for name, seq := range uu {
		assert.Equal(t, tcell.KeyCtrlD, AsKey(uitest.ScanKey(t, seq)), name)
	}
}

// TestAsKeyUnbindable verifies that key events k9s cannot bind do not alias a
// bound key: non-ASCII runes, Alt chords on printable runes, and Ctrl chords on
// runes without a legacy control code. Their raw values would otherwise land
// in k9s' private key range, on tcell's special keys (256 and up) or, once
// truncated to int16, on the control keys.
func TestAsKeyUnbindable(t *testing.T) {
	for name, r := range map[string]rune{
		"cyrillic el":  'Л',
		"cyrillic en":  'Н',
		"cyrillic a":   'А',
		"a macron":     'ā',
		"beyond int16": 0x10041,
	} {
		assert.Equal(t, tcell.KeyRune, AsKey(tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone)), name)
	}

	for name, seq := range map[string]string{
		"kitty ctrl+shift+2": "\x1b[50;6u",
		"kitty ctrl+?":       "\x1b[63;5u",
		"legacy alt+a":       "\x1ba",
		"legacy alt+shift+a": "\x1bA",
		"kitty alt+a":        "\x1b[97;3u",
	} {
		assert.Equal(t, tcell.KeyRune, AsKey(uitest.ScanKey(t, seq)), name)
	}
}

// TestKeyNamesNotClobbered verifies that tcell's control-key mnemonics survive
// k9s' own key name registration. The mnemonics back both the menu hints and
// the hotkey/plugin config round-trip, so no k9s key may share an integer with
// a tcell control key.
func TestKeyNamesNotClobbered(t *testing.T) {
	for k := tcell.KeyCtrlSpace; k <= tcell.KeyCtrlUnderscore; k++ {
		assert.Regexp(t, `^Ctrl-`, tcell.KeyNames[k], "key %d", k)
	}
	assert.Equal(t, "Ctrl-Space", tcell.KeyNames[tcell.KeyCtrlSpace])
	assert.Equal(t, "Shift-A", tcell.KeyNames[KeyShiftA])
	assert.Equal(t, "Shift-S", tcell.KeyNames[KeyShiftS])
	assert.Equal(t, "Shift-2", tcell.KeyNames[KeyShift2])
	assert.Equal(t, "Shift-6", tcell.KeyNames[KeyShift6])
}
