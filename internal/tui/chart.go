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

// renderChart renders series as a chart with a legend beneath it, scaled to
// fit width x height. mode selects independent lines, independent filled
// areas, or a cumulative stacked area.
func renderChart(series []graphite.Series, mode drawMode, width, height int) string {
	if width < 4 || height < 3 {
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

	if mode == drawStacked {
		// Stacked bands are built from cumulative sums starting at 0, so 0
		// must be part of the visible range.
		min, max = stackedRange(series)
	}
	if max == min {
		max = min + 1
	}

	canvas := newBrailleCanvas(width, plotHeight)
	subCols := canvas.subCols

	valueToSubY := func(v float64) int {
		frac := (v - min) / (max - min)
		y := canvas.subRows - 1 - int(math.Round(frac*float64(canvas.subRows-1)))
		return clampInt(y, 0, canvas.subRows-1)
	}

	switch mode {
	case drawStacked:
		renderStacked(canvas, series, valueToSubY, subCols)
	case drawArea:
		renderArea(canvas, series, valueToSubY, subCols)
	default:
		renderLines(canvas, series, valueToSubY, subCols)
	}

	lines := canvas.render()
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

// point is a plotted sample in sub-pixel canvas coordinates.
type point struct {
	subX, subY int
}

// subXFor maps data-index i (of n total) onto a sub-pixel column, spreading
// samples evenly across the available width.
func subXFor(i, n, subCols int) int {
	if n <= 1 || subCols <= 1 {
		return 0
	}
	return i * (subCols - 1) / (n - 1)
}

// seriesPoints converts s's non-nil datapoints into sub-pixel points. Nil
// datapoints are skipped (matching the original renderer's behavior), so a
// run of points on either side of a gap is connected directly across it.
func seriesPoints(s graphite.Series, valueToSubY func(float64) int, subCols int) []point {
	n := len(s.Datapoints)
	pts := make([]point, 0, n)
	for i, dp := range s.Datapoints {
		if dp.Value == nil {
			continue
		}
		pts = append(pts, point{subX: subXFor(i, n, subCols), subY: valueToSubY(*dp.Value)})
	}
	return pts
}

func renderLines(canvas *brailleCanvas, series []graphite.Series, valueToSubY func(float64) int, subCols int) {
	for si, s := range series {
		pts := seriesPoints(s, valueToSubY, subCols)
		for i, p := range pts {
			if i == 0 {
				canvas.set(p.subX, p.subY, si)
				continue
			}
			prev := pts[i-1]
			canvas.line(prev.subX, prev.subY, p.subX, p.subY, si)
		}
	}
}

// renderArea fills each series independently down to the bottom of the
// visible plot (not to value-zero — the data may float well above or below
// zero, and filling to zero in that case would just paint the whole plot
// solid). Later series are drawn over earlier ones where they overlap.
func renderArea(canvas *brailleCanvas, series []graphite.Series, valueToSubY func(float64) int, subCols int) {
	baseline := canvas.subRows - 1
	for si, s := range series {
		pts := seriesPoints(s, valueToSubY, subCols)
		fillUnderCurve(canvas, pts, baseline, si)
	}
}

// fillUnderCurve fills, for every sub-column spanned by pts, from baseline
// to the curve (linearly interpolated between samples where they're more
// than one sub-column apart).
func fillUnderCurve(canvas *brailleCanvas, pts []point, baseline, seriesIdx int) {
	switch len(pts) {
	case 0:
		return
	case 1:
		canvas.fillColumn(pts[0].subX, baseline, pts[0].subY, seriesIdx)
		return
	}
	for i := 1; i < len(pts); i++ {
		a, b := pts[i-1], pts[i]
		if b.subX == a.subX {
			canvas.fillColumn(a.subX, baseline, a.subY, seriesIdx)
			continue
		}
		for x := a.subX; x <= b.subX; x++ {
			t := float64(x-a.subX) / float64(b.subX-a.subX)
			y := a.subY + int(math.Round(t*float64(b.subY-a.subY)))
			canvas.fillColumn(x, baseline, y, seriesIdx)
		}
	}
}

// stackedCumulative returns, for each series in order, the running total of
// that series plus every series before it (nil datapoints count as 0, so
// stacking has a well-defined value at every index). All series are aligned
// by datapoint index and truncated to the shortest series' length — in
// practice they come from the same render request so lengths match.
func stackedCumulative(series []graphite.Series) [][]float64 {
	if len(series) == 0 {
		return nil
	}
	n := len(series[0].Datapoints)
	for _, s := range series {
		if len(s.Datapoints) < n {
			n = len(s.Datapoints)
		}
	}

	cum := make([][]float64, len(series))
	running := make([]float64, n)
	for si, s := range series {
		cum[si] = make([]float64, n)
		for i := 0; i < n; i++ {
			v := 0.0
			if s.Datapoints[i].Value != nil {
				v = *s.Datapoints[i].Value
			}
			running[i] += v
			cum[si][i] = running[i]
		}
	}
	return cum
}

// stackedRange returns the value range a stacked chart needs: 0 to the
// tallest cumulative total.
func stackedRange(series []graphite.Series) (min, max float64) {
	cum := stackedCumulative(series)
	if len(cum) == 0 {
		return 0, 1
	}
	top := cum[len(cum)-1]
	for _, v := range top {
		if v > max {
			max = v
		}
	}
	return 0, max
}

func renderStacked(canvas *brailleCanvas, series []graphite.Series, valueToSubY func(float64) int, subCols int) {
	cum := stackedCumulative(series)
	if len(cum) == 0 {
		return
	}
	n := len(cum[0])

	prevCum := make([]float64, n)
	for si := range series {
		lower := make([]point, n)
		upper := make([]point, n)
		for i := 0; i < n; i++ {
			x := subXFor(i, n, subCols)
			lower[i] = point{subX: x, subY: valueToSubY(prevCum[i])}
			upper[i] = point{subX: x, subY: valueToSubY(cum[si][i])}
		}
		fillBand(canvas, lower, upper, si)
		prevCum = cum[si]
	}
}

// fillBand fills the region between the lower and upper bound lines
// (interpolating both between samples), used for stacked areas.
func fillBand(canvas *brailleCanvas, lower, upper []point, seriesIdx int) {
	n := len(lower)
	switch n {
	case 0:
		return
	case 1:
		canvas.fillColumn(lower[0].subX, lower[0].subY, upper[0].subY, seriesIdx)
		return
	}
	for i := 1; i < n; i++ {
		x0, x1 := lower[i-1].subX, lower[i].subX
		if x1 == x0 {
			canvas.fillColumn(x0, lower[i-1].subY, upper[i-1].subY, seriesIdx)
			continue
		}
		for x := x0; x <= x1; x++ {
			t := float64(x-x0) / float64(x1-x0)
			ly := lower[i-1].subY + int(math.Round(t*float64(lower[i].subY-lower[i-1].subY)))
			uy := upper[i-1].subY + int(math.Round(t*float64(upper[i].subY-upper[i-1].subY)))
			canvas.fillColumn(x, ly, uy, seriesIdx)
		}
	}
}

func clampInt(n, lo, hi int) int {
	if n < lo {
		return lo
	}
	if n > hi {
		return hi
	}
	return n
}
