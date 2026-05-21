package screens

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
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
		BorderForeground(s.palette.Primary).
		Render(content)

	footer := lipgloss.NewStyle().
		Width(width).
		Background(s.palette.Background).
		Foreground(s.palette.Hint).
		Align(lipgloss.Right).
		Render(uikit.SymbolCopyright + " Kolosov Aleksandr")

	return lipgloss.JoinVertical(lipgloss.Left, box, footer)
}

func (s DefaultScreen) renderContent(width, height int, stackCount int) string {
	title := lipgloss.NewStyle().
		Background(s.palette.Background).
		Foreground(s.palette.Warning).
		Bold(true).
		Render(uikit.SymbolDashboard + " Team lead dashboard")
	titleRow := height / 2
	counter := lipgloss.NewStyle().
		Background(s.palette.Background).
		Foreground(s.palette.Primary).
		Render("Stacks: " + uikit.FormatInt(stackCount))
	counterRow := uikit.Min(titleRow+2, height-1)
	lines := make([]string, 0, height)

	for row := 0; row < height; row++ {
		if row == titleRow {
			lines = append(lines, s.centerLine(width, title))
			continue
		}
		if row == counterRow {
			lines = append(lines, s.centerLine(width, counter))
			continue
		}

		lines = append(lines, uikit.BackgroundSpaces(s.palette, width))
	}

	return strings.Join(lines, uikit.SymbolLineBreak)
}

func (s DefaultScreen) centerLine(width int, content string) string {
	contentWidth := lipgloss.Width(content)
	left := uikit.Max((width-contentWidth)/2, 0)
	right := uikit.Max(width-left-contentWidth, 0)

	return uikit.BackgroundSpaces(s.palette, left) + content + uikit.BackgroundSpaces(s.palette, right)
}
