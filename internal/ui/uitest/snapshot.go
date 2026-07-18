// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

// Package uitest provides render-snapshot helpers for k9s view tests. Rendering
// a primitive onto a headless tcell SimulationScreen and capturing the cell
// grid *with styles* catches appearance regressions (colors, selection, focus,
// bold) that GetText()-based assertions miss — GetText only returns the text.
package uitest

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func colorName(c tcell.Color) string {
	if c == tcell.ColorDefault {
		return "default"
	}
	return fmt.Sprintf("#%06x", c.Hex())
}

// attrNames spells out an AttrMask so goldens read semantically ("bold+dim")
// and stay comparable across tcell versions, which raw hex would not be.
func attrNames(attr tcell.AttrMask) string {
	if attr == 0 {
		return "none"
	}
	names := []string{}
	for _, a := range []struct {
		mask tcell.AttrMask
		name string
	}{
		{tcell.AttrBold, "bold"},
		{tcell.AttrBlink, "blink"},
		{tcell.AttrReverse, "reverse"},
		{tcell.AttrUnderline, "underline"},
		{tcell.AttrDim, "dim"},
		{tcell.AttrItalic, "italic"},
		{tcell.AttrStrikeThrough, "strike"},
	} {
		if attr&a.mask != 0 {
			names = append(names, a.name)
			attr &^= a.mask
		}
	}
	if attr != 0 {
		names = append(names, fmt.Sprintf("0x%x", attr))
	}
	return strings.Join(names, "+")
}

// styleDesc is a cell's visual style as a stable, human-readable string. Keying
// on this (rather than the raw tcell values) merges visually-identical styles
// and gives a deterministic ordering — distinct tcell.Color values can share an
// RGB/Hex, which would otherwise make legend assignment flaky.
func styleDesc(c *tcell.SimCell) string {
	fg, bg, attr := c.Style.Decompose()
	return fmt.Sprintf("fg=%s bg=%s attr=%s", colorName(fg), colorName(bg), attrNames(attr))
}

// RenderSnapshot draws p onto a headless SimulationScreen of the given size and
// returns a deterministic text+style snapshot. Each rendered line is followed by
// a legend line (one char per cell) encoding that cell's style, with the most
// common style shown as '.'. A trailing block maps each char to its
// fg/bg/attributes. The result is stable across runs and diff-friendly.
func RenderSnapshot(p tview.Primitive, w, h int) string {
	s := tcell.NewSimulationScreen("UTF-8")
	if err := s.Init(); err != nil {
		panic(err)
	}
	s.SetSize(w, h)
	p.SetRect(0, 0, w, h)
	p.Draw(s)
	s.Show() // commit back buffer so GetContents (front) sees it

	cells, cw, ch := s.GetContents()

	count := map[string]int{}
	for i := range cells {
		count[styleDesc(&cells[i])]++
	}
	descs := make([]string, 0, len(count))
	for d := range count {
		descs = append(descs, d)
	}
	sort.Strings(descs)

	// Pick the most common style deterministically (tie broken by sort order).
	maxN := 0
	for _, d := range descs {
		if count[d] > maxN {
			maxN = count[d]
		}
	}
	dominant := ""
	for _, d := range descs {
		if count[d] == maxN {
			dominant = d
			break
		}
	}
	legend := map[string]byte{dominant: '.'}
	next := byte('a')
	for _, d := range descs {
		if _, ok := legend[d]; ok {
			continue
		}
		legend[d] = next
		if next == 'z' {
			next = 'A'
		} else {
			next++
		}
	}

	var out strings.Builder
	for y := range ch {
		var line, leg strings.Builder
		for x := range cw {
			c := cells[y*cw+x]
			r := ' '
			if len(c.Runes) > 0 && c.Runes[0] != 0 {
				r = c.Runes[0]
			}
			line.WriteRune(r)
			leg.WriteByte(legend[styleDesc(&c)])
		}
		out.WriteString(strings.TrimRight(line.String(), " ") + "\n")
		out.WriteString(strings.TrimRight(leg.String(), ".") + "\n")
	}

	// Include every style — the dominant one ('.') too, so a regression in the
	// most common color (e.g. the logo's foreground) shows up in the diff.
	out.WriteString("--- style legend ---\n")
	entries := make([]string, 0, len(legend))
	for d, ch := range legend {
		entries = append(entries, fmt.Sprintf("%c = %s", ch, d))
	}
	sort.Strings(entries)
	out.WriteString(strings.Join(entries, "\n"))
	out.WriteString("\n")
	return out.String()
}

// AssertSnapshot compares got against testdata/snapshots/<name>.golden (relative
// to the calling package's directory). Run the tests with K9S_UPDATE_SNAPSHOTS=1
// to (re)generate the golden files.
func AssertSnapshot(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", "snapshots", name+".golden")
	if os.Getenv("K9S_UPDATE_SNAPSHOTS") != "" {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(got), 0o600); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("missing snapshot %q; regenerate with K9S_UPDATE_SNAPSHOTS=1 (%v)", path, err)
	}
	if string(want) != got {
		t.Errorf("snapshot %q mismatch (regenerate with K9S_UPDATE_SNAPSHOTS=1):\n--- want ---\n%s--- got ---\n%s", name, want, got)
	}
}
