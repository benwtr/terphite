package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/benwtr/terphite/internal/timerange"
)

func (m *Model) handleComposerKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		m.quitting = true
		return m, tea.Quit

	case "up", "k":
		if m.treeCursor > 0 {
			m.treeCursor--
		}
		return m, nil
	case "down", "j":
		if m.treeCursor < len(m.treeRows)-1 {
			m.treeCursor++
		}
		return m, nil

	case "enter", " ":
		return m.activateTreeCursor()

	case "[":
		return m.adjustTime(-timerange.UnitMinute, timerange.UnitMinute, timerange.Default)
	case "]":
		return m.adjustTime(timerange.UnitMinute, 0, "")
	case "{":
		return m.adjustTime(-timerange.UnitHour, timerange.UnitHour, "-1h")
	case "}":
		return m.adjustTime(timerange.UnitHour, 0, "")

	case "t":
		m.popup = popupSetTimeFrom
		m.input = newTextInput("from: ", m.timeFrom, 40)
		return m, nil

	case "m":
		m.popup = popupMetricsList
		m.metricsCursor = 0
		return m, nil

	case "i":
		m.popup = popupSetAutorefresh
		m.input = newTextInput("autorefresh seconds: ", fmt.Sprintf("%d", m.autorefreshInterval), 10)
		return m, nil

	case "a":
		m.autorefreshOn = !m.autorefreshOn
		if m.autorefreshOn {
			return m, autorefreshTickCmd(m.autorefreshInterval)
		}
		return m, nil

	case "x":
		m.popup = popupSetMaxDataPoints
		m.input = newTextInput("max datapoints (0=unlimited): ", fmt.Sprintf("%d", m.maxDataPoints), 10)
		return m, nil

	case "o":
		_ = openInBrowser(m.client.RenderURL(m.renderQuery()))
		return m, nil

	case "c":
		_ = copyToClipboard(m.client.RenderURL(m.renderQuery()))
		return m, nil

	case "S":
		m.popup = popupSaveDashboardName
		m.input = newTextInput("save to dashboard: ", "", 40)
		return m, nil

	case "D":
		m.popup = popupPickDashboard
		m.pickerCursor = 0
		return m, fetchDashboardListCmd(m.store)
	}
	return m, nil
}

// adjustTime shifts the current time_from by delta seconds. When decreasing
// (delta < 0) and the result would fall below floor, it resets to
// resetValue instead — mirroring the original's per-key minimum ("-1min"
// for the minute keys, "-1h" for the hour keys).
func (m *Model) adjustTime(delta, floor int, resetValue string) (tea.Model, tea.Cmd) {
	cur, err := timerange.ParseSeconds(m.timeFrom)
	if err != nil {
		cur = 0
	}
	next := cur + delta
	if delta < 0 && next < floor {
		m.timeFrom = resetValue
	} else {
		m.timeFrom = timerange.Format(next)
	}
	m.fetchGen++
	return m, fetchRenderCmd(m.client, m.selectedMetrics, m.timeFrom, m.maxDataPoints, m.fetchGen)
}

func (m *Model) activateTreeCursor() (tea.Model, tea.Cmd) {
	if m.treeCursor < 0 || m.treeCursor >= len(m.treeRows) {
		return m, nil
	}
	node := m.treeRows[m.treeCursor].node

	var cmd tea.Cmd
	if node.Leaf {
		m.toggleMetric(node.Path)
		m.fetchGen++
		cmd = fetchRenderCmd(m.client, m.selectedMetrics, m.timeFrom, m.maxDataPoints, m.fetchGen)
	}
	if len(node.Children) > 0 {
		m.expanded[node.Path] = !m.expanded[node.Path]
		m.refreshTreeRows()
	}
	return m, cmd
}

func (m *Model) toggleMetric(path string) {
	for i, p := range m.selectedMetrics {
		if p == path {
			m.selectedMetrics = append(m.selectedMetrics[:i], m.selectedMetrics[i+1:]...)
			return
		}
	}
	m.selectedMetrics = append(m.selectedMetrics, path)
}

func (m *Model) refreshTreeRows() {
	m.treeRows = flattenTree(m.tree, m.expanded)
	if m.treeCursor >= len(m.treeRows) {
		m.treeCursor = len(m.treeRows) - 1
	}
	if m.treeCursor < 0 {
		m.treeCursor = 0
	}
}

func panelTitle(targets []string) string {
	if len(targets) == 0 {
		return "(no targets)"
	}
	if len(targets) == 1 {
		return targets[0]
	}
	return fmt.Sprintf("%s +%d more", targets[0], len(targets)-1)
}
