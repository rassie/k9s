// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dialog

import (
	"context"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/ui/tviewx"
	"github.com/rivo/tview"
)

type promptAction func(ctx context.Context)

// ShowPrompt pops a prompt dialog.
func ShowPrompt(styles *config.Dialog, pages *ui.Pages, title, msg string, action promptAction, cancel cancelFunc) {
	f := tview.NewForm()
	f.SetItemPadding(0)
	f.SetButtonsAlign(tview.AlignCenter).
		SetLabelColor(styles.LabelFgColor.Color()).
		SetFieldTextColor(styles.FieldFgColor.Color())
	StyleFormButtons(f, styles)

	ctx, cancelCtx := context.WithCancel(context.Background())

	f.AddButton("Cancel", func() {
		dismiss(pages)
		cancelCtx()
		cancel()
	})

	f.SetFocus(0)
	modal := tviewx.NewModalForm("<"+title+">", f)
	modal.SetText(msg)
	modal.SetTextColor(styles.FgColor.Color())
	modal.SetDoneFunc(func(int, string) {
		dismiss(pages)
		cancelCtx()
		cancel()
	})

	pages.AddPage(dialogKey, modal, false, false)
	pages.ShowPage(dialogKey)

	go func() {
		action(ctx)
		dismiss(pages)
	}()
}
