// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dialog

import (
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/ui/tviewx"
	"github.com/rivo/tview"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type RestartFn func(*metav1.PatchOptions) bool

type RestartDialogOpts struct {
	Title, Message string
	FieldManager   string
	Ack            RestartFn
	Cancel         cancelFunc
}

func ShowRestart(styles *config.Dialog, pages *ui.Pages, opts *RestartDialogOpts) {
	f := tview.NewForm()
	f.SetItemPadding(0)
	f.SetButtonsAlign(tview.AlignCenter).
		SetLabelColor(styles.LabelFgColor.Color()).
		SetFieldTextColor(styles.FieldFgColor.Color())
	StyleFormButtons(f, styles)
	f.AddButton("Cancel", func() {
		dismissConfirm(pages)
		opts.Cancel()
	})

	modal := tviewx.NewModalForm("<"+opts.Title+">", f)

	args := metav1.PatchOptions{
		FieldManager: opts.FieldManager,
	}
	f.AddInputField("FieldManager:", args.FieldManager, 40, nil, func(v string) {
		args.FieldManager = v
	})

	f.AddButton("OK", func() {
		if !opts.Ack(&args) {
			return
		}
		dismissConfirm(pages)
		opts.Cancel()
	})
	f.SetFocus(1)

	message := opts.Message
	modal.SetText(message)
	modal.SetTextColor(styles.FgColor.Color())
	modal.SetDoneFunc(func(int, string) {
		dismissConfirm(pages)
		opts.Cancel()
	})
	pages.AddPage(confirmKey, modal, false, false)
	pages.ShowPage(confirmKey)
}
