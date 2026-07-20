// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package tviewx_test

import (
	"strings"
	"testing"

	"github.com/derailed/k9s/internal/ui/tviewx"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func translate(t *testing.T, in ...string) string {
	t.Helper()
	var sb strings.Builder
	w := tviewx.ANSIWriter(&sb)
	for _, s := range in {
		n, err := w.Write([]byte(s))
		require.NoError(t, err)
		require.Equal(t, len(s), n)
	}
	return sb.String()
}

func TestAnsiWriter(t *testing.T) {
	uu := map[string]struct {
		in, e string
	}{
		"plain": {
			in: "no escapes at all",
			e:  "no escapes at all",
		},
		"basic-color": {
			in: "\x1b[36mcyan\x1b[39mplain",
			e:  "[teal]cyan[-]plain",
		},
		"fg-bg": {
			in: "\x1b[31;44mtext",
			e:  "[maroon:navy]text",
		},
		"underline-off-explicit": {
			in: "a\x1b[4mb\x1b[24mc",
			e:  "a[::u]b[::U]c",
		},
		"underline-off-via-reset": {
			in: "x\x1b[4;36my\x1b[0mz",
			e:  "x[teal::u]y[-::U]z",
		},
		"underline-colon-styles": {
			in: "\x1b[4:3mcurly\x1b[4:0moff",
			e:  "[::u]curly[::U]off",
		},
		"bold-dim-off": {
			in: "\x1b[1;2ma\x1b[22mb",
			e:  "[::bd]a[::BD]b",
		},
		"eight-bit": {
			in: "\x1b[38;5;209mX",
			e:  "[#ff6633]X",
		},
		"true-color": {
			in: "\x1b[38;2;1;2;3mX",
			e:  "[#010203]X",
		},
		"true-color-colon-itu": {
			in: "\x1b[38:2::10:20:30mX",
			e:  "[#0a141e]X",
		},
		"attrs-after-extended-color": {
			in: "\x1b[38;5;209;4mX\x1b[0mY",
			e:  "[#ff6633::u]X[-::U]Y",
		},
		"redundant-sgr-emits-nothing": {
			in: "\x1b[39ma\x1b[24mb",
			e:  "ab",
		},
		"osc-hyperlink-swallowed": {
			in: "\x1b]8;;http://example.com\x07link\x1b]8;;\x07 done",
			e:  "link done",
		},
		"osc-st-terminated": {
			in: "\x1b]0;title\x1b\\after",
			e:  "after",
		},
		"full-reset-ris": {
			in: "\x1b[4;31mu\x1bcplain",
			e:  "[maroon::u]u[-::U]plain",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, translate(t, u.in))
		})
	}
}

func TestAnsiWriterStateAcrossWrites(t *testing.T) {
	// Attribute state must persist across Write calls: the off-switch arrives
	// in a later write (k9s writes one log line per call).
	got := translate(t, "\x1b[4munder\n", "still\x1b[0m plain\n")
	assert.Equal(t, "[::u]under\nstill[::U] plain\n", got)
}

// TestAnsiWriterUnderlineNotSticky renders through a real TextView onto a
// simulation screen and asserts the underline actually turns off. This is the
// regression test for the "everything underlined" log view: tview.ANSIWriter's
// SGR-0 translation clears only the underline attribute bit, not the
// tcell.UnderlineStyle that rendering is keyed on, so one ESC[4m in a log
// stream underlined everything after it.
func TestAnsiWriterUnderlineNotSticky(t *testing.T) {
	s := tcell.NewSimulationScreen("UTF-8")
	require.NoError(t, s.Init())
	defer s.Fini()
	s.SetSize(80, 10)

	tv := tview.NewTextView()
	tv.SetDynamicColors(true).SetWrap(false).SetRect(0, 0, 80, 10)

	w := tviewx.ANSIWriter(tv)
	_, err := w.Write([]byte("with \x1b[4munderline\x1b[0m tail\nnext line\n"))
	require.NoError(t, err)

	tv.Draw(s)
	s.Show()

	cells, cw, _ := s.GetContents()
	underlined := func(row, col int) bool {
		c := cells[row*cw+col]
		_, _, attrs := c.Style.Decompose()
		return attrs&tcell.AttrUnderline != 0 ||
			c.Style.GetUnderlineStyle() != tcell.UnderlineStyleNone
	}

	assert.False(t, underlined(0, 0), "prefix must not be underlined")
	assert.True(t, underlined(0, 5), "SGR 4 span must be underlined")
	assert.False(t, underlined(0, 15), "tail after SGR 0 must not be underlined")
	assert.False(t, underlined(1, 0), "next line must not be underlined")
}
