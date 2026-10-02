package components

import "github.com/dnwSilver/tld/internal/ui/uikit"

type NavModal struct {
	palette uikit.Palette
}

func NewNavModal(palette uikit.Palette) NavModal {
	return NavModal{palette: palette}
}

func (n NavModal) Render(width, height int, selectedIndex int, currentScreen uikit.Screen) string {
	modalWidth := uikit.Min(uikit.Max(width-8, 28), 40)
	contentWidth := uikit.Max(modalWidth-2, 1)
	modal := NewModal(n.palette, n.palette.Primary)
	rows := make([]string, 0, len(uikit.NavSections)+1)

	for index, section := range uikit.NavSections {
		color := n.palette.Hint
		marker := "  "
		if index == selectedIndex {
			color = n.palette.Primary
			marker = "> "
		} else if section.Screen == currentScreen {
			marker = "▸ "
		}

		line := marker + section.Binding.Symbol + " " + section.Binding.Hint
		rows = append(rows, modal.CenterLine(contentWidth, modal.Text(color, line)))
	}

	rows = ModalWindow(rows, selectedIndex, uikit.Max(height-3, 1))
	rows = append(rows, modal.CenterLine(contentWidth, modal.Text(n.palette.Hint, "[Enter] go  [Esc] close")))
	return modal.Render(contentWidth, modal.Title(uikit.SymbolToggleHead+" sections"), rows)
}
