// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dialog_test

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/ui/dialog"
	"github.com/derailed/k9s/internal/ui/uitest"
	"github.com/rivo/tview"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Render-snapshot tests capture the styled cell grid (colors, attributes) that
// GetText()-based tests cannot see — dialog button styling in particular only
// shows up in the rendered cells. Regenerate with:
// K9S_UPDATE_SNAPSHOTS=1 go test ./internal/ui/dialog/...

// giveFocus hands focus down to the dialog's focused element the way tview's
// Application would, so focus-dependent styles (activated buttons, focused
// fields) are rendered.
func giveFocus(p tview.Primitive) {
	var focus func(tview.Primitive)
	focus = func(pr tview.Primitive) { pr.Focus(focus) }
	focus(p)
}

func TestConfirmDialogSnapshot(t *testing.T) {
	s := config.NewStyles()
	s.Update()
	p := ui.NewPages()
	d := s.Dialog()
	dialog.ShowConfirm(&d, p, "Confirm", "Blow up the pod?", func() {}, func() {})
	giveFocus(p)

	uitest.AssertSnapshot(t, "confirm", uitest.RenderSnapshot(p, 60, 12))
}

func TestDeleteDialogSnapshot(t *testing.T) {
	s := config.NewStyles()
	s.Update()
	p := ui.NewPages()
	d := s.Dialog()
	dialog.ShowDelete(&d, p, "Delete pod fred?", func(*metav1.DeletionPropagation, bool) {}, func() {})
	giveFocus(p)

	uitest.AssertSnapshot(t, "delete", uitest.RenderSnapshot(p, 70, 16))
}

func TestSelectionDialogSnapshot(t *testing.T) {
	s := config.NewStyles()
	s.Update()
	p := ui.NewPages()
	d := s.Dialog()
	dialog.ShowSelection(&d, p, "Container", []string{"nginx", "istio-proxy"}, func(int) {})
	giveFocus(p)

	uitest.AssertSnapshot(t, "selection", uitest.RenderSnapshot(p, 60, 12))
}
