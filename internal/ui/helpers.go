package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func max(a, b int) int {
	if a > b {
		return a
	}

	return b
}

func backgroundSpaces(palette Palette, width int) string {
	return lipgloss.NewStyle().
		Background(palette.Background).
		Render(strings.Repeat(" ", max(width, 0)))
}

func backgroundBlock(palette Palette, width, height int) string {
	lines := make([]string, 0, max(height, 0))

	for range max(height, 0) {
		lines = append(lines, backgroundSpaces(palette, width))
	}

	return strings.Join(lines, symbolLineBreak)
}

func padBlockHeight(palette Palette, block string, height int) string {
	lines := strings.Split(block, symbolLineBreak)
	width := lipgloss.Width(block)

	for len(lines) < height {
		lines = append(lines, backgroundSpaces(palette, width))
	}

	return strings.Join(lines, symbolLineBreak)
}
