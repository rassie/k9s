// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/ui/uitest"
	"github.com/derailed/k9s/internal/xray"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestXrayRefreshKeepsActions verifies that a tree refresh which keeps the
// selection does not rebuild the view's actions. Every refresh rebuilds the
// tree and reselects the node by path, so the current node is a new node
// object each time; rivo/tview reports that as a selection change on the next
// draw, and rebuilding the actions reloads hotkeys and plugins from disk.
func TestXrayRefreshKeepsActions(t *testing.T) {
	// The resource metas are a process-wide registry that other tests register
	// real resources in, so this test uses a resource of its own.
	gvr := client.NewGVR("xraytest.k9s.io/v1/fakes")
	dao.MetaAccess.RegisterMeta(gvr.String(), &metav1.APIResource{
		Name:       "fakes",
		Kind:       "Fake",
		Namespaced: true,
		Verbs:      []string{"get", "list", "delete"},
	})
	s := newScenario(t)
	x, ok := NewXray(gvr).(*Xray)
	require.True(t, ok)
	require.NoError(t, x.Init(s.context()))

	// refresh mirrors the queued update in Xray.update: a freshly built tree
	// with the node at path reselected, followed by a draw.
	refresh := func(path string) {
		root := xray.NewTreeNode(gvr, "fakes")
		for _, id := range []string{"ns1/a", "ns1/b"} {
			root.Add(xray.NewTreeNode(gvr, id))
		}
		tn := makeTreeNode(root, true, false, s.app.Styles)
		var current *tview.TreeNode
		for _, c := range root.Children {
			cn := makeTreeNode(c, true, false, s.app.Styles)
			tn.AddChild(cn)
			if c.ID == path {
				current = cn
			}
		}
		require.NotNil(t, current)
		x.SetRoot(tn)
		x.SetCurrentNode(current)
		_ = uitest.RenderSnapshot(x, 80, 20)
	}

	sentinel := func() {
		x.Actions().Add(tcell.KeyF24, ui.NewKeyAction("sentinel", nil, false))
	}
	hasSentinel := func() bool {
		_, ok := x.Actions().Get(tcell.KeyF24)
		return ok
	}

	refresh("ns1/a")
	sentinel()
	refresh("ns1/a")
	assert.True(t, hasSentinel(), "a refresh keeping the selection must not rebuild the actions")

	refresh("ns1/b")
	assert.False(t, hasSentinel(), "selecting another node must rebuild the actions")
}
