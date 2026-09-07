// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"testing"
	"time"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/k9s/internal/ui/uitest"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Navigation keys on a table that renders only its header row (no resources,
// or a filter matching nothing) must return. rivo/tview's Table walks the
// cells looking for a selectable one and only stops when it gets back to the
// previously selected cell; its Draw parks the selection past the last row of
// a header-only table, so that cell is never reached and the key handler
// spins forever (rivo/tview#944). The render step matters: it is Draw that
// moves the selection out of range, so a test without it does not reproduce
// the freeze.
func TestTableEmptyNavigationReturns(t *testing.T) {
	keys := []*tcell.EventKey{
		tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone),
		tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone),
		tcell.NewEventKey(tcell.KeyPgDn, 0, tcell.ModNone),
		tcell.NewEventKey(tcell.KeyPgUp, 0, tcell.ModNone),
		tcell.NewEventKey(tcell.KeyHome, 0, tcell.ModNone),
		tcell.NewEventKey(tcell.KeyEnd, 0, tcell.ModNone),
		tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),
		tcell.NewEventKey(tcell.KeyRune, 'k', tcell.ModNone),
		tcell.NewEventKey(tcell.KeyRune, 'g', tcell.ModNone),
		tcell.NewEventKey(tcell.KeyRune, 'G', tcell.ModNone),
	}
	for _, evt := range keys {
		t.Run(evt.Name(), func(t *testing.T) {
			v := NewTable(client.PodGVR)
			if err := v.Init(makeContext(t)); err != nil {
				t.Fatal(err)
			}
			data := model1.NewTableDataWithRows(
				client.PodGVR,
				model1.Header{{Name: "NAME"}, {Name: "AGE"}},
				model1.NewRowEvents(0),
			)
			cdata := v.Update(data, false)
			v.UpdateUI(cdata, data)
			_ = uitest.RenderSnapshot(v, 80, 20)

			done := make(chan struct{})
			go func() {
				defer close(done)
				v.InputHandler()(evt, func(tview.Primitive) {})
			}()
			select {
			case <-done:
			case <-time.After(2 * time.Second):
				t.Fatalf("%s never returned on a header-only table", evt.Name())
			}
		})
	}
}
