package tui

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/benwtr/terphite/internal/graphite"
)

// seriesColors mirrors the original's 15-color cycle: reds/greens/etc. at
// standard intensity, then their "light" (bright) ANSI variants.
var seriesColors = []string{
	"1", "2", "3", "4", "5", "6", "7",
	"8", "9", "10", "11", "12", "13", "14", "15",
}

func seriesColor(i int) lipgloss.Color {
	return lipgloss.Color(seriesColors[i%len(seriesColors)])
}

// renderChart renders series as a simple point-plot line chart with a
// legend beneath it, scaled to fit width x height.
func renderChart(series []graphite.Series, width, height int) string {
	if width < 10 || height < 3 {
		return ""
	}

	legendHeight := len(series)
	if max := height / 2; legendHeight > max {
		legendHeight = max
	}
	plotHeight := height - legendHeight - 1
	if plotHeight < 1 {
		plotHeight = 1
	}

	hasData := false
	var min, max float64
	for _, s := range series {
		for _, dp := range s.Datapoints {
			if dp.Value == nil {
				continue
			}
			v := *dp.Value
			if !hasData || v < min {
				min = v
			}
			if !hasData || v > max {
				max = v
			}
			hasData = true
		}
	}
	if !hasData {
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, faintStyle.Render("no data"))
	}
	if max == min {
		max = min + 1
	}

	grid := make([][]rune, plotHeight)
	colorIdx := make([][]int, plotHeight)
	for i := range grid {
		grid[i] = make([]rune, width)
		colorIdx[i] = make([]int, width)
		for j := range grid[i] {
			grid[i][j] = ' '
			colorIdx[i][j] = -1
		}
	}

	for si, s := range series {
		n := len(s.Datapoints)
		if n == 0 {
			continue
		}
		cols := width
		if n < cols {
			cols = n
		}
		for c := 0; c < cols; c++ {
			idx := c
			if cols > 1 {
				idx = c * (n - 1) / (cols - 1)
			}
			dp := s.Datapoints[idx]
			if dp.Value == nil {
				continue
			}
			frac := (*dp.Value - min) / (max - min)
			row := plotHeight - 1 - int(math.Round(frac*float64(plotHeight-1)))
			row = clampMin(row, 0)
			if row >= plotHeight {
				row = plotHeight - 1
			}
			col := c * width / cols
			if col >= width {
				col = width - 1
			}
			grid[row][col] = '●'
			colorIdx[row][col] = si
		}
	}

	lines := make([]string, 0, height)
	for r := 0; r < plotHeight; r++ {
		var b strings.Builder
		for c := 0; c < width; c++ {
			if colorIdx[r][c] < 0 {
				b.WriteRune(grid[r][c])
				continue
			}
			style := lipgloss.NewStyle().Foreground(seriesColor(colorIdx[r][c]))
			b.WriteString(style.Render(string(grid[r][c])))
		}
		lines = append(lines, b.String())
	}

	lines = append(lines, strings.Repeat("─", width))
	for i, s := range series {
		if i >= legendHeight {
			break
		}
		swatch := lipgloss.NewStyle().Foreground(seriesColor(i)).Render("●")
		title := s.Target
		if len(title) > width-4 && width > 4 {
			title = title[:width-4]
		}
		lines = append(lines, fmt.Sprintf("%s %s", swatch, title))
	}

	return strings.Join(lines, "\n")
}
