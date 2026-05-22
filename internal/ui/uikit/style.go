package uikit

import "github.com/charmbracelet/lipgloss"

func Text(palette Palette, color lipgloss.Color, value string) string {
	return lipgloss.NewStyle().
		Background(palette.Background).
		Foreground(color).
		Render(value)
}

func BoldText(palette Palette, color lipgloss.Color, value string) string {
	return lipgloss.NewStyle().
		Background(palette.Background).
		Foreground(color).
		Bold(true).
		Render(value)
}

func CenterLine(palette Palette, width int, content string) string {
	contentWidth := lipgloss.Width(content)
	left := Max((width-contentWidth)/2, 0)
	right := Max(width-left-contentWidth, 0)

	return BackgroundSpaces(palette, left) + content + BackgroundSpaces(palette, right)
}
