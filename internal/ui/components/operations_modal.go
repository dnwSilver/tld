package components

import "github.com/dnwSilver/tld/internal/ui/uikit"

type OperationsModal struct {
	palette uikit.Palette
}

func NewOperationsModal(palette uikit.Palette) OperationsModal {
	return OperationsModal{palette: palette}
}

func (o OperationsModal) Render(width int, operations []uikit.SettingsOperation, selectedIndex int, projectName string) string {
	modalWidth := uikit.Min(uikit.Max(width-8, 32), 48)
	contentWidth := uikit.Max(modalWidth-2, 1)
	modal := NewModal(o.palette, o.palette.Primary)
	rows := make([]string, 0, len(operations)+1)

	for index, operation := range operations {
		color := o.palette.Hint
		marker := "  "
		if index == selectedIndex {
			color = o.palette.Primary
			marker = "> "
		}
		rows = append(rows, modal.Line(contentWidth, modal.Text(color, marker+operation.Title)))
	}

	rows = append(rows, modal.CenterLine(contentWidth, modal.Text(o.palette.Hint, "[Enter] apply  [Esc] close")))

	title := uikit.SymbolSettings + " operations"
	if projectName != "" {
		title += " · " + projectName
	}

	return modal.Render(contentWidth, modal.Title(title), rows)
}
