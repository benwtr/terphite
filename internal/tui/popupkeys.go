package tui

import (
	"strconv"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/benwtr/terphite/internal/dashboard"
	"github.com/benwtr/terphite/internal/timerange"
)

func (m *Model) handlePopupKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.popup {
	case popupMetricsList:
		return m.handleMetricsPopupKey(msg)
	case popupPickDashboard:
		return m.handleDashboardPickerKey(msg)
	default:
		return m.handleTextInputPopupKey(msg)
	}
}

func (m *Model) handleTextInputPopupKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.popup = popupNone
		return m, nil
	case "enter":
		return m.submitTextInput()
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m *Model) submitTextInput() (tea.Model, tea.Cmd) {
	value := m.input.Value()

	switch m.popup {
	case popupSetTimeFrom:
		m.popup = popupNone
		normalized, err := timerange.Normalize(value)
		if err != nil {
			m.errMsg = err.Error()
			return m, nil
		}
		m.timeFrom = normalized
		m.fetchGen++
		return m, fetchRenderCmd(m.client, m.selectedMetrics, m.timeFrom, m.maxDataPoints, m.fetchGen)

	case popupSetAutorefresh:
		m.popup = popupNone
		n, err := strconv.Atoi(value)
		if err != nil || n <= 0 {
			n = 1
		}
		m.autorefreshInterval = n
		m.autorefreshOn = true
		return m, autorefreshTickCmd(m.autorefreshInterval)

	case popupSetMaxDataPoints:
		m.popup = popupNone
		n, err := strconv.Atoi(value)
		if err != nil || n < 0 {
			n = 0
		}
		m.maxDataPoints = n
		m.fetchGen++
		return m, fetchRenderCmd(m.client, m.selectedMetrics, m.timeFrom, m.maxDataPoints, m.fetchGen)

	case popupAddTarget:
		m.popup = popupMetricsList
		if value != "" {
			m.selectedMetrics = append(m.selectedMetrics, value)
		}
		m.fetchGen++
		return m, fetchRenderCmd(m.client, m.selectedMetrics, m.timeFrom, m.maxDataPoints, m.fetchGen)

	case popupEditTarget:
		m.popup = popupMetricsList
		if m.metricsCursor >= 0 && m.metricsCursor < len(m.selectedMetrics) {
			m.selectedMetrics[m.metricsCursor] = value
		}
		m.fetchGen++
		return m, fetchRenderCmd(m.client, m.selectedMetrics, m.timeFrom, m.maxDataPoints, m.fetchGen)

	case popupSaveDashboardName:
		m.popup = popupNone
		if value == "" {
			return m, nil
		}
		return m, saveDashboardPanelCmd(m.store, value, dashboard.Panel{
			Title:    panelTitle(m.selectedMetrics),
			Targets:  append([]string(nil), m.selectedMetrics...),
			TimeFrom: m.timeFrom,
			DrawMode: m.drawMode.String(),
		})
	}
	return m, nil
}

func (m *Model) handleMetricsPopupKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.popup = popupNone
		m.fetchGen++
		return m, fetchRenderCmd(m.client, m.selectedMetrics, m.timeFrom, m.maxDataPoints, m.fetchGen)

	case "up", "k":
		if m.metricsCursor > 0 {
			m.metricsCursor--
		}
		return m, nil
	case "down", "j":
		if m.metricsCursor < len(m.selectedMetrics)-1 {
			m.metricsCursor++
		}
		return m, nil

	case "ctrl+a":
		m.popup = popupAddTarget
		m.input = newTextInput("add target: ", "", 60)
		return m, nil

	case "ctrl+d":
		if m.metricsCursor >= 0 && m.metricsCursor < len(m.selectedMetrics) {
			m.selectedMetrics = append(m.selectedMetrics[:m.metricsCursor], m.selectedMetrics[m.metricsCursor+1:]...)
			if m.metricsCursor >= len(m.selectedMetrics) {
				m.metricsCursor = len(m.selectedMetrics) - 1
			}
		}
		m.fetchGen++
		return m, fetchRenderCmd(m.client, m.selectedMetrics, m.timeFrom, m.maxDataPoints, m.fetchGen)

	case "enter":
		if m.metricsCursor >= 0 && m.metricsCursor < len(m.selectedMetrics) {
			m.popup = popupEditTarget
			m.input = newTextInput("edit target: ", m.selectedMetrics[m.metricsCursor], 60)
		}
		return m, nil
	}
	return m, nil
}

func (m *Model) handleDashboardPickerKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.popup = popupNone
		return m, nil
	case "up", "k":
		if m.pickerCursor > 0 {
			m.pickerCursor--
		}
		return m, nil
	case "down", "j":
		if m.pickerCursor < len(m.dashboardNames)-1 {
			m.pickerCursor++
		}
		return m, nil
	case "enter":
		if m.pickerCursor >= 0 && m.pickerCursor < len(m.dashboardNames) {
			return m, loadDashboardCmd(m.store, m.dashboardNames[m.pickerCursor])
		}
		return m, nil
	}
	return m, nil
}
