// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"fmt"
	"strconv"
	"time"

	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/ui/dialog"
	"github.com/derailed/k9s/internal/ui/tviewx"
	"github.com/rivo/tview"
)

const drainKey = "drain"

// DrainFunc represents a drain callback function.
type DrainFunc func(v ResourceViewer, sels []string, opts dao.DrainOptions)

// ShowDrain pops a node drain dialog.
func ShowDrain(view ResourceViewer, sels []string, opts dao.DrainOptions, okFn DrainFunc) {
	styles := view.App().Styles.Dialog()

	f := tview.NewForm().
		SetItemPadding(0).
		SetButtonsAlign(tview.AlignCenter).
		SetLabelColor(styles.LabelFgColor.Color()).
		SetFieldTextColor(styles.FieldFgColor.Color()).
		SetFieldBackgroundColor(styles.BgColor.Color())
	dialog.StyleFormButtons(f, &styles)

	f.AddInputField("GracePeriod:", strconv.Itoa(opts.GracePeriodSeconds), 0, nil, func(v string) {
		a, err := asIntOpt(v)
		if err != nil {
			view.App().Flash().Err(err)
			return
		}
		view.App().Flash().Clear()
		opts.GracePeriodSeconds = a
	})
	f.AddInputField("Timeout:", opts.Timeout.String(), 0, nil, func(v string) {
		a, err := asDurOpt(v)
		if err != nil {
			view.App().Flash().Err(err)
			return
		}
		view.App().Flash().Clear()
		opts.Timeout = a
	})
	f.AddCheckbox("Ignore DaemonSets:", opts.IgnoreAllDaemonSets, func(v bool) {
		opts.IgnoreAllDaemonSets = v
	})
	f.AddCheckbox("Delete EmptyDir Data:", opts.DeleteEmptyDirData, func(v bool) {
		opts.DeleteEmptyDirData = v
	})
	f.AddCheckbox("Force:", opts.Force, func(v bool) {
		opts.Force = v
	})
	f.AddCheckbox("Disable Eviction:", opts.DisableEviction, func(v bool) {
		opts.DisableEviction = v
	})

	pages := view.App().Content.Pages
	f.AddButton("Cancel", func() {
		DismissDrain(view, pages)
	})
	f.AddButton("OK", func() {
		DismissDrain(view, pages)
		okFn(view, sels, opts)
	})

	modal := tviewx.NewModalForm("<Drain>", f)
	path := "Drain "
	if len(sels) == 1 {
		path += sels[0]
	} else {
		path += fmt.Sprintf("(%d) nodes", len(sels))
	}
	path += "?"
	modal.SetText(path)
	modal.SetDoneFunc(func(int, string) {
		DismissDrain(view, pages)
	})

	pages.AddPage(drainKey, modal, false, true)
	pages.ShowPage(drainKey)
	view.App().SetFocus(pages.GetPage(drainKey))
}

// DismissDrain dismiss the port forward dialog.
func DismissDrain(v ResourceViewer, p *ui.Pages) {
	p.RemovePage(drainKey)
	_, it := p.GetFrontPage()
	v.App().SetFocus(it)
}

// ----------------------------------------------------------------------------
// Helpers...

func asDurOpt(v string) (time.Duration, error) {
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, err
	}

	return d, nil
}

func asIntOpt(v string) (int, error) {
	i, err := strconv.Atoi(v)
	if err != nil {
		return 0, err
	}

	return i, nil
}
