package components

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/dnwSilver/tld/internal/ui/uikit"
)

type Action struct {
	Hint  string
	Color lipgloss.Color
}

func RenderActions(palette uikit.Palette, actions []Action) string {
	separator := uikit.BackgroundSpaces(palette, 3)
	out := ""
	for index, action := range actions {
		if index > 0 {
			out += separator
		}
		out += uikit.Text(palette, action.Color, action.Hint)
	}

	return out
}
