package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/dnwSilver/tld/internal/safetext"
)

// renderSearchLine is part of the screen layout, so it reduces the available
// body height instead of covering rows as a modal would.
func (c Creator) renderSearchLine(width int, state OverlayRenderState) string {
	if width <= 0 || !state.SearchEditing {
		return ""
	}
	cursor := "▏"
	hint := "[Enter] done · [Esc] clear"
	summary := fmt.Sprintf("  [%d matches]  %s", state.SearchCount, hint)
	if width-lipgloss.Width("/ "+cursor+summary) < 8 {
		summary = ""
	}
	queryWidth := width - lipgloss.Width("/ "+cursor+summary)
	query := safetext.Plain(state.SearchQuery)
	if lipgloss.Width(query) > queryWidth {
		if queryWidth <= 1 {
			query = ""
		} else {
			query = "…" + ansi.Cut(query, lipgloss.Width(query)-queryWidth+1, lipgloss.Width(query))
		}
	}
	summary = lipgloss.NewStyle().Foreground(c.palette.Hint).Background(c.palette.Background).Render(summary)
	label := ansi.Truncate("/ "+query+cursor+summary, width, "…")
	return lipgloss.NewStyle().
		Foreground(c.palette.Info).
		Background(c.palette.Background).
		Width(width).
		Render(label)
}
