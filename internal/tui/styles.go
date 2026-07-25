package tui

import "github.com/charmbracelet/lipgloss"

var (
	borderStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("4"))
	borderColor = lipgloss.NewStyle().Foreground(lipgloss.Color("4"))
	errorStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true)
	faintStyle  = lipgloss.NewStyle().Faint(true)
)

func clampMin(n, min int) int {
	if n < min {
		return min
	}
	return n
}
