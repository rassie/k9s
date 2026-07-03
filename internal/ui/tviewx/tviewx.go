// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

// Package tviewx holds the tview extensions that k9s relied on in the (now
// retired) derailed/tview fork but that do not exist in upstream rivo/tview.
// They are reimplemented here on top of the public rivo/tview API so k9s can
// depend on upstream tview directly.
package tviewx

import "github.com/rivo/tview"

// Focusable provides a method which determines if a primitive has focus.
// Composed primitives may be focused based on the focused state of their
// contained primitives. derailed/tview exposed this via the Primitive
// interface; rivo/tview no longer does, so k9s carries it here.
type Focusable interface {
	HasFocus() bool
}

// EscapeBytes escapes the given text such that color and/or region tags are not
// recognized and substituted by tview's print functions. It is the byte-slice
// counterpart of tview.Escape (identical regex), ported from derailed/tview.
func EscapeBytes(text []byte) []byte {
	return []byte(tview.Escape(string(text)))
}
