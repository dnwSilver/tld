package app

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dnwSilver/tld/internal/projectsync"
	"github.com/dnwSilver/tld/internal/ui"
)

func settingsModelWithProject() model {
	m := newModel(nil)
	m.screen = ui.ScreenSettings
	m.projects = []ui.Project{{ID: 1, ProjectID: "group/repo", Name: "repo", SourceID: 2}}
	m.sources = []ui.Source{{ID: 2, Type: "gitlab", URL: "https://gitlab.example.com"}}
	m.projectCheckRows = []ui.ProjectCheckRow{{ProjectID: 1, ProjectName: "repo"}}
	m.selectedCheckProjectID = 1

	return m
}

func TestSettingsOperationsModalOpensOnU(t *testing.T) {
	m := settingsModelWithProject()

	next, _ := m.Update(key("u"))
	updated := next.(model)
	if !updated.operationsModalOpen {
		t.Fatal("operations modal did not open")
	}
	if updated.operationsModalIndex != 0 {
		t.Fatalf("modal index = %d, want 0", updated.operationsModalIndex)
	}
}

func TestSettingsOperationsModalRequiresSelectedProject(t *testing.T) {
	m := settingsModelWithProject()
	m.projectCheckRows = nil

	next, _ := m.Update(key("u"))
	updated := next.(model)
	if updated.operationsModalOpen {
		t.Fatal("operations modal opened without a selected project")
	}
}

func TestSettingsOperationsModalDoesNotOpenWhileChecksRun(t *testing.T) {
	m := settingsModelWithProject()
	m.checksStatus.Running = true

	next, _ := m.Update(key("u"))
	updated := next.(model)
	if updated.operationsModalOpen {
		t.Fatal("operations modal opened while checks are running")
	}
}

func TestSettingsOperationsModalNavigationAndClose(t *testing.T) {
	m := settingsModelWithProject()
	m.operationsModalOpen = true

	next, _ := m.Update(key("j"))
	updated := next.(model)
	if updated.operationsModalIndex != 1 {
		t.Fatalf("modal index after j = %d, want 1", updated.operationsModalIndex)
	}

	for i := 0; i < len(projectsync.ProjectOperations); i++ {
		next, _ = updated.Update(key("j"))
		updated = next.(model)
	}
	if updated.operationsModalIndex != len(projectsync.ProjectOperations)-1 {
		t.Fatalf("modal index went past the last operation: %d", updated.operationsModalIndex)
	}

	next, _ = updated.Update(key("k"))
	updated = next.(model)
	if updated.operationsModalIndex != len(projectsync.ProjectOperations)-2 {
		t.Fatalf("modal index after k = %d, want %d", updated.operationsModalIndex, len(projectsync.ProjectOperations)-2)
	}

	next, _ = updated.Update(key("u"))
	updated = next.(model)
	if updated.operationsModalOpen {
		t.Fatal("operations modal did not close on u")
	}
}

func TestSettingsOperationApplyStartsCommand(t *testing.T) {
	m := settingsModelWithProject()
	m.operationsModalOpen = true

	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := next.(model)
	if updated.operationsModalOpen {
		t.Fatal("operations modal did not close on enter")
	}
	if cmd == nil {
		t.Fatal("expected operation command")
	}
	if !updated.checksStatus.Running {
		t.Fatal("checks status is not marked as running")
	}
}

func TestSettingsOperationApplyReportsMissingSource(t *testing.T) {
	m := settingsModelWithProject()
	m.sources = nil
	m.operationsModalOpen = true

	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := next.(model)
	if cmd != nil {
		t.Fatal("operation command started without a source")
	}
	if updated.checksStatus.Error == "" {
		t.Fatal("missing source did not surface an error")
	}
}

func TestOperationAppliedMsgUpdatesStatus(t *testing.T) {
	m := settingsModelWithProject()
	m.checksStatus = ui.SettingsStatus{Message: "Applying...", Running: true}

	next, cmd := m.Update(operationAppliedMsg{title: "Share CI cache with all branches", projectID: 1, project: "repo"})
	updated := next.(model)
	if updated.checksStatus.Error != "" {
		t.Fatalf("unexpected error: %q", updated.checksStatus.Error)
	}
	if cmd == nil {
		t.Fatal("expected checks refresh after applied operation")
	}
	if !updated.checksStatus.Running {
		t.Fatal("checks refresh did not start")
	}
}
