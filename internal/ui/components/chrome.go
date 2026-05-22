package components

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/dnwSilver/tld/internal/ui/uikit"
)

func ScreenTitle(palette uikit.Palette, symbol, text string, count *int) string {
	value := symbol + " " + text
	if count != nil {
		value += " [" + uikit.FormatInt(*count) + "]"
	}

	return uikit.BoldText(palette, palette.Warning, value)
}

func ScreenFooter(palette uikit.Palette, width int, text string) string {
	return lipgloss.NewStyle().
		Width(width).
		Background(palette.Background).
		Foreground(palette.Hint).
		Align(lipgloss.Right).
		Render(uikit.SymbolCopyright + " " + text)
}
