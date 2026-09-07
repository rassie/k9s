// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"testing"
	"time"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config/mock"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/ui/uitest"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/labels"
)

// scenario drives a fully initialized k9s App headlessly and synchronously:
// keys run through the same chain tview's event loop uses (application input
// capture, then the root primitive's input handler), and the UI is drawn after
// every key, because tview draws after each event and several widgets change
// state while drawing (rivo/tview's Table re-walks its selection, TreeView
// fires its changed callback). The event loop itself does not run, so updates
// queued via QueueUpdateDraw never execute and each leaves a goroutine blocked
// until the test binary exits; scenarios must not depend on them.
type scenario struct {
	t        *testing.T
	app      *App
	w, h     int
	revealed bool
}

// pristineTviewStyles is tview's global theme before any App applied a skin.
// Primitives read tview.Styles when they are constructed, and NewApp builds
// some (crumbs, prompt) before its skin reaches tview.Styles, so a scenario
// would otherwise render with whatever skin the previous test left behind.
var pristineTviewStyles = tview.Styles

func newScenario(t *testing.T) *scenario {
	t.Helper()
	tview.Styles = pristineTviewStyles
	t.Cleanup(func() { tview.Styles = pristineTviewStyles })

	cfg := mock.NewMockConfig(t)
	cfg.K9s.UI.Headless = true
	cfg.K9s.UI.Logoless = true
	cfg.K9s.UI.Splashless = true

	a := NewApp(cfg)
	require.NoError(t, a.Init("test", 0))

	return &scenario{t: t, app: a, w: 80, h: 20}
}

func (s *scenario) context() context.Context {
	ctx := context.WithValue(context.Background(), internal.KeyApp, s.app)
	return context.WithValue(ctx, internal.KeyStyles, s.app.Styles)
}

// tableComponent makes a bare Table pushable. k9s pushes Browsers, which wrap
// a Table but need a connection factory to initialize; the table's own key
// bindings are what scenarios exercise, so the resource layer is left out.
type tableComponent struct {
	*Table
}

func (c tableComponent) InCmdMode() bool {
	return c.CmdBuff().InCmdMode()
}

// SetFilter and SetLabelSelector are the programmatic filter API used by
// command navigation; scenarios filter through the prompt instead.
func (tableComponent) SetFilter(string, bool)                 {}
func (tableComponent) SetLabelSelector(labels.Selector, bool) {}

// scenarioTableModel serves fixed rows. Unlike makeTableData its rows carry
// IDs, which marking and selection resolve items by.
type scenarioTableModel struct {
	mockTableModel

	data *model1.TableData
}

func newScenarioTableModel(names ...string) *scenarioTableModel {
	rows := make([]model1.RowEvent, 0, len(names))
	for _, n := range names {
		rows = append(rows, model1.RowEvent{
			Row: model1.Row{ID: "ns1/" + n, Fields: model1.Fields{"ns1", n}},
		})
	}

	return &scenarioTableModel{
		data: model1.NewTableDataWithRows(
			client.NewGVR("test"),
			model1.Header{{Name: "NAMESPACE"}, {Name: "NAME"}},
			model1.NewRowEventsWithEvts(rows...),
		),
	}
}

func (m *scenarioTableModel) Peek() *model1.TableData { return m.data }
func (m *scenarioTableModel) RowCount() int           { return m.data.RowCount() }

// pushTable initializes a table view over m and pushes it onto the content
// stack, which starts and focuses it like k9s' own navigation does.
func (s *scenario) pushTable(gvr *client.GVR, m ui.Tabular) *Table {
	s.t.Helper()
	v := NewTable(gvr)
	require.NoError(s.t, v.Init(s.context()))
	v.SetModel(m)
	v.Refresh()
	s.push(tableComponent{v})

	return v
}

// pushComponent initializes c and pushes it onto the content stack.
func (s *scenario) pushComponent(c model.Component) {
	s.t.Helper()
	require.NoError(s.t, c.Init(s.context()))
	s.push(c)
}

func (s *scenario) push(c model.Component) {
	s.app.Content.Push(c)
	s.reveal()
	s.draw()
}

// reveal shows the main page once, after the first view was pushed. App.Run
// does the same from a queued update: the initial view is focused before the
// page switch, so the switch hands focus down through the layout and the
// content pages learn how to focus pages added later, such as dialogs.
func (s *scenario) reveal() {
	if s.revealed {
		return
	}
	s.app.Main.SwitchToPage("main")
	s.revealed = true
}

func (s *scenario) draw() string {
	return uitest.RenderSnapshot(s.app.Main, s.w, s.h)
}

// press dispatches evt and redraws. A key handler that does not return within
// the timeout fails the test instead of hanging it (see rivo/tview#944).
func (s *scenario) press(evt *tcell.EventKey) {
	s.t.Helper()
	name := evt.Name()
	done := make(chan struct{})
	var dropped bool
	go func() {
		defer close(done)
		if capture := s.app.GetInputCapture(); capture != nil {
			if evt = capture(evt); evt == nil {
				return
			}
		}
		// Like tview's event loop, only a focused root receives keys.
		if dropped = !s.app.Main.HasFocus(); dropped {
			return
		}
		s.app.Main.InputHandler()(evt, func(p tview.Primitive) { s.app.SetFocus(p) })
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		s.t.Fatalf("key %s did not return", name)
	}
	if dropped {
		s.t.Fatalf("key %s was dropped: the root primitive has no focus", name)
	}
	s.draw()
}

func (s *scenario) pressKey(k tcell.Key) {
	s.t.Helper()
	s.press(tcell.NewEventKey(k, 0, tcell.ModNone))
}

func (s *scenario) pressRune(r rune) {
	s.t.Helper()
	s.press(tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone))
}

// pressCtrl presses a control key the way a terminal reports it, with ModCtrl.
func (s *scenario) pressCtrl(k tcell.Key) {
	s.t.Helper()
	s.press(tcell.NewEventKey(k, 0, tcell.ModCtrl))
}

func (s *scenario) typeText(text string) {
	s.t.Helper()
	for _, r := range text {
		s.pressRune(r)
	}
}

func (s *scenario) assertSnapshot(name string) {
	s.t.Helper()
	uitest.AssertSnapshot(s.t, name, s.draw())
}

func TestScenarioTableMarkRange(t *testing.T) {
	s := newScenario(t)
	v := s.pushTable(client.NewGVR("test"), newScenarioTableModel("a", "b", "c", "d"))

	// Without focus keys are dropped silently, and every later assertion
	// would pass against an untouched table.
	require.True(t, v.HasFocus(), "pushed table must hold focus")
	require.Equal(t, 5, v.GetRowCount())
	require.Equal(t, 1, v.GetSelectedRowIndex())

	s.pressRune(' ')
	s.pressKey(tcell.KeyDown)
	s.pressKey(tcell.KeyDown)
	s.pressCtrl(tcell.KeyCtrlSpace)

	assert.Equal(t, 3, v.GetSelectedRowIndex())
	assert.ElementsMatch(t, []string{"ns1/a", "ns1/b", "ns1/c"}, v.GetSelectedItems())
	s.assertSnapshot("scenario_table_mark_range")

	s.pressCtrl(tcell.KeyCtrlBackslash)
	assert.False(t, v.IsMarked("ns1/a"), "Ctrl-\\ must clear all marks")
	assert.Equal(t, []string{"ns1/c"}, v.GetSelectedItems())
}

// TestScenarioEmptyTableNavigation navigates a header-only table, as shown for
// "no resources found" or a filter without matches. rivo/tview spins forever
// on that (rivo/tview#944) once a draw has parked the selection past the last
// row; the table view swallows navigation keys on such a table. Without that
// guard the key times out here, which relies on the scenario drawing first.
func TestScenarioEmptyTableNavigation(t *testing.T) {
	s := newScenario(t)
	v := s.pushTable(client.NewGVR("test"), newScenarioTableModel())
	require.True(t, v.HasFocus(), "pushed table must hold focus")
	require.Equal(t, 1, v.GetRowCount())

	s.pressKey(tcell.KeyDown)
	s.pressRune('j')
	s.pressKey(tcell.KeyEnd)

	s.assertSnapshot("scenario_table_empty_nav")
}

// TestScenarioTableNavigationEdges moves through a table taller than the view:
// jumps to both ends and pages in between, which exercises the widget's
// selection walk and how it scrolls the selection into view.
func TestScenarioTableNavigationEdges(t *testing.T) {
	s := newScenario(t)
	names := make([]string, 30)
	for i := range names {
		names[i] = string([]rune{'a' + rune(i/26), 'a' + rune(i%26)})
	}
	v := s.pushTable(client.NewGVR("test"), newScenarioTableModel(names...))
	require.True(t, v.HasFocus(), "pushed table must hold focus")
	require.Equal(t, 31, v.GetRowCount())

	s.pressKey(tcell.KeyEnd)
	assert.Equal(t, 30, v.GetSelectedRowIndex(), "End")
	s.assertSnapshot("scenario_table_nav_end")

	s.pressKey(tcell.KeyHome)
	assert.Equal(t, 1, v.GetSelectedRowIndex(), "Home")

	s.pressRune('G')
	assert.Equal(t, 30, v.GetSelectedRowIndex(), "G")

	s.pressRune('g')
	assert.Equal(t, 1, v.GetSelectedRowIndex(), "g")

	s.pressKey(tcell.KeyPgDn)
	paged := v.GetSelectedRowIndex()
	assert.Greater(t, paged, 1, "PgDn")
	s.assertSnapshot("scenario_table_nav_pgdn")

	s.pressKey(tcell.KeyPgUp)
	assert.Equal(t, 1, v.GetSelectedRowIndex(), "PgUp after PgDn")
}

// TestScenarioCommandPrompt opens the command prompt and cancels it again. The
// prompt is spliced into the main layout while active, and canceling it must
// hand focus back to the view underneath. Typing is left out: command
// suggestions dereference the connection factory, which scenarios run without.
func TestScenarioCommandPrompt(t *testing.T) {
	s := newScenario(t)
	v := s.pushTable(client.NewGVR("test"), newScenarioTableModel("a", "b"))

	s.pressRune(':')
	require.True(t, s.app.Prompt().HasFocus(), "':' must focus the prompt")
	assert.True(t, s.app.InCmdMode())
	s.assertSnapshot("scenario_prompt_open")

	s.pressKey(tcell.KeyEscape)
	assert.False(t, s.app.InCmdMode())
	assert.True(t, v.HasFocus(), "Esc must hand focus back to the table")
	s.assertSnapshot("scenario_prompt_closed")
}

// TestScenarioHelpToggle opens the help view on top of a table and closes it
// again, which pushes and pops the content stack and moves focus both ways.
func TestScenarioHelpToggle(t *testing.T) {
	s := newScenario(t)
	v := s.pushTable(client.NewGVR("test"), newScenarioTableModel("a", "b"))

	s.pressRune('?')
	top := s.app.Content.Top()
	require.NotNil(t, top)
	require.Equal(t, "help", top.Name())
	assert.True(t, top.HasFocus(), "help must hold focus")
	s.assertSnapshot("scenario_help_open")

	s.pressKey(tcell.KeyEscape)
	assert.Equal(t, v.Name(), s.app.Content.Top().Name())
	assert.True(t, v.HasFocus(), "closing help must hand focus back to the table")
}
