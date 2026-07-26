package tui

import "github.com/charmbracelet/bubbles/textinput"

func newTextInput(prompt, value string, width int) textinput.Model {
	ti := textinput.New()
	ti.Prompt = prompt
	ti.SetValue(value)
	ti.CursorEnd()
	ti.Width = width
	ti.Focus()
	return ti
}
