// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"fmt"
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestScenarioDetailsSearch searches a YAML details view through the prompt
// and steps through the matches, which the text view highlights as regions.
func TestScenarioDetailsSearch(t *testing.T) {
	s := newScenario(t)
	d := NewDetails(s.app, "Describe", "ns1/fred", contentYAML, true)
	s.pushComponent(d)
	d.Update("apiVersion: v1\nkind: Pod\nmetadata:\n  name: fred\n  labels:\n    app: fred\n")
	s.draw()
	require.True(t, d.HasFocus(), "pushed details must hold focus")

	s.pressRune('/')
	require.True(t, s.app.Prompt().HasFocus(), "'/' must focus the prompt")
	s.typeText("fred")
	s.pressKey(tcell.KeyEnter)

	assert.True(t, d.HasFocus(), "accepting the search must hand focus back")
	assert.Equal(t, 2, d.maxRegions)
	assert.Equal(t, 0, d.currentRegion)
	s.assertSnapshot("scenario_details_search_first")

	s.pressRune('n')
	assert.Equal(t, 1, d.currentRegion, "n")
	s.assertSnapshot("scenario_details_search_next")

	s.pressRune('N')
	assert.Equal(t, 0, d.currentRegion, "N")
}

// logComponent pushes a log view without starting it: Log.Start tails the pod
// through the connection factory, which scenarios run without. Scenarios feed
// the lines directly instead.
type logComponent struct {
	*Log
}

func (logComponent) Start() {}

// TestScenarioLogNavigation fills a log view and scrolls through it with the
// keyboard.
func TestScenarioLogNavigation(t *testing.T) {
	s := newScenario(t)
	v := NewLog(client.PodGVR, &dao.LogOptions{Path: "fred/p1", Container: "blee"})
	s.pushComponent(logComponent{v})
	require.True(t, v.Logs().HasFocus(), "the log text must hold focus")

	// Real pod log lines carry an RFC3339 timestamp first, which rendering
	// without timestamps strips.
	items := dao.NewLogItems()
	for i := range 100 {
		items.Add(dao.NewLogItemFromString(fmt.Sprintf("2018-12-14T10:36:43.326972-07:00 line-%02d\n", i)))
	}
	ll := make([][]byte, items.Len())
	items.Lines(0, false, ll)
	// Like LogChanged: replace the waiting placeholder with the first lines.
	v.Logs().Clear()
	v.Flush(ll)
	s.draw()

	s.pressKey(tcell.KeyHome)
	row, _ := v.Logs().GetScrollOffset()
	assert.Equal(t, 0, row, "Home")
	s.assertSnapshot("scenario_log_home")

	s.pressKey(tcell.KeyPgDn)
	row, _ = v.Logs().GetScrollOffset()
	assert.Positive(t, row, "PgDn")

	s.pressKey(tcell.KeyEnd)
	s.assertSnapshot("scenario_log_end")
}
