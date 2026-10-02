package components

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/dnwSilver/tld/internal/ui/uikit"
)

type Hint struct {
	palette uikit.Palette
	color   lipgloss.Color
	icon    string
	text    string
	width   int
}

func NewHint(palette uikit.Palette, color lipgloss.Color, icon, text string, width int) Hint {
	return Hint{
		palette: palette,
		color:   color,
		icon:    icon,
		text:    text,
		width:   width,
	}
}

func (h Hint) Render() string {
	line := lipgloss.NewStyle().
		Background(h.palette.Background).
		Foreground(h.color).
		Render(h.icon) + uikit.BackgroundSpaces(h.palette, 1) + lipgloss.NewStyle().
		Background(h.palette.Background).
		Foreground(h.palette.Hint).
		Render(h.text)

	return line + uikit.BackgroundSpaces(h.palette, h.width-lipgloss.Width(line))
}
