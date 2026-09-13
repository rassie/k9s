// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view_test

import (
	"fmt"
	"testing"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/ui/uitest"
	"github.com/derailed/k9s/internal/view"
	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// specialKeyEncodings lists the byte sequences xterm-like terminals send for
// the non-printable keys k9s binds, in normal and application cursor mode.
var specialKeyEncodings = map[tcell.Key][]string{
	tcell.KeyEnter:     {"\r"},
	tcell.KeyEscape:    {"\x1b", "\x1b[27u"},
	tcell.KeyTab:       {"\t"},
	tcell.KeyBacktab:   {"\x1b[Z"},
	tcell.KeyBackspace: {"\x7f"},
	tcell.KeyDelete:    {"\x1b[3~"},
	tcell.KeyUp:        {"\x1b[A", "\x1bOA"},
	tcell.KeyDown:      {"\x1b[B", "\x1bOB"},
	tcell.KeyRight:     {"\x1b[C", "\x1bOC"},
	tcell.KeyLeft:      {"\x1b[D", "\x1bOD"},
	tcell.KeyHome:      {"\x1b[H", "\x1bOH", "\x1b[1~"},
	tcell.KeyEnd:       {"\x1b[F", "\x1bOF", "\x1b[4~"},
	tcell.KeyPgUp:      {"\x1b[5~"},
	tcell.KeyPgDn:      {"\x1b[6~"},
}

// unsendableKeys are bound keys that no terminal sends.
var unsendableKeys = map[tcell.Key]bool{
	tcell.KeyHelp: true,
}

// TestKeyBindingsReachable checks every key bound by k9s' main views against
// the byte sequences terminals send for it: the legacy encoding and, for
// control keys, the kitty CSI-u and xterm modifyOtherKeys encodings tcell
// enables on xterm-like terminals. Each must dispatch to the bound key. Every
// bound key also needs a name no other key shares, since menu hints and
// hotkey and plugin configs refer to keys by name.
func TestKeyBindingsReachable(t *testing.T) {
	ctx := makeCtx(t)
	app, ok := ctx.Value(internal.KeyApp).(*view.App)
	require.True(t, ok)

	po, ok := view.NewPod(client.PodGVR).(interface {
		view.ResourceViewer
		GetTable() *view.Table
	})
	require.True(t, ok)
	require.NoError(t, po.Init(ctx))
	app.Content.Push(po)

	help := view.NewHelp(app)
	require.NoError(t, help.Init(ctx))
	details := view.NewDetails(app, "Describe", "fred/p1", "yaml", true)
	require.NoError(t, details.Init(ctx))
	logs := view.NewLog(client.PodGVR, &dao.LogOptions{Path: "fred/p1", Container: "blee"})
	require.NoError(t, logs.Init(ctx))
	// Help only needed the pod view for its hints. Once initialized, the app
	// starts whatever is on top of the content stack, and starting a resource
	// view needs a connection.
	app.Content.Pop()
	require.NoError(t, app.Init("test", 0))

	bindings := map[string]*ui.KeyActions{
		"app":     app.GetActions(),
		"pods":    po.GetTable().Actions(),
		"help":    help.Actions(),
		"details": details.Actions(),
		"logs":    logs.Logs().Actions(),
	}

	printable := map[tcell.Key]rune{}
	for r := rune(' '); r <= '~'; r++ {
		printable[ui.AsKey(tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone))] = r
	}
	nameCount := map[string]int{}
	for _, n := range tcell.KeyNames {
		nameCount[n]++
	}

	for component, aa := range bindings {
		t.Run(component, func(t *testing.T) {
			aa.Range(func(k tcell.Key, a ui.KeyAction) {
				// Nobody can press or usefully configure a key no terminal sends.
				if unsendableKeys[k] {
					return
				}
				name, named := tcell.KeyNames[k]
				if assert.True(t, named, "key %d (%s) has no name", k, a.Description) {
					assert.Equal(t, 1, nameCount[name], "name %q of %s is not unique", name, a.Description)
				}
				seqs := terminalEncodings(k, printable)
				if !assert.NotEmpty(t, seqs, "no terminal sends key %d (%s)", k, a.Description) {
					return
				}
				for _, seq := range seqs {
					assert.Equal(t, k, ui.AsKey(uitest.ScanKey(t, seq)), "%s (%s) sent as %q", name, a.Description, seq)
				}
			})
		})
	}
}

// terminalEncodings returns the byte sequences a terminal may send for k.
func terminalEncodings(k tcell.Key, printable map[tcell.Key]rune) []string {
	if seqs, ok := specialKeyEncodings[k]; ok {
		return seqs
	}
	if r, ok := printable[k]; ok {
		return []string{string(r)}
	}

	var r rune
	switch {
	case k >= tcell.KeyCtrlA && k <= tcell.KeyCtrlZ:
		r = 'a' + rune(k-tcell.KeyCtrlA)
	case k == tcell.KeyCtrlSpace:
		r = ' '
	case k >= tcell.KeyCtrlBackslash && k <= tcell.KeyCtrlUnderscore:
		r = '@' + rune(k-tcell.KeyCtrlSpace)
	default:
		return nil
	}
	legacy := byte(k - tcell.KeyCtrlSpace)

	return []string{
		string([]byte{legacy}),
		fmt.Sprintf("\x1b[%d;5u", r),
		fmt.Sprintf("\x1b[27;5;%d~", r),
	}
}
