package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dnwSilver/tld/internal/safetext"
	"github.com/dnwSilver/tld/internal/ui/uikit"
)

func ScreenTitle(palette uikit.Palette, symbol, text string, count *int, searchQuery string) string {
	value := uikit.BoldText(palette, palette.Primary, symbol+" "+text)
	if count != nil {
		value += uikit.BoldText(palette, palette.Primary, " [") +
			uikit.BoldText(palette, palette.Text, uikit.FormatInt(*count)) +
			uikit.BoldText(palette, palette.Primary, "]")
	}
	if query := safetext.Plain(strings.TrimSpace(searchQuery)); query != "" {
		value += uikit.BoldText(palette, palette.Primary, "[") +
			uikit.BoldText(palette, palette.Info, query) +
			uikit.BoldText(palette, palette.Primary, "]")
	}

	return value
}

func ScreenFooter(palette uikit.Palette, width int, text string) string {
	return lipgloss.NewStyle().
		Width(width).
		Background(palette.Background).
		Foreground(palette.Hint).
		Align(lipgloss.Right).
		Render(uikit.SymbolCopyright + " " + text)
}
