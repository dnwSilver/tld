package screens

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dnwSilver/tld/internal/ui/components"
	"github.com/dnwSilver/tld/internal/ui/uikit"
)

type DefaultScreen struct {
	palette uikit.Palette
}

func NewDefaultScreen(palette uikit.Palette) DefaultScreen {
	return DefaultScreen{
		palette: palette,
	}
}

func (s DefaultScreen) Render(width, height int, stackCount int) string {
	boxWidth := uikit.Max(width, 2)
	boxHeight := uikit.Max(height-1, 3)
	contentWidth := uikit.Max(boxWidth-2, 1)
	contentHeight := uikit.Max(boxHeight-2, 1)

	content := s.renderContent(contentWidth, contentHeight, stackCount)

	box := lipgloss.NewStyle().
		Background(s.palette.Background).
		Width(contentWidth).
		Height(contentHeight).
		Border(lipgloss.NormalBorder()).
		BorderForeground(s.palette.Hint).
		Render(content)

	footer := components.ScreenFooter(s.palette, width, "Kolosov Aleksandr")

	return lipgloss.JoinVertical(lipgloss.Left, box, footer)
}

func (s DefaultScreen) renderContent(width, height int, stackCount int) string {
	title := components.ScreenTitle(s.palette, uikit.SymbolDashboard, "Team lead dashboard", nil)
	counter := uikit.Text(s.palette, s.palette.Primary, "Stacks: "+uikit.FormatInt(stackCount))
	titleRow := height / 2
	counterRow := uikit.Min(titleRow+2, height-1)
	lines := make([]string, 0, height)

	for row := 0; row < height; row++ {
		switch row {
		case titleRow:
			lines = append(lines, uikit.CenterLine(s.palette, width, title))
		case counterRow:
			lines = append(lines, uikit.CenterLine(s.palette, width, counter))
		default:
			lines = append(lines, uikit.BackgroundSpaces(s.palette, width))
		}
	}

	return strings.Join(lines, uikit.SymbolLineBreak)
}
