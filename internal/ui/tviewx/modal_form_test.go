// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package tviewx_test

import (
	"testing"

	"github.com/derailed/k9s/internal/ui/tviewx"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
)

// TestModalFormArrowNav verifies that arrow keys navigate between buttons like
// the retired derailed/tview fork did: while a button holds focus, the arrows
// are rewritten to Tab/Backtab. rivo/tview only navigates on Tab/Backtab out of
// the box (gdamore-tcell migration regression).
func TestModalFormArrowNav(t *testing.T) {
	form := tview.NewForm()
	form.AddButton("Cancel", nil)
	form.AddButton("OK", nil)
	tviewx.NewModalForm("<Test>", form)

	app := tview.NewApplication()
	app.SetRoot(form, true)
	app.SetFocus(form)

	// Sanity: a button holds focus.
	if _, button := form.GetFocusedItemIndex(); button < 0 {
		t.Fatalf("expected a button to hold focus, got item index %d", button)
	}

	capture := form.GetInputCapture()
	assert.NotNil(t, capture)

	uu := map[string]struct {
		in   tcell.Key
		want tcell.Key
	}{
		"down->tab":       {in: tcell.KeyDown, want: tcell.KeyTab},
		"right->tab":      {in: tcell.KeyRight, want: tcell.KeyTab},
		"up->backtab":     {in: tcell.KeyUp, want: tcell.KeyBacktab},
		"left->backtab":   {in: tcell.KeyLeft, want: tcell.KeyBacktab},
		"enter passthru":  {in: tcell.KeyEnter, want: tcell.KeyEnter},
		"escape passthru": {in: tcell.KeyEscape, want: tcell.KeyEscape},
	}
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			out := capture(tcell.NewEventKey(u.in, 0, tcell.ModNone))
			assert.NotNil(t, out)
			assert.Equal(t, u.want, out.Key())
		})
	}
}

// TestModalFormArrowNavCheckbox verifies that a focused checkbox no longer
// swallows the arrow keys: like in the derailed fork, they move focus to the
// neighboring form elements.
func TestModalFormArrowNavCheckbox(t *testing.T) {
	form := tview.NewForm()
	form.AddCheckbox("Force:", false, func(bool) {})
	form.AddButton("OK", nil)
	tviewx.NewModalForm("<Test>", form)

	app := tview.NewApplication()
	app.SetRoot(form, true)
	app.SetFocus(form)

	// The checkbox (form item 0) holds focus, not a button.
	item, button := form.GetFocusedItemIndex()
	assert.GreaterOrEqual(t, item, 0)
	assert.Equal(t, -1, button)

	capture := form.GetInputCapture()
	uu := map[string]struct {
		in   tcell.Key
		want tcell.Key
	}{
		"down->tab":     {in: tcell.KeyDown, want: tcell.KeyTab},
		"right->tab":    {in: tcell.KeyRight, want: tcell.KeyTab},
		"up->backtab":   {in: tcell.KeyUp, want: tcell.KeyBacktab},
		"left->backtab": {in: tcell.KeyLeft, want: tcell.KeyBacktab},
		"space toggles": {in: tcell.KeyRune, want: tcell.KeyRune},
	}
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			out := capture(tcell.NewEventKey(u.in, ' ', tcell.ModNone))
			assert.NotNil(t, out)
			assert.Equal(t, u.want, out.Key())
		})
	}
}

// TestModalFormArrowNavInputField verifies that a focused input field keeps
// Left/Right for cursor movement while Up/Down move between form elements.
func TestModalFormArrowNavInputField(t *testing.T) {
	form := tview.NewForm()
	form.AddInputField("Name:", "blee", 20, nil, nil)
	form.AddButton("OK", nil)
	tviewx.NewModalForm("<Test>", form)

	app := tview.NewApplication()
	app.SetRoot(form, true)
	app.SetFocus(form)

	capture := form.GetInputCapture()
	uu := map[string]struct {
		in   tcell.Key
		want tcell.Key
	}{
		"down->tab":      {in: tcell.KeyDown, want: tcell.KeyTab},
		"up->backtab":    {in: tcell.KeyUp, want: tcell.KeyBacktab},
		"left passthru":  {in: tcell.KeyLeft, want: tcell.KeyLeft},
		"right passthru": {in: tcell.KeyRight, want: tcell.KeyRight},
	}
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			out := capture(tcell.NewEventKey(u.in, 0, tcell.ModNone))
			assert.NotNil(t, out)
			assert.Equal(t, u.want, out.Key())
		})
	}
}

// TestModalFormArrowNavDropDown verifies drop-down arrow handling: while
// closed, Up moves to the previous element but Down still opens the option
// list; while open, both arrows pass through so the list can be navigated.
func TestModalFormArrowNavDropDown(t *testing.T) {
	form := tview.NewForm()
	form.AddDropDown("Propagation:", []string{"Background", "Foreground"}, 0, nil)
	form.AddButton("OK", nil)
	tviewx.NewModalForm("<Test>", form)

	app := tview.NewApplication()
	app.SetRoot(form, true)
	app.SetFocus(form)

	dd, ok := form.GetFormItem(0).(*tview.DropDown)
	assert.True(t, ok)
	capture := form.GetInputCapture()

	out := capture(tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone))
	assert.Equal(t, tcell.KeyBacktab, out.Key(), "Up on a closed drop-down must move to the previous element")
	out = capture(tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone))
	assert.Equal(t, tcell.KeyDown, out.Key(), "Down must reach the closed drop-down so it can open its list")

	// Open the list and ensure the arrows now pass through for navigation.
	dd.InputHandler()(tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone), func(tview.Primitive) {})
	assert.True(t, dd.IsOpen())
	out = capture(tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone))
	assert.Equal(t, tcell.KeyUp, out.Key(), "arrows must pass through to an open drop-down list")
}
