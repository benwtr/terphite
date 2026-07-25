package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/benwtr/terphite/internal/graphite"
)

type treeRow struct {
	node  *graphite.MetricNode
	depth int
}

func flattenTree(root *graphite.MetricNode, expanded map[string]bool) []treeRow {
	var rows []treeRow
	var walk func(n *graphite.MetricNode, depth int)
	walk = func(n *graphite.MetricNode, depth int) {
		for _, c := range n.Children {
			rows = append(rows, treeRow{node: c, depth: depth})
			if len(c.Children) > 0 && expanded[c.Path] {
				walk(c, depth+1)
			}
		}
	}
	if root != nil {
		walk(root, 0)
	}
	return rows
}

func renderTree(rows []treeRow, cursor int, expanded, selected map[string]bool, width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}
	if len(rows) == 0 {
		return faintStyle.Render("loading metrics…")
	}

	start := 0
	if cursor >= height {
		start = cursor - height + 1
	}
	end := start + height
	if end > len(rows) {
		end = len(rows)
	}

	lines := make([]string, 0, end-start)
	for i := start; i < end; i++ {
		row := rows[i]
		indent := strings.Repeat("  ", row.depth)

		marker := "  "
		if len(row.node.Children) > 0 {
			if expanded[row.node.Path] {
				marker = "▾ "
			} else {
				marker = "▸ "
			}
		}

		label := row.node.Name
		if row.node.Leaf {
			if selected[row.node.Path] {
				label = "[x] " + label
			} else {
				label = "[ ] " + label
			}
		}

		line := indent + marker + label
		if len(line) > width {
			line = line[:width]
		}

		style := lipgloss.NewStyle().Width(width)
		if row.node.Leaf && selected[row.node.Path] {
			style = style.Bold(true)
		}
		if i == cursor {
			style = style.Reverse(true)
		}
		lines = append(lines, style.Render(line))
	}
	return strings.Join(lines, "\n")
}
