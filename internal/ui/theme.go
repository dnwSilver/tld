package ui

import "github.com/charmbracelet/lipgloss"

type Palette struct {
	Background lipgloss.Color
	Hover      lipgloss.Color
	Primary    lipgloss.Color
	Hint       lipgloss.Color
	Disable    lipgloss.Color
	Error      lipgloss.Color
	Warning    lipgloss.Color
	Info       lipgloss.Color
}

func NewPalette() Palette {
	return Palette{
		Background: lipgloss.Color("#000000"),
		Hover:      lipgloss.Color("#292928"),
		Primary:    lipgloss.Color("#84BA64"),
		Hint:       lipgloss.Color("#71797E"),
		Disable:    lipgloss.Color("#303030"),
		Error:      lipgloss.Color("#CA3433"),
		Warning:    lipgloss.Color("#EC9706"),
		Info:       lipgloss.Color("#25799F"),
	}
}
