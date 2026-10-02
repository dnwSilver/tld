package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/dnwSilver/tld/internal/ui/uikit"
)

func TestSearchLineAppearsAboveFilteredTable(t *testing.T) {
	creator := NewCreator()
	state := RenderState{
		Width: 80, Height: 24, Screen: uikit.ScreenStacks,
		Stacks:  StacksRenderState{Items: []uikit.Stack{{ID: 1, Name: "Beta"}}},
		Overlay: OverlayRenderState{SearchEditing: true, SearchQuery: "beta\x1b[31m", SearchCount: 1},
	}
	view := creator.Render(state)
	if !strings.Contains(view, "Beta") || !strings.Contains(view, "[1 matches]") || strings.Contains(view, "beta\x1b[31m") {
		t.Fatalf("inline search line or results missing: %q", view)
	}
	if strings.Index(view, "/ beta") > strings.Index(view, "Stacks") {
		t.Fatal("search line was not above the table")
	}
	if lipgloss.Height(view) > 24 || lipgloss.Width(view) > 80 {
		t.Fatalf("search overflowed frame: %dx%d", lipgloss.Width(view), lipgloss.Height(view))
	}
	state.Overlay.SearchEditing = false
	state.Overlay.SearchQuery = "beta"
	closed := creator.Render(state)
	if !strings.Contains(closed, "Stacks [1][beta]") || strings.Contains(closed, "matches]") || strings.Contains(closed, "/ beta") {
		t.Fatalf("closed search was not confined to the title: %q", closed)
	}
	if strings.Count(closed, "│") != strings.Count(view, "│")+2 {
		t.Fatal("closing search did not return its row to the table")
	}
}

func TestSearchLineKeepsLatestInputVisibleInNarrowTerminal(t *testing.T) {
	line := NewCreator().renderSearchLine(20, OverlayRenderState{SearchEditing: true, SearchQuery: strings.Repeat("a", 40) + "last"})
	if lipgloss.Width(line) != 20 || !strings.Contains(line, "last") || !strings.Contains(line, "…") {
		t.Fatalf("latest typed text is hidden: %q", line)
	}
	for _, screen := range []uikit.Screen{uikit.ScreenProjects, uikit.ScreenView} {
		for _, width := range []int{48, 100} {
			view := NewCreator().Render(RenderState{
				Width: width, Height: 24, Screen: screen,
				Overlay: OverlayRenderState{SearchQuery: strings.Repeat("界", 120)},
			})
			if lipgloss.Width(view) > width || lipgloss.Height(view) > 24 || !strings.Contains(view, "…") {
				t.Fatalf("long filter overflowed screen %v: %dx%d", screen, lipgloss.Width(view), lipgloss.Height(view))
			}
		}
	}
}
