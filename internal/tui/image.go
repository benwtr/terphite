package tui

import (
	"fmt"

	"github.com/benwtr/terphite/internal/termimg"
)

// overlayAt returns esc wrapped so it draws at an absolute 1-indexed
// (row, col) and leaves the cursor exactly where it found it.
//
// Inline images can't just be embedded in the layout string: the escape
// sequence is a single "line" as far as lipgloss and bubbletea's line
// diffing are concerned, but the terminal draws it across many rows. Laying
// out around that mismatch makes every pane below the image get pushed off
// the bottom of the screen. Instead the layout reserves a correctly-sized
// blank region, and the image is painted over it afterwards via this
// overlay, so the frame's line count stays honest.
func overlayAt(row, col int, esc string) string {
	if esc == "" {
		return ""
	}
	return fmt.Sprintf("\x1b7\x1b[%d;%dH%s\x1b8", row, col, esc)
}

// imageEscapeCache avoids rebuilding (and re-base64-encoding) an inline
// image's escape sequence on every View() call — only when the underlying
// image data actually changed (tracked by version, bumped when new image
// bytes are stored) or the display size changed do we rebuild it.
type imageEscapeCache struct {
	version int
	cols    int
	rows    int
	str     string
	built   bool
}

// escape returns the cached escape sequence for (version, cols, rows),
// rebuilding it if anything relevant changed since the last call.
func (c *imageEscapeCache) escape(proto termimg.Protocol, png []byte, version, cols, rows int) string {
	if c.built && c.version == version && c.cols == cols && c.rows == rows {
		return c.str
	}
	str := buildImageEscape(proto, png, cols, rows)
	c.version, c.cols, c.rows, c.str, c.built = version, cols, rows, str, true
	return str
}

func buildImageEscape(proto termimg.Protocol, png []byte, cols, rows int) string {
	if len(png) == 0 || cols <= 0 || rows <= 0 {
		return ""
	}
	switch proto {
	case termimg.ProtocolITerm2:
		return termimg.ITerm2Escape(png, cols, rows)
	case termimg.ProtocolKitty:
		return termimg.KittyEscape(png, cols, rows)
	default:
		return ""
	}
}
