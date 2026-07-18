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
// no lasting effect. Note that the buttonFgColor/buttonBgColor skin settings
// take effect for the first time here: the derailed/tview fork never applied
// form-level button colors, so unfocused buttons used to render with the
// default foreground on the modal background regardless of the skin.
func StyleFormButtons(f *tview.Form, styles *config.Dialog) *tview.Form {
	f.SetButtonStyle(tcell.StyleDefault.
		Foreground(styles.ButtonFgColor.Color()).
		Background(styles.ButtonBgColor.Color()))
	f.SetButtonActivatedStyle(tcell.StyleDefault.
		Foreground(styles.ButtonFocusFgColor.Color()).
		Background(styles.ButtonFocusBgColor.Color()))

	return f
}
