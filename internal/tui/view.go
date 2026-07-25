package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m *Model) View() string {
	if m.quitting {
		return ""
	}
	if m.width == 0 || m.height == 0 {
		return "loading…"
	}
	if m.popup != popupNone {
		return m.viewPopup()
	}
	if m.viewMode == viewDashboard {
		return m.viewDashboardScreen()
	}
	return m.viewComposerScreen()
}

func (m *Model) viewComposerScreen() string {
	statusHeight := 3
	help := helpText(m.viewMode, m.popup)
	helpHeight := helpBoxHeight(help, m.height/2)
	bodyHeight := clampMin(m.height-statusHeight-helpHeight, 3)
	treeWidth := clampMin(m.width/4, 10)
	chartWidth := clampMin(m.width-treeWidth, 10)

	status := borderStyle.Width(clampMin(m.width-2, 1)).Height(clampMin(statusHeight-2, 1)).
		Render(renderStatus(m.timeFrom, m.autorefreshSeconds(), m.maxDataPoints, m.errMsg, m.width-4))

	tree := borderStyle.Width(clampMin(treeWidth-2, 1)).Height(clampMin(bodyHeight-2, 1)).
		Render(renderTree(m.treeRows, m.treeCursor, m.expanded, m.selectedSet(), treeWidth-4, bodyHeight-4))

	chart := borderStyle.Width(clampMin(chartWidth-2, 1)).Height(clampMin(bodyHeight-2, 1)).
		Render(renderChart(m.series, chartWidth-4, bodyHeight-4))

	body := lipgloss.JoinHorizontal(lipgloss.Top, tree, chart)

	helpBox := borderStyle.Width(clampMin(m.width-2, 1)).Height(clampMin(helpHeight-2, 1)).Render(help)

	return lipgloss.JoinVertical(lipgloss.Left, status, body, helpBox)
}

func (m *Model) viewDashboardScreen() string {
	if m.currentDashboard == nil {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, "no dashboard loaded")
	}

	statusHeight := 3
	help := helpText(viewDashboard, popupNone)
	helpHeight := helpBoxHeight(help, m.height/2)
	gridHeight := clampMin(m.height-statusHeight-helpHeight, 3)

	title := fmt.Sprintf("dashboard: %s  autorefresh: %s", m.currentDashboard.Name, onOff(m.dashboardAutorefreshOn))
	status := borderStyle.Width(clampMin(m.width-2, 1)).Height(clampMin(statusHeight-2, 1)).Render(title)

	panels := make([]dashboardPanelState, len(m.currentDashboard.Panels))
	for i, p := range m.currentDashboard.Panels {
		panels[i] = dashboardPanelState{Title: p.Title}
		if i < len(m.panelSeries) {
			panels[i].Series = m.panelSeries[i]
		}
		if i < len(m.panelErrs) {
			panels[i].Err = m.panelErrs[i]
		}
	}
	grid := renderDashboardGrid(panels, m.dashboardFocus, m.width, gridHeight)

	helpBox := borderStyle.Width(clampMin(m.width-2, 1)).Height(clampMin(helpHeight-2, 1)).Render(help)

	return lipgloss.JoinVertical(lipgloss.Left, status, grid, helpBox)
}

func (m *Model) viewPopup() string {
	var content string
	switch m.popup {
	case popupMetricsList:
		content = m.viewMetricsPopup()
	case popupPickDashboard:
		content = m.viewDashboardPicker()
	default:
		content = m.viewTextInputPopup()
	}
	box := borderStyle.Padding(1, 2).Render(content)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

func (m *Model) viewTextInputPopup() string {
	return fmt.Sprintf("%s\n\n%s\n\n[enter] submit  [esc] cancel", popupTitle(m.popup), m.input.View())
}

func popupTitle(p popupKind) string {
	switch p {
	case popupSetTimeFrom:
		return "Set relative \"from\" time (e.g. -1d12h)"
	case popupSetAutorefresh:
		return "Set autorefresh interval (seconds)"
	case popupSetMaxDataPoints:
		return "Set max datapoints (0 = unlimited)"
	case popupAddTarget:
		return "Add target"
	case popupEditTarget:
		return "Edit target"
	case popupSaveDashboardName:
		return "Save to dashboard (existing or new name)"
	}
	return ""
}

func (m *Model) viewMetricsPopup() string {
	var b strings.Builder
	b.WriteString("Selected metrics\n\n")
	if len(m.selectedMetrics) == 0 {
		b.WriteString(faintStyle.Render("(none — press ctrl+a to add a target)"))
	}
	for i, target := range m.selectedMetrics {
		style := lipgloss.NewStyle()
		if i == m.metricsCursor {
			style = style.Reverse(true)
		}
		b.WriteString(style.Render(target))
		b.WriteString("\n")
	}
	b.WriteString("\n" + helpText(viewComposer, popupMetricsList))
	return b.String()
}

func (m *Model) viewDashboardPicker() string {
	var b strings.Builder
	b.WriteString("Saved dashboards\n\n")
	if len(m.dashboardNames) == 0 {
		b.WriteString(faintStyle.Render("(none saved yet — press S in composer view to save one)"))
	}
	for i, name := range m.dashboardNames {
		style := lipgloss.NewStyle()
		if i == m.pickerCursor {
			style = style.Reverse(true)
		}
		b.WriteString(style.Render(name))
		b.WriteString("\n")
	}
	b.WriteString("\n[enter] open  [esc] cancel")
	return b.String()
}

func onOff(b bool) string {
	if b {
		return "on"
	}
	return "off"
}

// helpBoxHeight returns the total (border-inclusive) height needed to show
// help without truncation, sized to its actual content rather than a fixed
// fraction of the screen — a short guess would let help text overflow its
// box and push everything above it (including the status bar) off-screen.
func helpBoxHeight(help string, maxHeight int) int {
	needed := strings.Count(help, "\n") + 1 + 2
	if needed > maxHeight {
		needed = maxHeight
	}
	return clampMin(needed, 4)
}
