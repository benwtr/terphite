package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// brailleBits maps a sub-pixel's (column, row) position within a braille
// character cell to its dot bit, using the standard drawille layout:
//
//	col0,row0=0x01  col1,row0=0x08
//	col0,row1=0x02  col1,row1=0x10
//	col0,row2=0x04  col1,row2=0x20
//	col0,row3=0x40  col1,row3=0x80
var brailleBits = [2][4]byte{
	{0x01, 0x02, 0x04, 0x40},
	{0x08, 0x10, 0x20, 0x80},
}

// brailleBlank is the "all dots off" braille character.
const brailleBlank = 0x2800

// brailleCanvas is a plotting surface addressed in sub-pixels (2 columns x
// 4 rows per character cell), giving much finer resolution than plotting
// one point per cell. Each cell also tracks which series last touched it,
// since a terminal cell can only have one foreground color regardless of
// how many series' dots land in it.
type brailleCanvas struct {
	cols, rows       int // character-cell dimensions
	subCols, subRows int // sub-pixel dimensions (cols*2, rows*4)
	bits             [][]byte
	colorIdx         [][]int
}

func newBrailleCanvas(cols, rows int) *brailleCanvas {
	if cols < 1 {
		cols = 1
	}
	if rows < 1 {
		rows = 1
	}
	c := &brailleCanvas{
		cols: cols, rows: rows,
		subCols: cols * 2, subRows: rows * 4,
	}
	c.bits = make([][]byte, rows)
	c.colorIdx = make([][]int, rows)
	for r := 0; r < rows; r++ {
		c.bits[r] = make([]byte, cols)
		c.colorIdx[r] = make([]int, cols)
		for cc := range c.colorIdx[r] {
			c.colorIdx[r][cc] = -1
		}
	}
	return c
}

// set lights the sub-pixel at (subX, subY), tagging its cell with
// seriesIdx. Out-of-bounds coordinates are silently ignored.
func (c *brailleCanvas) set(subX, subY, seriesIdx int) {
	if subX < 0 || subY < 0 || subX >= c.subCols || subY >= c.subRows {
		return
	}
	cellCol, cellRow := subX/2, subY/4
	bitCol, bitRow := subX%2, subY%4
	c.bits[cellRow][cellCol] |= brailleBits[bitCol][bitRow]
	c.colorIdx[cellRow][cellCol] = seriesIdx
}

// line draws a Bresenham line between two sub-pixel points, coloring every
// touched cell with seriesIdx.
func (c *brailleCanvas) line(x0, y0, x1, y1, seriesIdx int) {
	dx := absInt(x1 - x0)
	sx := -1
	if x0 < x1 {
		sx = 1
	}
	dy := -absInt(y1 - y0)
	sy := -1
	if y0 < y1 {
		sy = 1
	}
	err := dx + dy
	x, y := x0, y0
	for {
		c.set(x, y, seriesIdx)
		if x == x1 && y == y1 {
			return
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x += sx
		}
		if e2 <= dx {
			err += dx
			y += sy
		}
	}
}

// fillColumn lights every sub-pixel row between y0 and y1 (inclusive, order
// independent) in sub-column subX.
func (c *brailleCanvas) fillColumn(subX, y0, y1, seriesIdx int) {
	if y0 > y1 {
		y0, y1 = y1, y0
	}
	for y := y0; y <= y1; y++ {
		c.set(subX, y, seriesIdx)
	}
}

// render converts the canvas into one string per character row, with each
// non-blank cell colored by its tagged series.
func (c *brailleCanvas) render() []string {
	lines := make([]string, c.rows)
	for r := 0; r < c.rows; r++ {
		var b strings.Builder
		for cc := 0; cc < c.cols; cc++ {
			bits := c.bits[r][cc]
			if bits == 0 {
				b.WriteByte(' ')
				continue
			}
			ch := rune(brailleBlank + int(bits))
			style := lipgloss.NewStyle().Foreground(seriesColor(c.colorIdx[r][cc]))
			b.WriteString(style.Render(string(ch)))
		}
		lines[r] = b.String()
	}
	return lines
}

func absInt(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
