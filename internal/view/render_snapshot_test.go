// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view_test

import (
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/ui/uitest"
	"github.com/derailed/k9s/internal/view"
	"github.com/stretchr/testify/require"
)

// Render-snapshot tests capture the styled cell grid (colors, attributes) that
// GetText()-based tests cannot see. Regenerate with:
// K9S_UPDATE_SNAPSHOTS=1 go test ./internal/view/...

func TestLogIndicatorSnapshot(t *testing.T) {
	li := view.NewLogIndicator(config.NewConfig(nil), config.NewStyles(), true)
	uitest.AssertSnapshot(t, "log_indicator", uitest.RenderSnapshot(li, 110, 1))
}

// TestLogAnsiSnapshot drives ANSI-colored log lines through k9s' own log
// pipeline (LogItem -> Flush -> the log view's ANSI writer) and snapshots the
// resulting cells, so the escape handling and color mapping k9s wires up are
// covered end to end.
func TestLogAnsiSnapshot(t *testing.T) {
	opts := dao.LogOptions{
		Path:      "fred/p1",
		Container: "blee",
	}
	v := view.NewLog(client.PodGVR, &opts)
	require.NoError(t, v.Init(makeContext(t)))

	// Real pod log lines carry an RFC3339 timestamp first (Timestamps: true);
	// rendering with showTime=false strips that leading token.
	items := dao.NewLogItems()
	items.Add(
		dao.NewLogItemFromString("2018-12-14T10:36:43.326972-07:00 \033[0;31mred error\033[0m plain\n"),
		dao.NewLogItemFromString("2018-12-14T10:36:44.326972-07:00 \033[1;32mbold green\033[0m tail\n"),
	)
	ll := make([][]byte, items.Len())
	items.Lines(0, false, ll)
	v.Flush(ll)

	uitest.AssertSnapshot(t, "log_ansi", uitest.RenderSnapshot(v.Logs(), 46, 5))
}
