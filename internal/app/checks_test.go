package app

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dnwSilver/tld/internal/ui"
)

func TestSettingsColumnNavigationUsesShiftArrows(t *testing.T) {
	m := newModel(nil)
	m.screen = ui.ScreenSettings
	m.checkColumns = []ui.ProjectCheck{{ID: "master"}, {ID: "dev"}}

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyShiftRight})
	updated := next.(model)
	if updated.settingsTableState.SelectedColumn != 1 {
		t.Fatalf("selected column after shift+right = %d, want 1", updated.settingsTableState.SelectedColumn)
	}

	next, _ = updated.Update(tea.KeyMsg{Type: tea.KeyShiftRight})
	updated = next.(model)
	next, _ = updated.Update(tea.KeyMsg{Type: tea.KeyShiftRight})
	updated = next.(model)
	if updated.settingsTableState.SelectedColumn != 2 {
		t.Fatalf("selected column moved past the last column: %d", updated.settingsTableState.SelectedColumn)
	}

	next, _ = updated.Update(tea.KeyMsg{Type: tea.KeyShiftLeft})
	updated = next.(model)
	if updated.settingsTableState.SelectedColumn != 1 {
		t.Fatalf("selected column after shift+left = %d, want 1", updated.settingsTableState.SelectedColumn)
	}
}

func TestSettingsSortsProjectsAndTogglesDirection(t *testing.T) {
	m := newModel(nil)
	m.screen = ui.ScreenSettings
	m.projectCheckRows = []ui.ProjectCheckRow{
		{ProjectID: 1, ProjectName: "Zulu"},
		{ProjectID: 2, ProjectName: "alpha"},
		{ProjectID: 3, ProjectName: "Beta"},
	}
	m.selectedCheckProjectID = 1

	next, _ := m.Update(key("s"))
	updated := next.(model)
	assertProjectCheckOrder(t, updated.projectCheckRows, 2, 3, 1)
	if !updated.settingsTableState.SortActive || updated.settingsTableState.SortDescending {
		t.Fatalf("first sort state = %#v, want active ascending", updated.settingsTableState)
	}
	if updated.selectedCheckProjectID != 1 {
		t.Fatalf("selected project changed after sort: %d", updated.selectedCheckProjectID)
	}

	next, _ = updated.Update(key("s"))
	updated = next.(model)
	assertProjectCheckOrder(t, updated.projectCheckRows, 1, 3, 2)
	if !updated.settingsTableState.SortDescending {
		t.Fatalf("second sort state = %#v, want descending", updated.settingsTableState)
	}
}

func TestSettingsSortsCheckColumnWithProblemsFirst(t *testing.T) {
	m := newModel(nil)
	m.screen = ui.ScreenSettings
	m.checkColumns = []ui.ProjectCheck{{ID: "master"}}
	m.settingsTableState.SelectedColumn = 1
	m.projectCheckRows = []ui.ProjectCheckRow{
		{ProjectID: 1, ProjectName: "pass", Results: map[string]ui.CheckState{"master": ui.CheckStatePass}},
		{ProjectID: 2, ProjectName: "unknown", Results: map[string]ui.CheckState{"master": ui.CheckStateUnknown}},
		{ProjectID: 3, ProjectName: "fail", Results: map[string]ui.CheckState{"master": ui.CheckStateFail}},
		{ProjectID: 4, ProjectName: "n/a", Results: map[string]ui.CheckState{"master": ui.CheckStateNotApplicable}},
		{ProjectID: 5, ProjectName: "warning", Results: map[string]ui.CheckState{"master": ui.CheckStateWarning}},
	}

	next, _ := m.Update(key("s"))
	updated := next.(model)
	assertProjectCheckOrder(t, updated.projectCheckRows, 3, 5, 2, 1, 4)
}

func TestSettingsReappliesSortAfterChecksReload(t *testing.T) {
	m := newModel(nil)
	m.screen = ui.ScreenSettings
	m.settingsTableState = ui.SettingsTableState{
		SortActive:     true,
		SortDescending: true,
	}

	next, _ := m.Update(projectChecksLoadedMsg{rows: []ui.ProjectCheckRow{
		{ProjectID: 1, ProjectName: "alpha"},
		{ProjectID: 2, ProjectName: "Zulu"},
	}})
	updated := next.(model)
	assertProjectCheckOrder(t, updated.projectCheckRows, 2, 1)
}

func assertProjectCheckOrder(t *testing.T, rows []ui.ProjectCheckRow, want ...int64) {
	t.Helper()
	if len(rows) != len(want) {
		t.Fatalf("row count = %d, want %d", len(rows), len(want))
	}
	for index, projectID := range want {
		if rows[index].ProjectID != projectID {
			t.Fatalf("row %d project = %d, want %d", index, rows[index].ProjectID, projectID)
		}
	}
}
