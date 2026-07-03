// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package tviewx

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ModalForm implements a modal window wrapping a custom form. Ported from
// derailed/tview onto the public rivo/tview API.
type ModalForm struct {
	*Modal
}

// NewModalForm returns a modal that hosts a custom form.
func NewModalForm(title string, form *tview.Form) *ModalForm {
	m := ModalForm{NewModal()}
	m.form = form
	m.form.SetBackgroundColor(tview.Styles.ContrastBackgroundColor).SetBorderPadding(0, 0, 0, 0)
	// rivo/tview only moves between form elements on Tab/Backtab, whereas the
	// derailed fork also honored the arrow keys. Restore that by translating
	// arrows into Tab/Backtab, sparing only the keys an item needs for itself:
	// an open drop-down list navigates with Up/Down, a closed one still opens
	// on Down, and input fields keep Left/Right for cursor movement.
	backtab := func() *tcell.EventKey { return tcell.NewEventKey(tcell.KeyBacktab, 0, tcell.ModNone) }
	tab := func() *tcell.EventKey { return tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone) }
	m.form.SetInputCapture(func(evt *tcell.EventKey) *tcell.EventKey {
		item, button := m.form.GetFocusedItemIndex()
		switch {
		case button >= 0:
			switch evt.Key() {
			case tcell.KeyUp, tcell.KeyLeft:
				return backtab()
			case tcell.KeyDown, tcell.KeyRight:
				return tab()
			}
		case item >= 0:
			switch fi := m.form.GetFormItem(item).(type) {
			case *tview.DropDown:
				if !fi.IsOpen() && evt.Key() == tcell.KeyUp {
					return backtab()
				}
			case *tview.Checkbox:
				switch evt.Key() {
				case tcell.KeyUp, tcell.KeyLeft:
					return backtab()
				case tcell.KeyDown, tcell.KeyRight:
					return tab()
				}
			default:
				switch evt.Key() {
				case tcell.KeyUp:
					return backtab()
				case tcell.KeyDown:
					return tab()
				}
			}
		}
		return evt
	})
	m.form.SetCancelFunc(func() {
		if m.done != nil {
			m.done(-1, "")
		}
	})
	m.frame = tview.NewFrame(m.form).SetBorders(0, 0, 1, 0, 0, 0)
	m.frame.SetBorder(true).
		SetBackgroundColor(tview.Styles.ContrastBackgroundColor).
		SetBorderPadding(1, 1, 1, 1)
	m.frame.SetTitle(title)
	m.frame.SetTitleColor(tcell.ColorAqua)

	return &m
}

// Draw draws this primitive onto the screen.
func (m *ModalForm) Draw(screen tcell.Screen) {
	buttonsWidth := 0
	for i := range m.form.GetButtonCount() {
		buttonsWidth += tview.TaggedStringWidth(m.form.GetButton(i).GetLabel()) + 4 + 2
	}
	buttonsWidth -= 2
	screenWidth, screenHeight := screen.Size()
	width := screenWidth / 3
	if width < buttonsWidth {
		width = buttonsWidth
	}

	m.frame.Clear()
	lines := tview.WordWrap(m.text, width)
	for _, line := range lines {
		m.frame.AddText(line, true, tview.AlignCenter, m.textColor)
	}

	height := len(lines) + m.form.GetFormItemCount() + m.form.GetButtonCount() + 5
	width += 4
	x := (screenWidth - width) / 2
	y := (screenHeight - height) / 2
	m.SetRect(x, y, width, height)

	m.frame.SetRect(x, y, width, height)
	m.frame.Draw(screen)
}
