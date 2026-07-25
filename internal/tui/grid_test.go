package tui

import "testing"

func TestGridColumns(t *testing.T) {
	cases := []struct {
		n    int
		want int
	}{
		{0, 1},
		{1, 1},
		{2, 2},
		{3, 2},
		{4, 2},
		{5, 3},
		{9, 3},
		{10, 4},
	}
	for _, c := range cases {
		got := gridColumns(c.n)
		if got != c.want {
			t.Errorf("gridColumns(%d) = %d, want %d", c.n, got, c.want)
		}
	}
}
