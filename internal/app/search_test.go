package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dnwSilver/tld/internal/ui"
)

func TestSearchSelectsMatchingProjectByID(t *testing.T) {
	m := newModel(nil)
	m.screen = ui.ScreenProjects
	m.width, m.height = 100, 24
	m.projects = []ui.Project{{ID: 10, Name: "Alpha"}, {ID: 20, Name: "Beta Project"}}
	opened, _ := m.Update(key("/"))
	searched, _ := opened.(model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("beta")})
	if got := searched.(model).searchMatches(); len(got) != 1 || got[0].id != 20 {
		t.Fatalf("matches = %#v", got)
	}
	if view := searched.(model).View(); !strings.Contains(view, "Beta Project") || strings.Contains(view, "Alpha") || !strings.Contains(view, "[1 matches]") {
		t.Fatalf("table was not filtered while typing: %q", view)
	}
	selected, _ := searched.(model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	if selected.(model).selectedProjectID != 20 || selected.(model).searchOpen || selected.(model).searchQuery != "beta" {
		t.Fatalf("selection = %d, open = %v, query = %q", selected.(model).selectedProjectID, selected.(model).searchOpen, selected.(model).searchQuery)
	}
	if view := selected.(model).View(); strings.Contains(view, "Alpha") {
		t.Fatalf("Enter removed active filter: %q", view)
	}
	if view := selected.(model).View(); !strings.Contains(view, "Projects [1][beta]") || strings.Contains(view, "matches]") || strings.Contains(view, "[Enter] done") {
		t.Fatalf("Enter did not move the filter into the panel title: %q", view)
	}
	reopened, _ := selected.(model).Update(key("/"))
	if view := reopened.(model).View(); !strings.Contains(view, "/ beta▏") || !strings.Contains(view, "[1 matches]") {
		t.Fatalf("reopening search did not preserve the query: %q", view)
	}
	canceled, _ := reopened.(model).Update(tea.KeyMsg{Type: tea.KeyEsc})
	if view := canceled.(model).View(); !strings.Contains(view, "Projects [2]") || !strings.Contains(view, "Alpha") || strings.Contains(view, "[beta]") || strings.Contains(view, "matches]") {
		t.Fatalf("Esc did not clear and hide search: %q", view)
	}
}

func TestSearchCancelPreservesSelection(t *testing.T) {
	m := newModel(nil)
	m.screen = ui.ScreenSources
	m.sources = []ui.Source{{ID: 1, Name: "first"}, {ID: 2, Name: "second"}}
	m.selectedSourceID = 1
	opened, _ := m.Update(key("/"))
	canceled, _ := opened.(model).Update(tea.KeyMsg{Type: tea.KeyEsc})
	if canceled.(model).selectedSourceID != 1 || canceled.(model).searchOpen {
		t.Fatal("cancel changed source selection")
	}
}

func TestSearchNavigationOnlyVisitsVisibleRows(t *testing.T) {
	m := newModel(nil)
	m.screen = ui.ScreenSources
	m.width, m.height = 80, 36
	m.sources = []ui.Source{
		{ID: 1, Name: "alpha-prod"}, {ID: 2, Name: "hidden"},
		{ID: 3, Name: "beta-prod"}, {ID: 4, Name: "gamma"},
	}
	m.selectedSourceID = 1
	next, _ := m.Update(key("/"))
	next, _ = next.(model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("PROD")})
	m = next.(model)
	if m.selectedSourceID != 1 || strings.Contains(m.View(), "hidden") || strings.Contains(m.View(), "gamma") {
		t.Fatalf("filter did not preserve matching selection: %d", m.selectedSourceID)
	}
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = next.(model)
	if m.selectedSourceID != 3 {
		t.Fatalf("Down selected hidden row: %d", m.selectedSourceID)
	}
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(model)
	next, _ = m.Update(key("k"))
	m = next.(model)
	if m.selectedSourceID != 1 {
		t.Fatalf("k after Enter selected hidden row: %d", m.selectedSourceID)
	}
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = next.(model)
	if m.searchQuery != "" || m.searchOpen || !strings.Contains(m.View(), "hidden") {
		t.Fatalf("Esc did not restore the full table: query=%q open=%t view=%q", m.searchQuery, m.searchOpen, m.View())
	}
}

func TestSearchMatchesOnlyObjectNamesAndHasSafeEmptyResult(t *testing.T) {
	m := newModel(nil)
	m.screen = ui.ScreenSources
	m.width, m.height = 80, 36
	m.sources = []ui.Source{{ID: 1, Name: "alpha", URL: "https://beta.example"}}
	m.selectedSourceID = 1
	next, _ := m.Update(key("/"))
	next, _ = next.(model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("beta")})
	m = next.(model)
	if len(m.searchMatches()) != 0 || m.selectedSourceID != 0 || !strings.Contains(m.View(), "[0 matches]") {
		t.Fatalf("unexpected matches or selected hidden row: %#v, %d", m.searchMatches(), m.selectedSourceID)
	}
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if next.(model).selectedSourceID != 1 {
		t.Fatal("clearing empty filter did not restore a visible selection")
	}
}

func TestSearchClearsWhenSwitchingScreens(t *testing.T) {
	m := newModel(nil)
	m.screen = ui.ScreenSources
	m.searchQuery = "alpha"
	m.sources = []ui.Source{{ID: 1, Name: "alpha"}}
	next, _ := m.switchToScreen(ui.ScreenStacks)
	if next.(model).searchQuery != "" || next.(model).searchOpen {
		t.Fatal("filter leaked to another screen")
	}
}

func TestSearchKeepsHiddenRowsUnselectedAfterReload(t *testing.T) {
	m := newModel(nil)
	m.screen = ui.ScreenSources
	m.searchQuery = "beta"
	m.selectedSourceID = 0
	next, _ := m.Update(sourcesLoadedMsg{sources: []ui.Source{{ID: 1, Name: "alpha"}}})
	m = next.(model)
	if m.selectedSourceID != 0 || len(m.searchMatches()) != 0 {
		t.Fatalf("reload selected a hidden row: %d", m.selectedSourceID)
	}
	next, _ = m.Update(sourcesLoadedMsg{sources: []ui.Source{{ID: 1, Name: "alpha"}, {ID: 2, Name: "beta"}}})
	if next.(model).selectedSourceID != 2 {
		t.Fatalf("reload did not select the newly visible row: %d", next.(model).selectedSourceID)
	}
}

func TestSearchFiltersProjectRowsOnSummaryScreens(t *testing.T) {
	tests := []struct {
		name   string
		screen ui.Screen
		setup  func(*model)
	}{
		{"dashboard", ui.ScreenDefault, func(m *model) {
			m.dashboard.focus = ui.DashboardPaneAttention
			m.dashboard.attentionRows = []ui.DashboardAttentionRow{{ProjectID: 1, ProjectName: "AlphaUnique"}, {ProjectID: 2, ProjectName: "BetaUnique"}}
		}},
		{"view", ui.ScreenView, func(m *model) {
			m.dependencyView.Rows = []ui.DependencyViewRow{{ProjectID: 1, ProjectName: "AlphaUnique"}, {ProjectID: 2, ProjectName: "BetaUnique"}}
		}},
		{"settings", ui.ScreenSettings, func(m *model) {
			m.projectCheckRows = []ui.ProjectCheckRow{{ProjectID: 1, ProjectName: "AlphaUnique"}, {ProjectID: 2, ProjectName: "BetaUnique"}}
		}},
		{"releases", ui.ScreenReleases, func(m *model) {
			m.releaseRows = []ui.ReleaseRow{{ProjectID: 1, ProjectName: "AlphaUnique"}, {ProjectID: 2, ProjectName: "BetaUnique"}}
		}},
		{"vulnerabilities", ui.ScreenVulnerabilities, func(m *model) {
			m.vulnRows = []ui.VulnProjectRow{{ProjectID: 1, ProjectName: "AlphaUnique"}, {ProjectID: 2, ProjectName: "BetaUnique"}}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newModel(nil)
			m.screen = tt.screen
			m.width, m.height = 120, 40
			m.searchQuery = "beta"
			tt.setup(&m)
			view := m.View()
			if !strings.Contains(view, "BetaUnique") || strings.Contains(view, "AlphaUnique") || !strings.Contains(view, "[1][beta]") || strings.Contains(view, "matches]") {
				t.Fatalf("project rows were not filtered: %q", view)
			}
		})
	}
}

func TestListPageNavigationKeepsStableID(t *testing.T) {
	m := newModel(nil)
	m.screen = ui.ScreenSources
	m.height = 15
	for index := int64(1); index <= 10; index++ {
		m.sources = append(m.sources, ui.Source{ID: index, Name: "source"})
	}
	m.selectedSourceID = 1
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	if got := next.(model).selectedSourceID; got != 4 {
		t.Fatalf("PageDown selected ID %d, want 4", got)
	}
	last, _ := next.(model).Update(tea.KeyMsg{Type: tea.KeyEnd})
	if got := last.(model).selectedSourceID; got != 10 {
		t.Fatalf("End selected ID %d, want 10", got)
	}
}
