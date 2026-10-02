package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/dnwSilver/tld/internal/ui/uikit"
)

type TabbedPanel struct {
	palette     uikit.Palette
	borderColor lipgloss.Color
}

func NewTabbedPanel(palette uikit.Palette, borderColor lipgloss.Color) TabbedPanel {
	return TabbedPanel{
		palette:     palette,
		borderColor: borderColor,
	}
}

func (p TabbedPanel) Render(contentWidth int, contentHeight int, title string, tabs []Tab, content string) string {
	contentLines := strings.Split(content, uikit.SymbolLineBreak)
	for len(contentLines) < contentHeight {
		contentLines = append(contentLines, p.spaces(contentWidth))
	}
	if len(contentLines) > contentHeight {
		contentLines = contentLines[:contentHeight]
	}

	lines := make([]string, 0, len(contentLines)+2)
	lines = append(lines, p.topBorder(contentWidth, title, tabs))
	for _, line := range contentLines {
		lines = append(lines, p.borderPart("│")+p.line(contentWidth, line)+p.borderPart("│"))
	}
	lines = append(lines, p.borderPart("└")+p.borderFill(contentWidth)+p.borderPart("┘"))

	return strings.Join(lines, uikit.SymbolLineBreak)
}

func (p TabbedPanel) topBorder(width int, title string, tabs []Tab) string {
	title = ansi.Truncate(title, uikit.Max(width-2, 0), "…")
	titlePart := p.spaces(1) + title + p.spaces(1)
	tabsContent := NewTabs(p.palette).RenderInline(tabs)
	maxTabsWidth := uikit.Max(width-lipgloss.Width(titlePart)-2, 0)
	if lipgloss.Width(tabsContent) > maxTabsWidth {
		tabsContent = ansi.Cut(tabsContent, 0, maxTabsWidth)
	}
	tabsPart := ""
	if tabsContent != "" {
		tabsPart = p.spaces(1) + tabsContent + p.spaces(1)
	}
	fillWidth := width - lipgloss.Width(titlePart) - lipgloss.Width(tabsPart)

	return p.borderPart("┌") + titlePart + p.borderFill(fillWidth) + tabsPart + p.borderPart("┐")
}

func (p TabbedPanel) line(width int, content string) string {
	return content + p.spaces(width-lipgloss.Width(content))
}

func (p TabbedPanel) spaces(width int) string {
	return lipgloss.NewStyle().
		Background(p.palette.Background).
		Render(strings.Repeat(" ", uikit.Max(width, 0)))
}

func (p TabbedPanel) borderFill(width int) string {
	return p.borderPart(strings.Repeat("─", uikit.Max(width, 0)))
}

func (p TabbedPanel) borderPart(value string) string {
	return lipgloss.NewStyle().
		Background(p.palette.Background).
		Foreground(p.borderColor).
		Render(value)
}
