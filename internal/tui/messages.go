package tui

import (
	"github.com/benwtr/terphite/internal/dashboard"
	"github.com/benwtr/terphite/internal/graphite"
)

type metricsLoadedMsg struct{ tree *graphite.MetricNode }
type metricsErrMsg struct{ err error }

type renderLoadedMsg struct {
	series []graphite.Series
	gen    int
}
type renderErrMsg struct {
	err error
	gen int
}

type autorefreshTickMsg struct{}

type dashboardListLoadedMsg struct{ names []string }
type dashboardLoadedMsg struct{ d *dashboard.Dashboard }
type dashboardSavedMsg struct{}
type dashboardErrMsg struct{ err error }

type panelRenderLoadedMsg struct {
	panelIndex int
	series     []graphite.Series
	gen        int
}
type panelRenderErrMsg struct {
	panelIndex int
	err        error
	gen        int
}

type dashboardAutorefreshTickMsg struct{}
