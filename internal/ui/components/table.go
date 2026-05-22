package components

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/dnwSilver/tld/internal/ui/uikit"
)

type TableCell struct {
	Value      string
	Width      int
	Foreground lipgloss.Color
	Bold       bool
}

func RenderTableRow(palette uikit.Palette, background lipgloss.Color, rowWidth int, cells []TableCell) string {
	content := ""
	for _, cell := range cells {
		style := lipgloss.NewStyle().
			Background(background).
			Foreground(cell.Foreground)
		if cell.Width > 0 {
			style = style.Width(cell.Width)
		}
		if cell.Bold {
			style = style.Bold(true)
		}
		content += style.Render(cell.Value)
	}

	rowPalette := palette
	rowPalette.Background = background
	return content + uikit.BackgroundSpaces(rowPalette, rowWidth-lipgloss.Width(content))
}

func ColumnWidth[T any](items []T, get func(T) string, minWidth int) int {
	width := minWidth
	for _, item := range items {
		width = uikit.Max(width, lipgloss.Width(get(item)))
	}

	return width
}
