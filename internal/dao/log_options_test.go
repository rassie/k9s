// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao_test

import (
	"testing"

	"github.com/derailed/k9s/internal/dao"
	"github.com/stretchr/testify/assert"
)

func TestLogOptionsToPodLogOptionsTailLines(t *testing.T) {
	lines, buffer := int64(100), int64(5000)
	uu := map[string]struct {
		opts      dao.LogOptions
		tailLines *int64
		sinceSecs *int64
	}{
		"tail": {
			opts:      dao.LogOptions{Lines: lines, Buffer: buffer},
			tailLines: &lines,
		},
		"since-window-capped-by-buffer": {
			opts:      dao.LogOptions{Lines: lines, Buffer: buffer, SinceSeconds: 300},
			tailLines: &buffer,
			sinceSecs: func() *int64 { s := int64(300); return &s }(),
		},
		"since-window-unbounded-buffer": {
			opts:      dao.LogOptions{Lines: lines, SinceSeconds: 300},
			tailLines: nil,
			sinceSecs: func() *int64 { s := int64(300); return &s }(),
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			plo := u.opts.ToPodLogOptions()
			assert.Equal(t, u.tailLines, plo.TailLines)
			assert.Equal(t, u.sinceSecs, plo.SinceSeconds)
		})
	}
}

func TestLogOptionsToggleAllContainers(t *testing.T) {
	uu := map[string]struct {
		opts dao.LogOptions
		co   string
		want bool
	}{
		"empty": {
			opts: dao.LogOptions{},
			want: true,
		},
		"container": {
			opts: dao.LogOptions{Container: "blee"},
			want: true,
		},
		"default-container": {
			opts: dao.LogOptions{AllContainers: true},
			co:   "blee",
		},
		"single-container": {
			opts: dao.LogOptions{Container: "blee", SingleContainer: true},
			co:   "blee",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			u.opts.DefaultContainer = "blee"
			u.opts.ToggleAllContainers()
			assert.Equal(t, u.want, u.opts.AllContainers)
			assert.Equal(t, u.co, u.opts.Container)
		})
	}
}
