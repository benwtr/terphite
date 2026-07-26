package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/benwtr/terphite/internal/termimg"
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
	help := helpText(m.viewMode, m.popup, m.drawMode, m.imageProtocol, m.width-4, m.helpBudgetRows(), m.helpCollapsed)
	helpHeight := helpBoxHeight(help)
	bodyHeight := clampMin(m.height-statusHeight-helpHeight, 3)
	treeWidth := clampMin(m.width/4, 10)
	chartWidth := clampMin(m.width-treeWidth, 10)

	status := borderStyle.Width(clampMin(m.width-2, 1)).Height(clampMin(statusHeight-2, 1)).
		Render(renderStatus(m.timeFrom, m.autorefreshSeconds(), m.maxDataPoints, m.errMsg, m.width-4))

	tree := borderStyle.Width(clampMin(treeWidth-2, 1)).Height(clampMin(bodyHeight-2, 1)).
		Render(renderTree(m.treeRows, m.treeCursor, m.expanded, m.selectedSet(), treeWidth-4, bodyHeight-4))

	chart := m.renderChartPane(chartWidth, bodyHeight)

	body := lipgloss.JoinHorizontal(lipgloss.Top, tree, chart)

	helpBox := borderStyle.Width(clampMin(m.width-2, 1)).Height(clampMin(helpHeight-2, 1)).Render(help)

	frame := lipgloss.JoinVertical(lipgloss.Left, status, body, helpBox)

	// The chart pane's inner area starts one row below the status box and
	// one column inside the chart pane's left border (both 1-indexed).
	return frame + m.composerImageOverlay(statusHeight+2, treeWidth+2, chartWidth, bodyHeight)
}

// renderChartPane renders the composer's chart pane at the given outer
// width x height (border included). In graphical mode it renders an
// empty box of the correct size; the image itself is painted over that
// region by composerImageOverlay (see overlayAt for why).
func (m *Model) renderChartPane(width, height int) string {
	box := borderStyle.Width(clampMin(width-2, 1)).Height(clampMin(height-2, 1))
	if m.imageActive() {
		return box.Render("")
	}
	return box.Render(renderChart(m.series, m.drawMode, width-4, height-4))
}

// imageActive reports whether the composer should show a rendered image
// rather than the ASCII chart.
func (m *Model) imageActive() bool {
	return m.imageProtocol != termimg.ProtocolNone && len(m.imageBytes) > 0
}

func (m *Model) composerImageOverlay(row, col, chartWidth, bodyHeight int) string {
	if !m.imageActive() {
		return ""
	}
	cols := clampMin(chartWidth-2, 1)
	rows := clampMin(bodyHeight-2, 1)
	return overlayAt(row, col, m.imageCache.escape(m.imageProtocol, m.imageBytes, m.imageVersion, cols, rows))
}

func (m *Model) viewDashboardScreen() string {
	if m.currentDashboard == nil {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, "no dashboard loaded")
	}

	statusHeight := 3
	help := helpText(viewDashboard, popupNone, drawLine, m.imageProtocol, m.width-4, m.helpBudgetRows(), m.helpCollapsed)
	helpHeight := helpBoxHeight(help)
	gridHeight := clampMin(m.height-statusHeight-helpHeight, 3)

	title := fmt.Sprintf("dashboard: %s  autorefresh: %s", m.currentDashboard.Name, onOff(m.dashboardAutorefreshOn))
	status := borderStyle.Width(clampMin(m.width-2, 1)).Height(clampMin(statusHeight-2, 1)).Render(title)

	panels := make([]dashboardPanelState, len(m.currentDashboard.Panels))
	for i, p := range m.currentDashboard.Panels {
		panels[i] = dashboardPanelState{
			Title:      p.Title,
			Mode:       parseDrawMode(p.DrawMode),
			ImageProto: m.imageProtocol,
		}
		if i < len(m.panelSeries) {
			panels[i].Series = m.panelSeries[i]
		}
		if i < len(m.panelErrs) {
			panels[i].Err = m.panelErrs[i]
		}
		if i < len(m.panelImages) {
			panels[i].ImageBytes = m.panelImages[i]
		}
		if i < len(m.panelImageErrs) && m.panelImageErrs[i] != nil {
			panels[i].Err = m.panelImageErrs[i]
		}
		if i < len(m.panelImageVersions) {
			panels[i].ImageVersion = m.panelImageVersions[i]
		}
		if i < len(m.panelImageCaches) {
			panels[i].ImageCache = &m.panelImageCaches[i]
		}
	}
	// The grid starts directly below the status box; renderDashboardGrid
	// returns the image overlays separately so the frame's line count stays
	// honest (see overlayAt).
	grid, overlays := renderDashboardGrid(panels, m.dashboardFocus, m.width, gridHeight, statusHeight)

	helpBox := borderStyle.Width(clampMin(m.width-2, 1)).Height(clampMin(helpHeight-2, 1)).Render(help)

	return lipgloss.JoinVertical(lipgloss.Left, status, grid, helpBox) + overlays
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
	// The popup's help is always expanded — it's only four bindings, and
	// the popup is sized to its content rather than the screen.
	b.WriteString("\n" + helpText(viewComposer, popupMetricsList, m.drawMode, m.imageProtocol, 0, len(metricsPopupKeys), false))
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

// helpBudgetRows is how many rows of help the layout is willing to spend,
// leaving the rest of the screen for the chart. helpText fits itself into
// this by adding columns.
func (m *Model) helpBudgetRows() int {
	return clampMin(m.height/4, 1)
}

// helpBoxHeight returns the border-inclusive height of the help box, sized
// to the text helpText actually produced. It must not be a guess: lipgloss
// pads to a declared height but never truncates, so a box declared shorter
// than its content silently overflows and pushes the frame off-screen.
func helpBoxHeight(help string) int {
	return strings.Count(help, "\n") + 1 + 2
}
