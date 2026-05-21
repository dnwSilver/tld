package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type DefaultScreen struct {
	palette Palette
}

func NewDefaultScreen(palette Palette) DefaultScreen {
	return DefaultScreen{
		palette: palette,
	}
}

func (s DefaultScreen) Render(width, height int) string {
	boxWidth := max(width, 2)
	boxHeight := max(height-1, 3)
	contentWidth := max(boxWidth-2, 1)
	contentHeight := max(boxHeight-2, 1)

	content := s.renderContent(contentWidth, contentHeight)

	box := lipgloss.NewStyle().
		Background(s.palette.Background).
		Width(contentWidth).
		Height(contentHeight).
		Border(lipgloss.NormalBorder()).
		BorderForeground(s.palette.Primary).
		Render(content)

	footer := lipgloss.NewStyle().
		Width(width).
		Background(s.palette.Background).
		Foreground(s.palette.Hint).
		Align(lipgloss.Right).
		Render(symbolCopyright + " Kolosov Aleksandr")

	return lipgloss.JoinVertical(lipgloss.Left, box, footer)
}

func (s DefaultScreen) renderContent(width, height int) string {
	title := lipgloss.NewStyle().
		Background(s.palette.Background).
		Foreground(s.palette.Warning).
		Bold(true).
		Render(symbolDashboard + " Team lead dashboard")
	titleWidth := lipgloss.Width(title)
	titleRow := height / 2
	titleLeft := max((width-titleWidth)/2, 0)
	titleRight := max(width-titleLeft-titleWidth, 0)
	lines := make([]string, 0, height)

	for row := 0; row < height; row++ {
		if row == titleRow {
			lines = append(lines, backgroundSpaces(s.palette, titleLeft)+title+backgroundSpaces(s.palette, titleRight))
			continue
		}

		lines = append(lines, backgroundSpaces(s.palette, width))
	}

	return strings.Join(lines, symbolLineBreak)
}
