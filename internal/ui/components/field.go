package components

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/dnwSilver/tld/internal/ui/uikit"
)

type FormField struct {
	Label      string
	Value      string
	Focused    bool
	LabelWidth int
	ValueWidth int
}

func RenderFormField(palette uikit.Palette, field FormField) string {
	prefix := "  "
	if field.Focused {
		prefix = "> "
	}

	prefixPart := uikit.Text(palette, palette.Hint, prefix)
	labelPart := lipgloss.NewStyle().
		Background(palette.Background).
		Foreground(palette.Hint).
		Width(field.LabelWidth).
		Align(lipgloss.Right).
		Render(field.Label)
	separator := uikit.Text(palette, palette.Hint, ": ")
	valuePart := uikit.Text(palette, palette.Text, field.Value)
	padding := uikit.BackgroundSpaces(palette, field.ValueWidth-lipgloss.Width(valuePart))

	return prefixPart + labelPart + separator + valuePart + padding
}
