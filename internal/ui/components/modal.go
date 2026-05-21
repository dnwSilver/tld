package components

import (
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dnwSilver/tld/internal/ui/uikit"
)

type Modal struct {
	palette     uikit.Palette
	borderColor lipgloss.Color
}

func NewModal(palette uikit.Palette, borderColor lipgloss.Color) Modal {
	return Modal{
		palette:     palette,
		borderColor: borderColor,
	}
}

func (m Modal) Render(contentWidth int, title string, rows []string) string {
	lines := make([]string, 0, len(rows)+2)
	lines = append(lines, m.topBorder(contentWidth, title))
	for _, row := range rows {
		lines = append(lines, m.borderPart("│")+m.Line(contentWidth, row)+m.borderPart("│"))
	}
	lines = append(lines, m.borderPart("└")+m.borderFill(contentWidth)+m.borderPart("┘"))

	return strings.Join(lines, uikit.SymbolLineBreak)
}

func (m Modal) Line(width int, content string) string {
	contentWidth := lipgloss.Width(content)
	return content + m.Spaces(width-contentWidth)
}

func (m Modal) CenterLine(width int, content string) string {
	contentWidth := lipgloss.Width(content)
	left := uikit.Max((width-contentWidth)/2, 0)
	right := uikit.Max(width-left-contentWidth, 0)

	return m.Spaces(left) + content + m.Spaces(right)
}

func (m Modal) Spaces(width int) string {
	return lipgloss.NewStyle().
		Background(m.palette.Background).
		Render(strings.Repeat(" ", uikit.Max(width, 0)))
}

func (m Modal) Title(title string) string {
	return lipgloss.NewStyle().
		Background(m.palette.Background).
		Foreground(m.palette.Warning).
		Bold(true).
		Render(title)
}

func (m Modal) Text(color lipgloss.Color, text string) string {
	return lipgloss.NewStyle().
		Background(m.palette.Background).
		Foreground(color).
		Render(text)
}

func (m Modal) topBorder(width int, title string) string {
	titleWidth := lipgloss.Width(title)
	available := uikit.Max(width-titleWidth-2, 0)
	left := int(math.Floor(float64(available) / 2))
	right := uikit.Max(available-left, 0)

	return m.borderPart("┌") +
		m.borderFill(left) +
		m.borderPart(" ") +
		title +
		m.borderPart(" ") +
		m.borderFill(right) +
		m.borderPart("┐")
}

func (m Modal) borderFill(width int) string {
	return m.borderPart(strings.Repeat("─", uikit.Max(width, 0)))
}

func (m Modal) borderPart(value string) string {
	return lipgloss.NewStyle().
		Background(m.palette.Background).
		Foreground(m.borderColor).
		Render(value)
}
