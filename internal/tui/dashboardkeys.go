package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) handleDashboardKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.currentDashboard == nil {
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "esc":
			m.viewMode = viewComposer
		}
		return m, nil
	}

	n := len(m.currentDashboard.Panels)
	cols := gridColumns(n)

	switch msg.String() {
	case "q", "ctrl+c":
		m.quitting = true
		return m, tea.Quit

	case "esc":
		m.viewMode = viewComposer
		return m, nil

	case "left", "h":
		if n > 0 && m.dashboardFocus%cols > 0 {
			m.dashboardFocus--
		}
		return m, nil

	case "right", "l":
		if n > 0 && m.dashboardFocus%cols < cols-1 && m.dashboardFocus+1 < n {
			m.dashboardFocus++
		}
		return m, nil

	case "up", "k":
		if m.dashboardFocus-cols >= 0 {
			m.dashboardFocus -= cols
		}
		return m, nil

	case "down", "j":
		if m.dashboardFocus+cols < n {
			m.dashboardFocus += cols
		}
		return m, nil

	case "enter":
		if n == 0 {
			return m, nil
		}
		p := m.currentDashboard.Panels[m.dashboardFocus]
		m.selectedMetrics = append([]string(nil), p.Targets...)
		m.timeFrom = p.TimeFrom
		m.drawMode = parseDrawMode(p.DrawMode)
		m.viewMode = viewComposer
		return m, m.refreshCmd()

	case "ctrl+d":
		if n == 0 {
			return m, nil
		}
		d := m.currentDashboard
		d.Panels = append(d.Panels[:m.dashboardFocus], d.Panels[m.dashboardFocus+1:]...)
		if m.dashboardFocus >= len(d.Panels) {
			m.dashboardFocus = len(d.Panels) - 1
		}
		if m.dashboardFocus < 0 {
			m.dashboardFocus = 0
		}
		return m, saveDashboardCmd(m.store, d)

	case "a":
		m.dashboardAutorefreshOn = !m.dashboardAutorefreshOn
		if m.dashboardAutorefreshOn {
			return m, dashboardAutorefreshTickCmd()
		}
		return m, nil
	}
	return m, nil
}
