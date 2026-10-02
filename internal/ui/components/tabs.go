package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dnwSilver/tld/internal/ui/uikit"
)

type Tab struct {
	ID     int64
	Icon   string
	Label  string
	Active bool
}

type Tabs struct {
	palette uikit.Palette
}

func NewTabs(palette uikit.Palette) Tabs {
	return Tabs{palette: palette}
}

func (t Tabs) Render(width int, tabs []Tab) string {
	content := t.RenderInline(tabs)

	return content + uikit.BackgroundSpaces(t.palette, width-lipgloss.Width(content))
}

func (t Tabs) RenderInline(tabs []Tab) string {
	parts := make([]string, 0, len(tabs))
	for _, tab := range tabs {
		color := t.palette.Hint
		if tab.Active {
			color = t.palette.Primary
		}
		label := tab.Label
		if tab.Icon != "" {
			label = tab.Icon + " " + label
		}
		parts = append(parts, lipgloss.NewStyle().
			Background(t.palette.Background).
			Foreground(color).
			Bold(tab.Active).
			Render(label))
	}

	return strings.Join(parts, uikit.BackgroundSpaces(t.palette, 3))
}
