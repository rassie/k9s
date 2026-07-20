// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package tviewx

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// ANSIWriter translates ANSI SGR escape sequences in the written text into
// tview style tags, replacing tview.ANSIWriter which mistranslates attribute
// resets for tview >= 0.34:
//
//  1. tview's tag parser applies flag letters additively (lowercase turns an
//     attribute on, uppercase turns it off; a missing letter means "keep").
//     tview.ANSIWriter only ever emits the currently-on set, so an attribute
//     switched off mid-stream (SGR 22/24/...) is never switched off in the
//     rendered text.
//  2. Underline carries extra state beyond the attribute bit: tcell renders
//     from Style's UnderlineStyle, which only an explicit 'U' flag clears.
//     The '-' flag reset emitted by tview.ANSIWriter for SGR 0 clears just the
//     bit, so a single ESC[4m in a log stream leaves everything after it
//     underlined for the lifetime of the view.
//
// This writer tracks the SGR state itself and emits an explicit off-letter for
// every attribute that changes state, which is correct under both quirks.
func ANSIWriter(w io.Writer) io.Writer {
	return &ansiWriter{w: w}
}

// sgrAttrs is a bitmask of the text attributes expressible as tview flags.
type sgrAttrs uint8

const (
	sgrBold sgrAttrs = 1 << iota
	sgrDim
	sgrItalic
	sgrUnderline
	sgrBlink
	sgrReverse
	sgrStrike
)

// flagLetters maps each attribute bit to its lowercase tview flag letter.
var flagLetters = []struct {
	bit    sgrAttrs
	letter byte
}{
	{sgrBold, 'b'},
	{sgrDim, 'd'},
	{sgrItalic, 'i'},
	{sgrUnderline, 'u'},
	{sgrBlink, 'l'},
	{sgrReverse, 'r'},
	{sgrStrike, 's'},
}

const (
	ansiText = iota
	ansiEscape
	ansiCSI
	ansiOSC
)

type ansiWriter struct {
	w     io.Writer
	state int
	param bytes.Buffer // CSI parameter bytes

	// current SGR state; fg/bg hold tview color words ("" = terminal default)
	fg, bg string
	attrs  sgrAttrs
}

func (a *ansiWriter) Write(p []byte) (int, error) {
	var out bytes.Buffer
	out.Grow(len(p))

	for _, r := range string(p) {
		switch a.state {
		case ansiEscape:
			switch r {
			case '[':
				a.param.Reset()
				a.state = ansiCSI
			case 'c': // RIS: full reset.
				a.writeTag(&out, "", "", 0)
				a.state = ansiText
			case 'P', ']', 'X', '^', '_': // DCS/OSC/SOS/PM/APC: swallow.
				a.state = ansiOSC
			default:
				a.state = ansiText
			}
		case ansiCSI:
			switch {
			case r >= 0x30 && r <= 0x3f: // Parameter bytes.
				a.param.WriteRune(r)
			case r >= 0x20 && r <= 0x2f: // Intermediate bytes: ignore.
			case r >= 0x40 && r <= 0x7e: // Final byte.
				switch r {
				case 'm':
					a.sgr(&out, a.param.String())
				case 'E': // CNL: next line.
					count, _ := strconv.Atoi(a.param.String())
					if count == 0 {
						count = 1
					}
					out.WriteString(strings.Repeat("\n", count))
				}
				a.state = ansiText
			default: // Malformed: abort sequence.
				a.state = ansiText
			}
		case ansiOSC:
			// Terminated by BEL or by ESC (usually the ESC of an ST "ESC \").
			switch r {
			case '\a':
				a.state = ansiText
			case 27:
				a.state = ansiEscape
			}
		default: // ansiText
			if r == 27 {
				a.state = ansiEscape
			} else {
				out.WriteRune(r)
			}
		}
	}

	if _, err := a.w.Write(out.Bytes()); err != nil {
		return 0, err
	}
	return len(p), nil
}

// basicColors are the tview words for the 16 base ANSI colors, matching the
// table used by tview.ANSIWriter.
var basicColors = []string{
	"black", "maroon", "green", "olive", "navy", "purple", "teal", "silver",
	"gray", "red", "lime", "yellow", "blue", "fuchsia", "aqua", "white",
}

// eightBit resolves an 8-bit palette index to a tview color word.
func eightBit(n int) string {
	switch {
	case n < 0 || n > 255:
		return ""
	case n <= 15:
		return basicColors[n]
	case n <= 231: // 6x6x6 cube.
		r, g, b := (n-16)/36, ((n-16)/6)%6, (n-16)%6
		return fmt.Sprintf("#%02x%02x%02x", 255*r/5, 255*g/5, 255*b/5)
	default: // Greyscale ramp.
		grey := 255 * (n - 232) / 23
		return fmt.Sprintf("#%02x%02x%02x", grey, grey, grey)
	}
}

// sgr applies one SGR sequence to the tracked state and emits a style tag for
// whatever changed.
func (a *ansiWriter) sgr(out *bytes.Buffer, params string) {
	fg, bg, attrs := a.fg, a.bg, a.attrs

	fields := strings.Split(params, ";")
	if params == "" {
		fields = []string{"0"}
	}
	for i := 0; i < len(fields); i++ {
		// Support colon sub-parameters (e.g. "4:3", "38:5:196").
		sub := strings.Split(fields[i], ":")
		n, err := strconv.Atoi(sub[0])
		if err != nil {
			continue
		}
		switch n {
		case 0:
			fg, bg, attrs = "", "", 0
		case 1:
			attrs |= sgrBold
		case 2:
			attrs |= sgrDim
		case 3:
			attrs |= sgrItalic
		case 4:
			if len(sub) > 1 && sub[1] == "0" { // "4:0" = underline off.
				attrs &^= sgrUnderline
			} else {
				attrs |= sgrUnderline
			}
		case 5:
			attrs |= sgrBlink
		case 7:
			attrs |= sgrReverse
		case 9:
			attrs |= sgrStrike
		case 22:
			attrs &^= sgrBold | sgrDim
		case 23:
			attrs &^= sgrItalic
		case 24:
			attrs &^= sgrUnderline
		case 25:
			attrs &^= sgrBlink
		case 27:
			attrs &^= sgrReverse
		case 29:
			attrs &^= sgrStrike
		case 30, 31, 32, 33, 34, 35, 36, 37:
			fg = basicColors[n-30]
		case 39:
			fg = ""
		case 40, 41, 42, 43, 44, 45, 46, 47:
			bg = basicColors[n-40]
		case 49:
			bg = ""
		case 90, 91, 92, 93, 94, 95, 96, 97:
			fg = basicColors[n-82]
		case 100, 101, 102, 103, 104, 105, 106, 107:
			bg = basicColors[n-92]
		case 38, 48:
			var color string
			// Colon form carries the sub-parameters inline; semicolon form
			// consumes the following fields.
			args := sub[1:]
			if len(args) == 0 {
				args = fields[i+1:]
			}
			switch {
			case len(args) >= 2 && args[0] == "5":
				if p, err := strconv.Atoi(args[1]); err == nil {
					color = eightBit(p)
				}
				if len(sub) == 1 {
					i += 2
				}
			case len(args) >= 4 && args[0] == "2":
				rgb := args[1:4]
				// The ITU colon form has a (blank) colorspace slot: 38:2::R:G:B.
				if len(sub) > 1 && len(args) >= 5 && args[1] == "" {
					rgb = args[2:5]
				}
				r, e1 := strconv.Atoi(rgb[0])
				g, e2 := strconv.Atoi(rgb[1])
				b, e3 := strconv.Atoi(rgb[2])
				if e1 == nil && e2 == nil && e3 == nil {
					color = fmt.Sprintf("#%02x%02x%02x", r, g, b)
				}
				if len(sub) == 1 {
					i += 4
				}
			default:
				// Unknown extended-color form: bail out of this sequence to
				// avoid misreading its arguments as SGR codes.
				a.writeTag(out, fg, bg, attrs)
				return
			}
			if color != "" {
				if n == 38 {
					fg = color
				} else {
					bg = color
				}
			}
		}
	}

	a.writeTag(out, fg, bg, attrs)
}

// writeTag emits a style tag transitioning the parser from the current state
// to the given one, with explicit uppercase letters for attributes turning
// off, and updates the tracked state. No-op if nothing changed.
func (a *ansiWriter) writeTag(out *bytes.Buffer, fg, bg string, attrs sgrAttrs) {
	var fgField, bgField, flags string
	if fg != a.fg {
		if fgField = fg; fg == "" {
			fgField = "-"
		}
	}
	if bg != a.bg {
		if bgField = bg; bg == "" {
			bgField = "-"
		}
	}
	for _, fl := range flagLetters {
		switch on, was := attrs&fl.bit != 0, a.attrs&fl.bit != 0; {
		case on && !was:
			flags += string(fl.letter)
		case !on && was:
			flags += string(fl.letter - ('a' - 'A'))
		}
	}
	a.fg, a.bg, a.attrs = fg, bg, attrs

	switch {
	case flags != "":
		fmt.Fprintf(out, "[%s:%s:%s]", fgField, bgField, flags)
	case bgField != "":
		fmt.Fprintf(out, "[%s:%s]", fgField, bgField)
	case fgField != "":
		fmt.Fprintf(out, "[%s]", fgField)
	}
}
