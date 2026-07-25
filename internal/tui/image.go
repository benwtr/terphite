package tui

import "github.com/benwtr/terphite/internal/termimg"

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
