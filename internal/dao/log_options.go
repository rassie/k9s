// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"fmt"
	"time"

	"github.com/derailed/k9s/internal/client"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// LogOptions represents logger options.
type LogOptions struct {
	CreateDuration   time.Duration
	Path             string
	Container        string
	DefaultContainer string
	SinceTime        string
	Lines            int64
	Buffer           int64
	LimitBytes       int64
	SinceSeconds     int64
	Head             bool
	FullLog          bool
	Previous         bool
	SingleContainer  bool
	MultiPods        bool
	ShowTimestamp    bool
	AllContainers    bool
	LogBufferSize    int
}

// Info returns the option pod and container info.
func (o *LogOptions) Info() string {
	if o.Container != "" {
		return fmt.Sprintf("%s (%s)", o.Path, o.Container)
	}

	return o.Path
}

// Clone clones options.
func (o *LogOptions) Clone() *LogOptions {
	return &LogOptions{
		Path:             o.Path,
		Container:        o.Container,
		DefaultContainer: o.DefaultContainer,
		Lines:            o.Lines,
		Buffer:           o.Buffer,
		LimitBytes:       o.LimitBytes,
		Previous:         o.Previous,
		Head:             o.Head,
		FullLog:          o.FullLog,
		SingleContainer:  o.SingleContainer,
		MultiPods:        o.MultiPods,
		ShowTimestamp:    o.ShowTimestamp,
		SinceTime:        o.SinceTime,
		SinceSeconds:     o.SinceSeconds,
		AllContainers:    o.AllContainers,
		LogBufferSize:    o.LogBufferSize,
	}
}

// HasContainer checks if a container is present.
func (o *LogOptions) HasContainer() bool {
	return o.Container != ""
}

// ToggleAllContainers toggles single or all-containers if possible.
func (o *LogOptions) ToggleAllContainers() {
	if o.SingleContainer {
		return
	}
	o.AllContainers = !o.AllContainers
	if o.AllContainers {
		o.DefaultContainer, o.Container = o.Container, ""
		return
	}

	if o.DefaultContainer != "" {
		o.Container = o.DefaultContainer
	}
}

// ToPodLogOptions returns pod log options.
func (o *LogOptions) ToPodLogOptions() *v1.PodLogOptions {
	opts := v1.PodLogOptions{
		Follow:     true,
		Timestamps: true,
		Container:  o.Container,
		Previous:   o.Previous,
		TailLines:  &o.Lines,
	}
	if o.Head {
		var maxBytes int64 = 5000
		opts.Follow = false
		opts.TailLines, opts.SinceSeconds, opts.SinceTime = nil, nil, nil
		opts.LimitBytes = &maxBytes
		return &opts
	}
	if o.FullLog {
		// Fetch the whole retained log from the container's start as a bounded,
		// non-following snapshot: no tail/since anchors, capped by LimitBytes so
		// memory can't grow without bound (the kube API offers no size preflight
		// and no server-side content filter, only this byte cap).
		opts.Follow = false
		opts.TailLines, opts.SinceSeconds, opts.SinceTime = nil, nil, nil
		if o.LimitBytes > 0 {
			opts.LimitBytes = &o.LimitBytes
		}
		return &opts
	}
	if o.SinceSeconds < 0 {
		return &opts
	}

	if o.SinceSeconds != 0 {
		opts.SinceSeconds, opts.SinceTime = &o.SinceSeconds, nil
		// The time window governs which lines to fetch; don't also cap by the
		// (small) tail count or the window is silently truncated to the last
		// Lines entries on a busy pod. Bound the fetch by the retention buffer
		// instead (nil == unbounded, e.g. full-log mode).
		opts.TailLines = o.bufferLimit()
		return &opts
	}

	if o.SinceTime == "" {
		return &opts
	}
	if t, err := time.Parse(time.RFC3339, o.SinceTime); err == nil {
		opts.SinceTime = &metav1.Time{Time: t.Add(time.Second)}
	}

	return &opts
}

// bufferLimit returns the retention buffer as a tail bound, or nil when the
// buffer is unbounded (Buffer <= 0), e.g. for full-log mode.
func (o *LogOptions) bufferLimit() *int64 {
	if o.Buffer <= 0 {
		return nil
	}
	return &o.Buffer
}

// ToLogItem add a log header to display po/co information along with the log message.
func (o *LogOptions) ToLogItem(bytes []byte) *LogItem {
	item := NewLogItem(bytes)
	if len(bytes) == 0 {
		return item
	}
	item.SingleContainer = o.SingleContainer
	if item.SingleContainer {
		item.Container = o.Container
	}
	if o.MultiPods {
		_, pod := client.Namespaced(o.Path)
		item.Pod, item.Container = pod, o.Container
	} else {
		item.Container = o.Container
	}

	return item
}

func (*LogOptions) ToErrLogItem(err error) *LogItem {
	t := time.Now().UTC().Format(time.RFC3339Nano)
	item := NewLogItem([]byte(fmt.Sprintf("%s [orange::b]%s[::-]\n", t, err)))
	item.IsError = true
	return item
}
