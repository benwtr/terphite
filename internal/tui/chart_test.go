package tui

import (
	"strings"
	"testing"

	"github.com/benwtr/terphite/internal/graphite"
)

func val(v float64) *float64 { return &v }

func seriesOf(target string, values ...float64) graphite.Series {
	dps := make([]graphite.Datapoint, len(values))
	for i, v := range values {
		dps[i] = graphite.Datapoint{Value: val(v)}
	}
	return graphite.Series{Target: target, Datapoints: dps}
}

func TestSubXForSpreadsEvenly(t *testing.T) {
	if got := subXFor(0, 5, 100); got != 0 {
		t.Errorf("subXFor(0,5,100) = %d, want 0", got)
	}
	if got := subXFor(4, 5, 100); got != 99 {
		t.Errorf("subXFor(4,5,100) = %d, want 99 (last sample maps to last column)", got)
	}
	if got := subXFor(0, 1, 100); got != 0 {
		t.Errorf("subXFor(0,1,100) = %d, want 0 (single sample doesn't divide by zero)", got)
	}
}

func TestSeriesPointsSkipsNilDatapoints(t *testing.T) {
	s := graphite.Series{Datapoints: []graphite.Datapoint{
		{Value: val(1)},
		{Value: nil},
		{Value: val(2)},
	}}
	identity := func(v float64) int { return int(v) }
	pts := seriesPoints(s, identity, 10)
	if len(pts) != 2 {
		t.Fatalf("got %d points, want 2 (nil skipped)", len(pts))
	}
	if pts[0].subY != 1 || pts[1].subY != 2 {
		t.Errorf("points = %+v, want subY 1 then 2", pts)
	}
}

func TestStackedCumulativeSum(t *testing.T) {
	series := []graphite.Series{
		seriesOf("a", 1, 2, 3),
		seriesOf("b", 10, 10, 10),
	}
	cum := stackedCumulative(series)
	if len(cum) != 2 {
		t.Fatalf("got %d series in cumulative, want 2", len(cum))
	}
	wantA := []float64{1, 2, 3}
	wantB := []float64{11, 12, 13}
	for i := range wantA {
		if cum[0][i] != wantA[i] {
			t.Errorf("cum[0][%d] = %v, want %v", i, cum[0][i], wantA[i])
		}
		if cum[1][i] != wantB[i] {
			t.Errorf("cum[1][%d] = %v, want %v", i, cum[1][i], wantB[i])
		}
	}
}

func TestStackedCumulativeTreatsNilAsZero(t *testing.T) {
	series := []graphite.Series{
		{Datapoints: []graphite.Datapoint{{Value: val(5)}, {Value: nil}}},
		{Datapoints: []graphite.Datapoint{{Value: val(1)}, {Value: val(1)}}},
	}
	cum := stackedCumulative(series)
	if cum[1][1] != 1 {
		t.Errorf("cum[1][1] = %v, want 1 (nil in series 0 contributes 0)", cum[1][1])
	}
}

func TestStackedRangeUsesTopOfCumulative(t *testing.T) {
	series := []graphite.Series{
		seriesOf("a", 1, 5),
		seriesOf("b", 2, 2),
	}
	min, max := stackedRange(series)
	if min != 0 {
		t.Errorf("min = %v, want 0", min)
	}
	if max != 7 { // max of cumulative totals: (1+2)=3, (5+2)=7
		t.Errorf("max = %v, want 7", max)
	}
}

func TestStackedRangeEmptySeries(t *testing.T) {
	min, max := stackedRange(nil)
	if min != 0 || max != 1 {
		t.Errorf("got (%v, %v), want (0, 1) for empty input", min, max)
	}
}

func TestRenderChartNoData(t *testing.T) {
	out := renderChart(nil, drawLine, 40, 10)
	if !strings.Contains(out, "no data") {
		t.Errorf("expected 'no data' placeholder, got %q", out)
	}
}

func TestRenderChartTooSmall(t *testing.T) {
	series := []graphite.Series{seriesOf("a", 1, 2, 3)}
	if out := renderChart(series, drawLine, 2, 1); out != "" {
		t.Errorf("expected empty string for too-small dimensions, got %q", out)
	}
}

func TestRenderChartIncludesLegend(t *testing.T) {
	series := []graphite.Series{seriesOf("stats.foo", 1, 2, 3)}
	out := renderChart(series, drawLine, 40, 12)
	if !strings.Contains(out, "stats.foo") {
		t.Errorf("expected legend to contain target name, got:\n%s", out)
	}
}

func TestRenderChartAllDrawModesProduceOutput(t *testing.T) {
	series := []graphite.Series{
		seriesOf("a", 1, 3, 2, 5, 4),
		seriesOf("b", 2, 2, 3, 1, 2),
	}
	for _, mode := range []drawMode{drawLine, drawArea, drawStacked} {
		out := renderChart(series, mode, 40, 12)
		if out == "" {
			t.Errorf("mode %v: expected non-empty chart output", mode)
		}
		if !strings.Contains(out, "a") || !strings.Contains(out, "b") {
			t.Errorf("mode %v: expected legend entries for both series", mode)
		}
	}
}
