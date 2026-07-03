// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui_test

import (
	"context"
	"testing"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/config/mock"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/ui/uitest"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/stretchr/testify/require"
)

// These render-snapshot tests capture the styled cell grid (colors, selection,
// focus), which GetText()-based tests cannot see — e.g. a cursor row whose
// highlight collapses to the terminal default would pass every text assertion.
// Regenerate goldens with: K9S_UPDATE_SNAPSHOTS=1 go test ./internal/ui/...

func TestTableSelectionSnapshot(t *testing.T) {
	v := ui.NewTable(client.NewGVR("fred"))
	v.Init(makeContext())
	m := new(mockModel)
	v.SetModel(m)
	data := m.Peek()
	cdata := v.Update(data, false)
	v.UpdateUI(cdata, data)

	// Color the cells the way a skin does in production so the selected row's
	// highlight has to compose with per-cell colors, not just the defaults.
	for r := range v.GetRowCount() {
		for c := range v.GetColumnCount() {
			if cell := v.GetCell(r, c); cell != nil {
				cell.SetTextColor(tcell.ColorAqua)
			}
		}
	}
	v.SelectRow(1, 0, true) // triggers selectionChanged -> SetSelectedStyle

	uitest.AssertSnapshot(t, "table_selection", uitest.RenderSnapshot(v, 60, 6))
}

func TestLogoSnapshot(t *testing.T) {
	v := ui.NewLogo(config.NewStyles())
	uitest.AssertSnapshot(t, "logo", uitest.RenderSnapshot(v, 30, 7))
}

func TestCrumbsSnapshot(t *testing.T) {
	v := ui.NewCrumbs(config.NewStyles())
	v.StackPushed(makeComponent("c1"))
	v.StackPushed(makeComponent("c2"))
	v.StackPushed(makeComponent("c3"))

	uitest.AssertSnapshot(t, "crumbs", uitest.RenderSnapshot(v, 70, 1))
}

func TestIndicatorSnapshot(t *testing.T) {
	i := ui.NewStatusIndicator(ui.NewApp(mock.NewMockConfig(t), ""), config.NewStyles())
	i.Info("Blee")

	uitest.AssertSnapshot(t, "indicator_info", uitest.RenderSnapshot(i, 30, 1))
}

func TestPromptSnapshot(t *testing.T) {
	v := ui.NewPrompt(nil, false, config.NewStyles())
	m := model.NewFishBuff(':', model.CommandBuffer)
	v.SetModel(m)
	m.AddListener(v)
	m.SetText("po", "ds", false)
	m.SetActive(true)

	uitest.AssertSnapshot(t, "prompt", uitest.RenderSnapshot(v, 40, 3))
}

func TestMenuSnapshot(t *testing.T) {
	v := ui.NewMenu(config.NewStyles())
	v.HydrateMenu(model.MenuHints{
		{Mnemonic: "a", Description: "Attach", Visible: true},
		{Mnemonic: "ctrl-d", Description: "Delete", Visible: true},
		{Mnemonic: "0", Description: "All", Visible: true},
	})

	uitest.AssertSnapshot(t, "menu", uitest.RenderSnapshot(v, 60, 3))
}

func TestTreeSnapshot(t *testing.T) {
	tree := ui.NewTree()
	ctx := context.WithValue(context.Background(), internal.KeyStyles, config.NewStyles())
	require.NoError(t, tree.Init(ctx))

	root := tview.NewTreeNode("root").SetExpanded(true)
	c1 := tview.NewTreeNode("child-1")
	c1.AddChild(tview.NewTreeNode("leaf"))
	root.AddChild(c1)
	root.AddChild(tview.NewTreeNode("child-2"))
	tree.SetRoot(root)
	tree.SetCurrentNode(root)

	uitest.AssertSnapshot(t, "tree", uitest.RenderSnapshot(tree, 40, 8))
}
