package tui

import (
	"fmt"
	"strings"

	"github.com/benwtr/terphite/internal/termimg"
)

type keyBinding struct {
	Key  string
	Help string
}

var composerKeys = []keyBinding{
	{"↑/↓", "move metrics tree cursor"},
	{"enter", "select metric / expand"},
	{"[  {", "decrease time 1min, 1h"},
	{"]  }", "increase time 1min, 1h"},
	{"t", "set relative \"from\" time"},
	{"m", "metrics list popup"},
	{"i", "set autorefresh interval"},
	{"a", "autorefresh toggle"},
	{"x", "set max datapoints"},
	{"g", "cycle graph style"},
	{"I", "toggle graphical mode"},
	{"o", "open in browser"},
	{"c", "copy graphite URI to clipboard"},
	{"S", "save graph to a dashboard"},
	{"D", "open a saved dashboard"},
	{"l", "redraw screen"},
	{"?", "collapse/expand this help"},
	{"q", "quit"},
}

var metricsPopupKeys = []keyBinding{
	{"ctrl+a", "append new target"},
	{"ctrl+d", "delete selected"},
	{"enter", "edit selected"},
	{"esc", "close"},
}

var dashboardKeys = []keyBinding{
	{"←/→/↑/↓", "move between panels"},
	{"enter", "edit panel in composer"},
	{"ctrl+d", "remove panel"},
	{"a", "toggle autorefresh"},
	{"?", "collapse/expand this help"},
	{"esc", "back to composer"},
	{"q", "quit"},
}

func bindingsFor(mode viewMode, popup popupKind) []keyBinding {
	switch {
	case popup == popupMetricsList:
		return metricsPopupKeys
	case mode == viewDashboard:
		return dashboardKeys
	default:
		return composerKeys
	}
}

// annotate appends the current value to the bindings that toggle between
// states, so the help doubles as a status readout.
func annotate(b keyBinding, graphMode drawMode, imageProtocol termimg.Protocol) string {
	switch b.Key {
	case "g":
		return fmt.Sprintf("%s [%s]", b.Help, graphMode)
	case "I":
		return fmt.Sprintf("%s [%s]", b.Help, imageProtocol)
	}
	return b.Help
}

// helpText lays the key bindings out in as many columns as fit, keeping
// the help bar short instead of spending a third of the screen on one
// binding per line. When collapsed it shrinks to a single hint line.
//
// The result is guaranteed to be at most maxRows lines and at most width
// columns wide: callers size the help box from what this returns, and
// lipgloss's Height() only pads, so anything larger than the declared box
// would silently overflow and push the rest of the frame off-screen.
func helpText(mode viewMode, popup popupKind, graphMode drawMode, imageProtocol termimg.Protocol, width, maxRows int, collapsed bool) string {
	if collapsed {
		return fmt.Sprintf("? help   g %s   I %s   q quit", graphMode, imageProtocol)
	}

	bindings := bindingsFor(mode, popup)
	n := len(bindings)
	if n == 0 {
		return ""
	}
	maxRows = clampMin(maxRows, 1)
	if width <= 0 {
		width = 1 << 20 // unconstrained (popups size to their content)
	}

	keyW, helpW := 0, 0
	for _, b := range bindings {
		if r := len([]rune(b.Key)); r > keyW {
			keyW = r
		}
		if r := len([]rune(annotate(b, graphMode, imageProtocol))); r > helpW {
			helpW = r
		}
	}

	const gutter = 3
	// Enough columns to fit the row budget, widened further if there's room.
	cols := (n + maxRows - 1) / maxRows
	if byWidth := width / (keyW + 1 + helpW + gutter); byWidth > cols {
		cols = byWidth
	}
	cols = clampMin(cols, 1)
	if cols > n {
		cols = n
	}
	rows := (n + cols - 1) / cols

	// Share the width across the columns, truncating labels if the row
	// budget forced more columns than would naturally fit.
	cellW := clampMin(width/cols-gutter, keyW+2)
	labelW := clampMin(cellW-keyW-1, 1)

	cells := make([]string, n)
	for i, b := range bindings {
		label := []rune(annotate(b, graphMode, imageProtocol))
		if len(label) > labelW {
			label = label[:labelW]
		}
		cells[i] = fmt.Sprintf("%-*s %-*s", keyW, b.Key, labelW, string(label))
	}

	// Column-major fill so each column reads top-to-bottom.
	lines := make([]string, rows)
	for r := 0; r < rows; r++ {
		var b strings.Builder
		for c := 0; c < cols; c++ {
			i := c*rows + r
			if i >= n {
				break
			}
			if c > 0 {
				b.WriteString(strings.Repeat(" ", gutter))
			}
			b.WriteString(cells[i])
		}
		line := []rune(strings.TrimRight(b.String(), " "))
		if len(line) > width {
			line = line[:width]
		}
		lines[r] = string(line)
	}
	return strings.Join(lines, "\n")
}
