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
	{"[  {", "decrease time 1min, 1h"},
	{"]  }", "increase time 1min, 1h"},
	{"t", "set relative \"from\" time"},
	{"m", "metrics list popup"},
	{"i", "set autorefresh interval"},
	{"a", "autorefresh toggle"},
	{"x", "set max datapoints"},
	{"o", "open in browser"},
	{"c", "copy graphite URI to clipboard"},
	{"S", "save current graph to a dashboard"},
	{"D", "open a saved dashboard"},
	{"g", "cycle graph style (line/area/stacked)"},
	{"I", "toggle graphical mode (iterm2/kitty image)"},
	{"↑/↓", "move metrics tree cursor"},
	{"enter", "select metric / expand"},
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
	{"esc", "back to composer"},
	{"q", "quit"},
}

func helpText(mode viewMode, popup popupKind, graphMode drawMode, imageProtocol termimg.Protocol) string {
	var bindings []keyBinding
	switch {
	case popup == popupMetricsList:
		bindings = metricsPopupKeys
	case mode == viewDashboard:
		bindings = dashboardKeys
	default:
		bindings = composerKeys
	}
	lines := make([]string, 0, len(bindings))
	for _, b := range bindings {
		help := b.Help
		switch b.Key {
		case "g":
			help = fmt.Sprintf("%s [%s]", help, graphMode)
		case "I":
			help = fmt.Sprintf("%s [%s]", help, imageProtocol)
		}
		lines = append(lines, fmt.Sprintf("%-8s %s", b.Key, help))
	}
	return strings.Join(lines, "\n")
}
