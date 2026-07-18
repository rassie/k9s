// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dialog

import (
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/ui/tviewx"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ShowError pops an error dialog.
func ShowError(styles *config.Dialog, pages *ui.Pages, msg string) {
	f := tview.NewForm()
	f.SetItemPadding(0)
	f.SetButtonsAlign(tview.AlignCenter).
		SetLabelColor(styles.LabelFgColor.Color()).
		SetFieldTextColor(tcell.ColorIndianRed)
	StyleFormButtons(f, styles)
	f.AddButton("Dismiss", func() {
		dismiss(pages)
	})
	f.SetFocus(0)
	modal := tviewx.NewModalForm("<error>", f)
	modal.SetText(cowTalk(msg))
	modal.SetTextColor(tcell.ColorOrangeRed)
	modal.SetDoneFunc(func(int, string) {
		dismiss(pages)
	})
	pages.AddPage(dialogKey, modal, false, false)
	pages.ShowPage(dialogKey)
}

func cowTalk(says string) string {
	msg := fmt.Sprintf("< Ruroh? %s >", strings.TrimSuffix(says, "\n"))
	buff := make([]string, 0, len(cow)+3)
	buff = append(buff, msg)
	buff = append(buff, cow...)

	return strings.Join(buff, "\n")
}

var cow = []string{
	`\   ^__^            `,
	` \  (oo)\_______    `,
	`    (__)\       )\/\`,
	`        ||----w |   `,
	`        ||     ||   `,
}
