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

func ProjectNameColor(palette Palette, freezing, endOfLife bool, defaultColor lipgloss.Color) lipgloss.Color {
	switch {
	case endOfLife:
		return palette.Hint
	case freezing:
		return palette.Info
	default:
		return defaultColor
	}
}

func RenderProjectName(palette Palette, name string, freezing, endOfLife bool, background, defaultColor lipgloss.Color, width int) string {
	style := lipgloss.NewStyle().
		Background(background).
		Foreground(ProjectNameColor(palette, freezing, endOfLife, defaultColor))
	if endOfLife {
		style = style.Strikethrough(true)
	}
	if width > 0 {
		style = style.Width(width)
	}

	return style.Render(name)
}

func RenderProjectWithIcon(
	palette Palette,
	icon string,
	iconColor lipgloss.Color,
	name string,
	freezing, endOfLife bool,
	background, defaultNameColor lipgloss.Color,
	iconWidth int,
) string {
	iconPart := lipgloss.NewStyle().
		Background(background).
		Foreground(iconColor).
		Width(iconWidth).
		Render(icon)

	return iconPart + RenderProjectName(palette, name, freezing, endOfLife, background, defaultNameColor, 0)
}
