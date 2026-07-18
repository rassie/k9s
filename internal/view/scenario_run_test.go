// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"testing"
	"time"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/ui/uitest"
	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/require"
)

// runLoop starts tview's event loop on a simulation screen, so the updates
// views queue through QueueUpdateDraw run the way they do in k9s. App.Run is
// left out: without a connection its default command opens a resource view
// that needs a connection factory. Once the loop runs, scenario state belongs
// to the loop goroutine; read it through inLoop and send keys to the screen.
func (s *scenario) runLoop() tcell.SimulationScreen {
	s.t.Helper()
	screen := tcell.NewSimulationScreen("UTF-8")
	// Not every tview initializes a screen handed to SetScreen before Run.
	require.NoError(s.t, screen.Init())
	screen.SetSize(s.w, s.h)
	s.app.SetScreen(screen)

	done := make(chan error, 1)
	go func() { done <- s.app.Application.Run() }()
	s.t.Cleanup(func() {
		go s.app.Application.Stop()
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			s.t.Error("event loop did not stop")
		}
	})

	return screen
}

// inLoop runs f on the event loop goroutine and waits for it to finish.
func (s *scenario) inLoop(f func()) {
	s.t.Helper()
	done := make(chan struct{})
	go func() {
		s.app.Application.QueueUpdate(f)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		s.t.Fatal("event loop did not run the update")
	}
}

// eventually waits until cond, evaluated on the event loop goroutine, holds.
// Key events and queued updates reach the loop through separate channels, so
// the effect of a key is only observable after an unknown number of updates.
func (s *scenario) eventually(msg string, cond func() bool) {
	s.t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for {
		var ok bool
		s.inLoop(func() { ok = cond() })
		if ok {
			return
		}
		if time.Now().After(deadline) {
			s.t.Fatalf("timed out waiting: %s", msg)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// assertLoopSnapshot renders the UI on the event loop goroutine.
func (s *scenario) assertLoopSnapshot(name string) {
	s.t.Helper()
	var got string
	s.inLoop(func() { got = s.draw() })
	uitest.AssertSnapshot(s.t, name, got)
}

// TestScenarioLoopToggleHeader shows the header with Ctrl-E, which rebuilds
// the main layout from a queued update.
func TestScenarioLoopToggleHeader(t *testing.T) {
	s := newScenario(t)
	s.pushTable(client.NewGVR("test"), newScenarioTableModel("a", "b"))
	screen := s.runLoop()

	screen.InjectKey(tcell.KeyCtrlE, 0, tcell.ModCtrl)
	s.eventually("Ctrl-E shows the header", func() bool { return s.app.showHeader })
	s.assertLoopSnapshot("scenario_loop_header")
}
