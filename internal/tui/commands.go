package tui

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/benwtr/terphite/internal/dashboard"
	"github.com/benwtr/terphite/internal/graphite"
)

const requestTimeout = 15 * time.Second

// Pixel dimensions requested when fetching a rendered image. These don't
// need to precisely match the terminal display size — the image protocols
// scale to whatever cell box we tell them to display at — just be "high
// enough resolution" for a reasonably sized pane.
const (
	composerImageWidth  = 1000
	composerImageHeight = 500
	panelImageWidth     = 500
	panelImageHeight    = 300
)

func fetchMetricsCmd(client *graphite.Client) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
		defer cancel()
		names, err := client.FetchMetricsIndex(ctx)
		if err != nil {
			return metricsErrMsg{err: err}
		}
		return metricsLoadedMsg{tree: graphite.BuildTree(names)}
	}
}

func fetchRenderCmd(client *graphite.Client, targets []string, from string, maxDataPoints, gen int) tea.Cmd {
	targets = append([]string(nil), targets...)
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
		defer cancel()
		series, err := client.FetchRender(ctx, graphite.RenderQuery{
			Targets:       targets,
			From:          from,
			MaxDataPoints: maxDataPoints,
		})
		if err != nil {
			return renderErrMsg{err: err, gen: gen}
		}
		return renderLoadedMsg{series: series, gen: gen}
	}
}

func fetchImageCmd(client *graphite.Client, targets []string, from, areaMode string, gen int) tea.Cmd {
	targets = append([]string(nil), targets...)
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
		defer cancel()
		png, err := client.FetchRenderImage(ctx, graphite.ImageQuery{
			Targets:  targets,
			From:     from,
			Width:    composerImageWidth,
			Height:   composerImageHeight,
			AreaMode: areaMode,
		})
		if err != nil {
			return imageErrMsg{err: err, gen: gen}
		}
		return imageLoadedMsg{png: png, gen: gen}
	}
}

func autorefreshTickCmd(intervalSeconds int) tea.Cmd {
	return tea.Tick(time.Duration(intervalSeconds)*time.Second, func(time.Time) tea.Msg {
		return autorefreshTickMsg{}
	})
}

func dashboardAutorefreshTickCmd() tea.Cmd {
	return tea.Tick(defaultAutorefreshInterval*time.Second, func(time.Time) tea.Msg {
		return dashboardAutorefreshTickMsg{}
	})
}

func fetchDashboardListCmd(store *dashboard.Store) tea.Cmd {
	return func() tea.Msg {
		names, err := store.List()
		if err != nil {
			return dashboardErrMsg{err: err}
		}
		return dashboardListLoadedMsg{names: names}
	}
}

func loadDashboardCmd(store *dashboard.Store, name string) tea.Cmd {
	return func() tea.Msg {
		d, err := store.Load(name)
		if err != nil {
			return dashboardErrMsg{err: err}
		}
		return dashboardLoadedMsg{d: d}
	}
}

func saveDashboardCmd(store *dashboard.Store, d *dashboard.Dashboard) tea.Cmd {
	return func() tea.Msg {
		if err := store.Save(d); err != nil {
			return dashboardErrMsg{err: err}
		}
		return dashboardSavedMsg{}
	}
}

func saveDashboardPanelCmd(store *dashboard.Store, name string, panel dashboard.Panel) tea.Cmd {
	return func() tea.Msg {
		d, err := store.Load(name)
		if err != nil {
			d = &dashboard.Dashboard{Name: name}
		}
		d.Panels = append(d.Panels, panel)
		if err := store.Save(d); err != nil {
			return dashboardErrMsg{err: err}
		}
		return dashboardSavedMsg{}
	}
}

func fetchAllPanelsCmd(client *graphite.Client, panels []dashboard.Panel, gen int) tea.Cmd {
	cmds := make([]tea.Cmd, len(panels))
	for i, p := range panels {
		i, p := i, p
		cmds[i] = func() tea.Msg {
			ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
			defer cancel()
			series, err := client.FetchRender(ctx, graphite.RenderQuery{
				Targets:       p.Targets,
				From:          p.TimeFrom,
				MaxDataPoints: defaultMaxDataPoints,
			})
			if err != nil {
				return panelRenderErrMsg{panelIndex: i, err: err, gen: gen}
			}
			return panelRenderLoadedMsg{panelIndex: i, series: series, gen: gen}
		}
	}
	return tea.Batch(cmds...)
}

func fetchAllPanelImagesCmd(client *graphite.Client, panels []dashboard.Panel, gen int) tea.Cmd {
	cmds := make([]tea.Cmd, len(panels))
	for i, p := range panels {
		i, p := i, p
		cmds[i] = func() tea.Msg {
			ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
			defer cancel()
			png, err := client.FetchRenderImage(ctx, graphite.ImageQuery{
				Targets:  p.Targets,
				From:     p.TimeFrom,
				Width:    panelImageWidth,
				Height:   panelImageHeight,
				AreaMode: parseDrawMode(p.DrawMode).graphiteAreaMode(),
			})
			if err != nil {
				return panelImageErrMsg{panelIndex: i, err: err, gen: gen}
			}
			return panelImageLoadedMsg{panelIndex: i, png: png, gen: gen}
		}
	}
	return tea.Batch(cmds...)
}
