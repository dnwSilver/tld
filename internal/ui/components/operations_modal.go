package components

import (
	"github.com/charmbracelet/x/ansi"
	"github.com/dnwSilver/tld/internal/safetext"
	"github.com/dnwSilver/tld/internal/ui/uikit"
)

type OperationsModal struct {
	palette uikit.Palette
}

func NewOperationsModal(palette uikit.Palette) OperationsModal {
	return OperationsModal{palette: palette}
}

func (o OperationsModal) Render(width, height int, operations []uikit.SettingsOperation, selectedIndex int, projectName, preview string, confirming bool) string {
	modalWidth := uikit.Min(uikit.Max(width-8, 32), 48)
	contentWidth := uikit.Max(modalWidth-2, 1)
	modal := NewModal(o.palette, o.palette.Primary)
	rows := make([]string, 0, len(operations)+1)

	if confirming {
		if selectedIndex >= 0 && selectedIndex < len(operations) {
			rows = append(rows, modal.Line(contentWidth, modal.Text(o.palette.Primary, "Apply: "+operations[selectedIndex].Title)))
			rows = append(rows, modal.Line(contentWidth, modal.Text(o.palette.Hint, ansi.Cut(safetext.Plain(operations[selectedIndex].Change), 0, contentWidth))))
		}
		if preview == "" {
			preview = "source not found"
		}
		rows = append(rows, modal.Line(contentWidth, modal.Text(o.palette.Hint, ansi.Cut(safetext.Plain(preview), 0, contentWidth))))
		rows = append(rows, modal.CenterLine(contentWidth, modal.Text(o.palette.Hint, "[Enter] confirm  [Esc] back")))
	} else {
		for index, operation := range operations {
			color := o.palette.Hint
			marker := "  "
			if index == selectedIndex {
				color = o.palette.Primary
				marker = "> "
			}
			rows = append(rows, modal.Line(contentWidth, modal.Text(color, marker+operation.Title)))
		}

		rows = ModalWindow(rows, selectedIndex, uikit.Max(height-3, 1))
		rows = append(rows, modal.CenterLine(contentWidth, modal.Text(o.palette.Hint, "[Enter] apply  [Esc] close")))
	}

	title := uikit.SymbolSettings + " operations"
	if projectName != "" {
		title += " · " + projectName
	}

	return modal.Render(contentWidth, modal.Title(title), rows)
}
