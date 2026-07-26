package tui

import (
	"testing"

	"github.com/benwtr/terphite/internal/graphite"
	"github.com/benwtr/terphite/internal/timerange"
)

func newTestModel(t *testing.T) *Model {
	t.Helper()
	m, err := New(Config{
		GraphiteURI:  "http://example.com",
		DashboardDir: t.TempDir(),
	})
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func TestToggleMetric(t *testing.T) {
	m := newTestModel(t)
	m.toggleMetric("stats.foo")
	if len(m.selectedMetrics) != 1 || m.selectedMetrics[0] != "stats.foo" {
		t.Fatalf("after add: %v", m.selectedMetrics)
	}
	m.toggleMetric("stats.foo")
	if len(m.selectedMetrics) != 0 {
		t.Fatalf("after remove: %v", m.selectedMetrics)
	}
}

func TestAdjustTimeMinuteClampsToDefault(t *testing.T) {
	m := newTestModel(t)
	m.timeFrom = timerange.Default // -1min, exactly at the floor
	m.adjustTime(-timerange.UnitMinute, timerange.UnitMinute, timerange.Default)
	if m.timeFrom != timerange.Default {
		t.Errorf("timeFrom = %q, want %q", m.timeFrom, timerange.Default)
	}
}

func TestAdjustTimeMinuteIncreases(t *testing.T) {
	m := newTestModel(t)
	m.timeFrom = "-5min"
	m.adjustTime(timerange.UnitMinute, 0, "")
	if m.timeFrom != "-6min" {
		t.Errorf("timeFrom = %q, want -6min", m.timeFrom)
	}
}

func TestAdjustTimeHourClampsToOneHour(t *testing.T) {
	m := newTestModel(t)
	m.timeFrom = "-30min"
	m.adjustTime(-timerange.UnitHour, timerange.UnitHour, "-1h")
	if m.timeFrom != "-1h" {
		t.Errorf("timeFrom = %q, want -1h", m.timeFrom)
	}
}

func TestRefreshTreeRowsClampsCursor(t *testing.T) {
	m := newTestModel(t)
	m.tree = graphite.BuildTree([]string{"a", "b", "c"})
	m.treeCursor = 100
	m.refreshTreeRows()
	if m.treeCursor != len(m.treeRows)-1 {
		t.Errorf("treeCursor = %d, want %d", m.treeCursor, len(m.treeRows)-1)
	}
}

func TestPanelTitle(t *testing.T) {
	cases := []struct {
		targets []string
		want    string
	}{
		{nil, "(no targets)"},
		{[]string{"stats.foo"}, "stats.foo"},
		{[]string{"stats.foo", "stats.bar"}, "stats.foo +1 more"},
	}
	for _, c := range cases {
		got := panelTitle(c.targets)
		if got != c.want {
			t.Errorf("panelTitle(%v) = %q, want %q", c.targets, got, c.want)
		}
	}
}
