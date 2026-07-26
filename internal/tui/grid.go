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

// renderDashboardGrid lays the panels out as a grid and returns it along
// with any inline-image overlay escapes, which the caller appends after the
// whole frame. topOffset is the 0-indexed frame line the grid starts on,
// used to place those overlays at absolute terminal coordinates.
func renderDashboardGrid(panels []dashboardPanelState, focus, width, height, topOffset int) (string, string) {
	if len(panels) == 0 {
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center,
			faintStyle.Render("this dashboard has no panels yet — press S in composer view to add one")), ""
	}

	cols := gridColumns(len(panels))
	rows := int(math.Ceil(float64(len(panels)) / float64(cols)))
	tileWidth := clampMin(width/cols, 12)
	tileHeight := clampMin(height/rows, 6)

	var overlays strings.Builder
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
			overlays.WriteString(panelImageOverlay(panels[idx], r, c, tileWidth, tileHeight, topOffset))
		}
		gridLines = append(gridLines, lipgloss.JoinHorizontal(lipgloss.Top, tiles...))
	}
	return lipgloss.JoinVertical(lipgloss.Left, gridLines...), overlays.String()
}

// panelImageOverlay places a tile's image over its reserved blank area.
// Inside a tile the border takes one row and the title another, so the
// image body starts two rows down and one column in (all 1-indexed).
func panelImageOverlay(p dashboardPanelState, r, c, tileWidth, tileHeight, topOffset int) string {
	if !p.imageActive() {
		return ""
	}
	imgCols := clampMin(tileWidth-2, 1)
	imgRows := clampMin(tileHeight-3, 1)
	row := topOffset + r*tileHeight + 3
	col := c*tileWidth + 2
	return overlayAt(row, col, p.ImageCache.escape(p.ImageProto, p.ImageBytes, p.ImageVersion, imgCols, imgRows))
}

// imageActive reports whether this panel should show a rendered image
// rather than the ASCII chart.
func (p dashboardPanelState) imageActive() bool {
	return p.Err == nil &&
		p.ImageProto != termimg.ProtocolNone &&
		len(p.ImageBytes) > 0 &&
		p.ImageCache != nil
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

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(tileBorderColor).
		Width(clampMin(width-2, 1)).
		Height(clampMin(height-2, 1))

	var body string
	switch {
	case p.Err != nil:
		body = errorStyle.Render("error: " + p.Err.Error())
	case p.imageActive():
		// Reserve blank space; panelImageOverlay paints the image over it.
		body = ""
	default:
		body = renderChart(p.Series, p.Mode, innerWidth, innerHeight)
	}
	return style.Render(title + "\n" + body)
}
