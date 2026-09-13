// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/ui/dialog"
	"github.com/derailed/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// deleteOutcome records which callbacks of the delete dialog fired. The OK
// button calls the cancel callback too, after confirming.
type deleteOutcome struct {
	confirmed, canceled bool
	propagation         *metav1.DeletionPropagation
}

// showDeleteDialog opens the delete dialog over the content pages the way the
// resource views do.
func (s *scenario) showDeleteDialog() *deleteOutcome {
	s.t.Helper()
	out := new(deleteOutcome)
	d := s.app.Styles.Dialog()
	dialog.ShowDelete(&d, s.app.Content.Pages, "Delete a?",
		func(p *metav1.DeletionPropagation, _ bool) { out.confirmed, out.propagation = true, p },
		func() { out.canceled = true },
	)
	s.draw()

	return out
}

// TestScenarioDeleteDialogCancel opens the delete dialog over a table, checks
// that it captures the keyboard, and cancels it with Esc.
func TestScenarioDeleteDialogCancel(t *testing.T) {
	s := newScenario(t)
	v := s.pushTable(client.NewGVR("test"), newScenarioTableModel("a", "b"))

	out := s.showDeleteDialog()
	require.True(t, s.app.Content.IsTopDialog())
	assert.False(t, v.HasFocus(), "the dialog must take focus from the table")
	s.assertSnapshot("scenario_delete_dialog_open")

	// App-wide keys are disabled while a dialog is on top.
	s.pressRune('?')
	assert.NotEqual(t, "help", s.app.Content.Top().Name())

	s.pressKey(tcell.KeyTab)
	s.pressKey(tcell.KeyBacktab)
	assert.True(t, s.app.Content.IsTopDialog(), "moving between the buttons must keep the dialog open")

	s.pressKey(tcell.KeyEscape)
	assert.False(t, out.confirmed, "Esc must not confirm the deletion")
	assert.True(t, out.canceled)
	assert.False(t, s.app.Content.IsTopDialog())
	assert.True(t, v.HasFocus(), "closing the dialog must hand focus back to the table")
}

// TestScenarioDeleteDialogDefaultButton presses Enter right after the delete
// dialog opens. The dialog focuses Cancel, so a reflexive Enter must not delete.
func TestScenarioDeleteDialogDefaultButton(t *testing.T) {
	s := newScenario(t)
	s.pushTable(client.NewGVR("test"), newScenarioTableModel("a", "b"))

	out := s.showDeleteDialog()
	s.pressKey(tcell.KeyEnter)

	assert.False(t, out.confirmed, "Enter on the freshly opened dialog must not confirm the deletion")
	assert.False(t, s.app.Content.IsTopDialog())
}
