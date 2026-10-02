package uikit

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func Max(a, b int) int {
	if a > b {
		return a
	}

	return b
}

func Min(a, b int) int {
	if a < b {
		return a
	}

	return b
}

func FormatInt(value int) string {
	return strconv.Itoa(value)
}

func BackgroundSpaces(palette Palette, width int) string {
	return lipgloss.NewStyle().
		Background(palette.Background).
		Render(strings.Repeat(" ", Max(width, 0)))
}

func BackgroundBlock(palette Palette, width, height int) string {
	lines := make([]string, 0, Max(height, 0))

	for range Max(height, 0) {
		lines = append(lines, BackgroundSpaces(palette, width))
	}

	return strings.Join(lines, SymbolLineBreak)
}

func PadBlockHeight(palette Palette, block string, height int) string {
	lines := strings.Split(block, SymbolLineBreak)
	width := lipgloss.Width(block)

	for len(lines) < height {
		lines = append(lines, BackgroundSpaces(palette, width))
	}

	return strings.Join(lines, SymbolLineBreak)
}

func NormalizeHexColor(color string) string {
	color = strings.TrimSpace(color)
	if len(color) == 6 && !strings.HasPrefix(color, "#") {
		return "#" + color
	}

	return color
}
