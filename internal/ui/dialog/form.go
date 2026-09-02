// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dialog

import (
	"github.com/derailed/k9s/internal/config"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// StyleFormButtons applies the dialog button colors to a form. rivo/tview
// reapplies the form-level button styles to every button on each draw, so the
// focus colors must be set on the form itself; styling individual buttons has
// no lasting effect.
//
// Unfocused buttons deliberately render like they did with the derailed/tview
// fork: primary text on the modal background, so they blend into the dialog and
// only the focused button stands out. The fork never applied form-level button
// colors, so the buttonFgColor/buttonBgColor skin keys have never had an effect;
// honoring them here would paint every button and obscure which one is active.
func StyleFormButtons(f *tview.Form, styles *config.Dialog) *tview.Form {
	f.SetButtonStyle(tcell.StyleDefault.
		Foreground(tview.Styles.PrimaryTextColor).
		Background(tview.Styles.ContrastBackgroundColor))
	f.SetButtonActivatedStyle(tcell.StyleDefault.
		Foreground(styles.ButtonFocusFgColor.Color()).
		Background(styles.ButtonFocusBgColor.Color()))

	return f
}
