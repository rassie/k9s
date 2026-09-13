// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view_test

import (
	"testing"

	"github.com/derailed/k9s/internal/config/mock"
	"github.com/derailed/k9s/internal/view"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppNew(t *testing.T) {
	a := view.NewApp(mock.NewMockConfig(t))
	_ = a.Init("blee", 10)

	assert.Equal(t, 14, a.GetActions().Len())
}

func TestAppSuggestCommandWithoutConnection(t *testing.T) {
	a := view.NewApp(mock.NewMockConfig(t))
	require.NoError(t, a.Init("blee", 10))

	assert.NotPanics(t, func() {
		for _, r := range "po" {
			a.CmdBuff().Add(r)
		}
	})
	assert.Empty(t, a.CmdBuff().Suggestions())
}
