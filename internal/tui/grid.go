package tui

import (
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/benwtr/terphite/internal/graphite"
	"github.com/benwtr/terphite/internal/termimg"
)

type dashboardPanelState struct {
	Title  string
	Series []graphite.Series
	Mode   drawMode
	Err    error

	ImageProto   termimg.Protocol
	ImageBytes   []byte
	ImageVersion int
	ImageCache   *imageEscapeCache
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
	tileBorderColor := lipgloss.Color("8")
	if focused {
		tileBorderColor = lipgloss.Color("4")
	}

	innerWidth := clampMin(width-4, 1)
	innerHeight := clampMin(height-5, 1)

	title := p.Title
	if len(title) > innerWidth {
		title = title[:innerWidth]
	}

	if p.Err == nil && p.ImageProto != termimg.ProtocolNone && len(p.ImageBytes) > 0 && p.ImageCache != nil {
		if img := p.ImageCache.escape(p.ImageProto, p.ImageBytes, p.ImageVersion, innerWidth, innerHeight); img != "" {
			return renderImageTile(title, img, tileBorderColor, width)
		}
	}

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(tileBorderColor).
		Width(clampMin(width-2, 1)).
		Height(clampMin(height-2, 1))

	var body string
	if p.Err != nil {
		body = errorStyle.Render("error: " + p.Err.Error())
	} else {
		body = renderChart(p.Series, p.Mode, innerWidth, innerHeight)
	}
	return style.Render(title + "\n" + body)
}

// renderImageTile frames an inline image with hand-drawn top/bottom bars
// rather than lipgloss's Border(), since the image protocols emit one
// opaque multi-row block that can't have left/right border characters
// interleaved on its own rows (see renderChartPane in view.go).
func renderImageTile(title, img string, color lipgloss.Color, width int) string {
	colorStyle := lipgloss.NewStyle().Foreground(color)
	bar := strings.Repeat("─", clampMin(width-2, 0))
	top := colorStyle.Render("╭" + bar + "╮")
	bottom := colorStyle.Render("╰" + bar + "╯")
	titleLine := colorStyle.Render("│") + " " + title
	return top + "\n" + titleLine + "\n" + img + "\n" + bottom
}
