package tui

import "testing"

func TestBrailleCanvasSetSingleDot(t *testing.T) {
	c := newBrailleCanvas(2, 1) // 4 sub-cols x 4 sub-rows
	c.set(0, 0, 0)
	if c.bits[0][0] != 0x01 {
		t.Errorf("bits[0][0] = %#x, want 0x01 (top-left dot)", c.bits[0][0])
	}
	if c.colorIdx[0][0] != 0 {
		t.Errorf("colorIdx[0][0] = %d, want 0", c.colorIdx[0][0])
	}
}

func TestBrailleCanvasSetAllDotsInCell(t *testing.T) {
	c := newBrailleCanvas(1, 1)
	for x := 0; x < 2; x++ {
		for y := 0; y < 4; y++ {
			c.set(x, y, 0)
		}
	}
	if c.bits[0][0] != 0xFF {
		t.Errorf("bits[0][0] = %#x, want 0xff (all 8 dots lit)", c.bits[0][0])
	}
	// render() wraps non-blank cells in ANSI color codes, so just check the
	// expected rune (all 8 dots) appears somewhere in the styled output.
	lines := c.render()
	if !containsRune(lines[0], rune(brailleBlank+0xFF)) {
		t.Errorf("render()[0] = %q, want it to contain the all-dots-lit rune", lines[0])
	}
}

func TestBrailleCanvasSetOutOfBoundsIgnored(t *testing.T) {
	c := newBrailleCanvas(1, 1)
	c.set(-1, 0, 0)
	c.set(0, -1, 0)
	c.set(100, 0, 0)
	c.set(0, 100, 0)
	if c.bits[0][0] != 0 {
		t.Errorf("bits[0][0] = %#x, want 0 (all sets were out of bounds)", c.bits[0][0])
	}
}

func TestBrailleCanvasBlankCellRendersSpace(t *testing.T) {
	c := newBrailleCanvas(3, 1)
	lines := c.render()
	if lines[0] != "   " {
		t.Errorf("render()[0] = %q, want three spaces", lines[0])
	}
}

func TestBrailleCanvasLineHorizontal(t *testing.T) {
	c := newBrailleCanvas(4, 1) // 8 sub-cols
	c.line(0, 0, 7, 0, 0)
	for x := 0; x < 8; x++ {
		col := x / 2
		if c.bits[0][col] == 0 {
			t.Errorf("expected sub-col %d (cell col %d) to be lit by horizontal line", x, col)
		}
	}
}

func TestBrailleCanvasLineColorTagging(t *testing.T) {
	c := newBrailleCanvas(2, 1)
	c.line(0, 0, 3, 3, 2)
	if c.colorIdx[0][0] != 2 {
		t.Errorf("colorIdx[0][0] = %d, want 2", c.colorIdx[0][0])
	}
}

func TestFillColumn(t *testing.T) {
	c := newBrailleCanvas(1, 1) // 2 sub-cols x 4 sub-rows
	c.fillColumn(0, 3, 0, 0)    // reversed order should still fill 0..3
	for y := 0; y < 4; y++ {
		if c.bits[0][0]&brailleBits[0][y] == 0 {
			t.Errorf("expected sub-row %d in sub-col 0 to be filled", y)
		}
	}
	// sub-col 1 (the other column in this cell) should be untouched.
	for y := 0; y < 4; y++ {
		if c.bits[0][0]&brailleBits[1][y] != 0 {
			t.Errorf("sub-col 1, row %d should not be filled", y)
		}
	}
}

func containsRune(s string, r rune) bool {
	for _, c := range s {
		if c == r {
			return true
		}
	}
	return false
}
