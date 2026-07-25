// Package tui implements terphite's terminal UI: a composer view for
// building a single graph, and a dashboard view for viewing several saved
// graphs at once as a grid.
package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/benwtr/terphite/internal/dashboard"
	"github.com/benwtr/terphite/internal/graphite"
	"github.com/benwtr/terphite/internal/timerange"
)

type viewMode int

const (
	viewComposer viewMode = iota
	viewDashboard
)

type popupKind int

const (
	popupNone popupKind = iota
	popupSetTimeFrom
	popupSetAutorefresh
	popupSetMaxDataPoints
	popupMetricsList
	popupEditTarget
	popupAddTarget
	popupSaveDashboardName
	popupPickDashboard
)

const (
	defaultAutorefreshInterval = 10
	defaultMaxDataPoints       = 300
)

// Config configures a new Model.
type Config struct {
	GraphiteURI  string
	Username     string
	Password     string
	DashboardDir string // empty uses dashboard.DefaultDir()
}

// Model is terphite's root bubbletea model.
type Model struct {
	client *graphite.Client
	store  *dashboard.Store

	width, height int
	quitting      bool

	viewMode viewMode
	popup    popupKind
	input    textinput.Model
	errMsg   string

	tree            *graphite.MetricNode
	treeRows        []treeRow
	treeCursor      int
	expanded        map[string]bool
	selectedMetrics []string
	timeFrom        string

	autorefreshOn       bool
	autorefreshInterval int
	maxDataPoints       int

	series   []graphite.Series
	fetchGen int

	metricsCursor int

	dashboardNames []string
	pickerCursor   int

	currentDashboard       *dashboard.Dashboard
	panelSeries            [][]graphite.Series
	panelErrs              []error
	dashboardFocus         int
	dashboardAutorefreshOn bool
	dashboardFetchGen      int
}

// New builds a Model ready to run.
func New(cfg Config) (*Model, error) {
	client, err := graphite.NewClient(cfg.GraphiteURI)
	if err != nil {
		return nil, err
	}
	if cfg.Username != "" || cfg.Password != "" {
		client.SetAuth(cfg.Username, cfg.Password)
	}

	dir := cfg.DashboardDir
	if dir == "" {
		dir, err = dashboard.DefaultDir()
		if err != nil {
			return nil, err
		}
	}

	return &Model{
		client:              client,
		store:               dashboard.NewStore(dir),
		viewMode:            viewComposer,
		timeFrom:            timerange.Default,
		autorefreshInterval: defaultAutorefreshInterval,
		maxDataPoints:       defaultMaxDataPoints,
		expanded:            make(map[string]bool),
	}, nil
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		fetchMetricsCmd(m.client),
		fetchRenderCmd(m.client, m.selectedMetrics, m.timeFrom, m.maxDataPoints, m.fetchGen),
	)
}

func (m *Model) autorefreshSeconds() int {
	if m.autorefreshOn {
		return m.autorefreshInterval
	}
	return 0
}

func (m *Model) renderQuery() graphite.RenderQuery {
	return graphite.RenderQuery{
		Targets:       m.selectedMetrics,
		From:          m.timeFrom,
		MaxDataPoints: m.maxDataPoints,
	}
}

func (m *Model) selectedSet() map[string]bool {
	set := make(map[string]bool, len(m.selectedMetrics))
	for _, p := range m.selectedMetrics {
		set[p] = true
	}
	return set
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case metricsLoadedMsg:
		m.tree = msg.tree
		m.refreshTreeRows()
		return m, nil
	case metricsErrMsg:
		m.errMsg = msg.err.Error()
		return m, nil

	case renderLoadedMsg:
		if msg.gen == m.fetchGen {
			m.series = msg.series
			m.errMsg = ""
		}
		return m, nil
	case renderErrMsg:
		if msg.gen == m.fetchGen {
			m.errMsg = msg.err.Error()
		}
		return m, nil

	case autorefreshTickMsg:
		if !m.autorefreshOn {
			return m, nil
		}
		return m, tea.Batch(
			fetchRenderCmd(m.client, m.selectedMetrics, m.timeFrom, m.maxDataPoints, m.fetchGen),
			autorefreshTickCmd(m.autorefreshInterval),
		)

	case dashboardListLoadedMsg:
		m.dashboardNames = msg.names
		m.pickerCursor = 0
		return m, nil
	case dashboardLoadedMsg:
		m.currentDashboard = msg.d
		m.panelSeries = make([][]graphite.Series, len(msg.d.Panels))
		m.panelErrs = make([]error, len(msg.d.Panels))
		m.dashboardFocus = 0
		m.viewMode = viewDashboard
		m.popup = popupNone
		m.dashboardFetchGen++
		return m, fetchAllPanelsCmd(m.client, msg.d.Panels, m.dashboardFetchGen)
	case dashboardSavedMsg:
		m.popup = popupNone
		m.errMsg = ""
		return m, nil
	case dashboardErrMsg:
		m.errMsg = msg.err.Error()
		m.popup = popupNone
		return m, nil

	case panelRenderLoadedMsg:
		if msg.gen == m.dashboardFetchGen && msg.panelIndex < len(m.panelSeries) {
			m.panelSeries[msg.panelIndex] = msg.series
			m.panelErrs[msg.panelIndex] = nil
		}
		return m, nil
	case panelRenderErrMsg:
		if msg.gen == m.dashboardFetchGen && msg.panelIndex < len(m.panelErrs) {
			m.panelErrs[msg.panelIndex] = msg.err
		}
		return m, nil
	case dashboardAutorefreshTickMsg:
		if !m.dashboardAutorefreshOn || m.currentDashboard == nil {
			return m, nil
		}
		return m, tea.Batch(
			fetchAllPanelsCmd(m.client, m.currentDashboard.Panels, m.dashboardFetchGen),
			dashboardAutorefreshTickCmd(),
		)

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.popup != popupNone {
		return m.handlePopupKey(msg)
	}
	if m.viewMode == viewDashboard {
		return m.handleDashboardKey(msg)
	}
	return m.handleComposerKey(msg)
}
