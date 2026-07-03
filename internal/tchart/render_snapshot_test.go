// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package tchart_test

import (
	"testing"

	"github.com/derailed/k9s/internal/tchart"
	"github.com/derailed/k9s/internal/ui/uitest"
)

// Render-snapshot tests for the pulse-view widgets, which draw directly onto
// the screen with their own Draw implementations. The SparkLine is deliberately
// not covered: its x-axis prints wall-clock labels (time.Now), which makes its
// rendering nondeterministic. Regenerate with:
// K9S_UPDATE_SNAPSHOTS=1 go test ./internal/tchart/...

func TestGaugeSnapshot(t *testing.T) {
	g := tchart.NewGauge("pods")
	g.Add(25, 3)

	uitest.AssertSnapshot(t, "gauge", uitest.RenderSnapshot(g, 40, 8))
}
