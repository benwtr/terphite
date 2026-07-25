package tui

import (
	"math"

	"github.com/charmbracelet/lipgloss"

	"github.com/benwtr/terphite/internal/graphite"
)

type dashboardPanelState struct {
	Title  string
	Series []graphite.Series
	Err    error
}

func gridColumns(n int) int {
	if n <= 1 {
		return 1
	}
	return int(math.Ceil(math.Sqrt(float64(n))))
}

func renderDashboardGrid(panels []dashboardPanelState, focus, width, height int) string {
	if len(panels) == 0 {
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center,
			faintStyle.Render("this dashboard has no panels yet — press S in composer view to add one"))
	}

	cols := gridColumns(len(panels))
	rows := int(math.Ceil(float64(len(panels)) / float64(cols)))
	tileWidth := clampMin(width/cols, 12)
	tileHeight := clampMin(height/rows, 6)

	gridLines := make([]string, 0, rows)
	for r := 0; r < rows; r++ {
		tiles := make([]string, 0, cols)
		for c := 0; c < cols; c++ {
			idx := r*cols + c
			if idx >= len(panels) {
				tiles = append(tiles, lipgloss.NewStyle().Width(tileWidth).Height(tileHeight).Render(""))
				continue
			}
			tiles = append(tiles, renderPanelTile(panels[idx], idx == focus, tileWidth, tileHeight))
		}
		gridLines = append(gridLines, lipgloss.JoinHorizontal(lipgloss.Top, tiles...))
	}
	return lipgloss.JoinVertical(lipgloss.Left, gridLines...)
}

func renderPanelTile(p dashboardPanelState, focused bool, width, height int) string {
	borderColor := lipgloss.Color("8")
	if focused {
		borderColor = lipgloss.Color("4")
	}
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(clampMin(width-2, 1)).
		Height(clampMin(height-2, 1))

	innerWidth := clampMin(width-4, 1)
	innerHeight := clampMin(height-5, 1)

	title := p.Title
	if len(title) > innerWidth {
		title = title[:innerWidth]
	}

	var body string
	if p.Err != nil {
		body = errorStyle.Render("error: " + p.Err.Error())
	} else {
		body = renderChart(p.Series, innerWidth, innerHeight)
	}

	return style.Render(title + "\n" + body)
}
