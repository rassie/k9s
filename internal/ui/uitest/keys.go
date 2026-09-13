// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package uitest

import (
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
)

// ScanKey feeds seq through tcell's terminal input parser, the way bytes read
// from a terminal are decoded, and returns the key event it produces. A lone
// ESC is only reported once tcell's escape timeout has passed.
func ScanKey(t testing.TB, seq string) *tcell.EventKey {
	t.Helper()
	evts := make(chan tcell.Event, 4)
	tcell.NewInputProcessor(evts).ScanUTF8([]byte(seq))
	select {
	case evt := <-evts:
		k, ok := evt.(*tcell.EventKey)
		if !ok {
			t.Fatalf("%q produced %T, not a key event", seq, evt)
		}
		return k
	case <-time.After(time.Second):
		t.Fatalf("no key event for %q", seq)
		return nil
	}
}
