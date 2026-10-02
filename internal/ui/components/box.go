package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dnwSilver/tld/internal/ui/uikit"
)

type Box struct {
	modal Modal
}

func NewBox(palette uikit.Palette, borderColor lipgloss.Color) Box {
	return Box{modal: NewModal(palette, borderColor)}
}

func (b Box) Render(contentWidth, contentHeight int, title, content string) string {
	contentLines := strings.Split(content, uikit.SymbolLineBreak)
	for len(contentLines) < contentHeight {
		contentLines = append(contentLines, b.modal.Spaces(contentWidth))
	}
	if len(contentLines) > contentHeight {
		contentLines = contentLines[:contentHeight]
	}

	return b.modal.Render(contentWidth, title, contentLines)
}
