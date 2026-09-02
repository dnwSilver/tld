package app

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dnwSilver/tld/internal/projectsync"
	"github.com/dnwSilver/tld/internal/ui"
)

type operationAppliedMsg struct {
	title     string
	projectID int64
	project   string
	err       error
}

func settingsOperationItems() []ui.SettingsOperation {
	items := make([]ui.SettingsOperation, 0, len(projectsync.ProjectOperations))
	for _, operation := range projectsync.ProjectOperations {
		items = append(items, ui.SettingsOperation{ID: operation.ID, Title: operation.Title})
	}

	return items
}

func (m model) openOperationsModal() model {
	if _, ok := findByID(m.projectCheckRows, m.selectedCheckProjectID, projectCheckRowID); !ok {
		return m
	}
	m.operationsModalOpen = true
	m.operationsModalIndex = 0

	return m
}

func (m model) updateOperationsModal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch {
	case isCancelKey(msg), ui.KeyOperations.Matches(key):
		m.operationsModalOpen = false
	case ui.KeyPrev.Matches(key):
		if m.operationsModalIndex > 0 {
			m.operationsModalIndex--
		}
	case ui.KeyNext.Matches(key):
		if m.operationsModalIndex < len(projectsync.ProjectOperations)-1 {
			m.operationsModalIndex++
		}
	case isEnterKey(msg):
		return m.applySelectedOperation()
	}

	return m, nil
}

func (m model) applySelectedOperation() (tea.Model, tea.Cmd) {
	m.operationsModalOpen = false
	if m.operationsModalIndex < 0 || m.operationsModalIndex >= len(projectsync.ProjectOperations) {
		return m, nil
	}
	operation := projectsync.ProjectOperations[m.operationsModalIndex]

	project, ok := findByID(m.projects, m.selectedCheckProjectID, projectID)
	if !ok {
		return m, nil
	}
	source, ok := findByID(m.sources, project.SourceID, sourceID)
	if !ok {
		m.checksStatus = ui.SettingsStatus{Error: "source not found for " + project.Name}
		return m, nil
	}

	m.checksStatus = ui.SettingsStatus{
		Message: "Applying " + operation.Title + " to " + project.Name + "...",
		Running: true,
	}

	return m, applyProjectOperation(source, project, operation)
}

func applyProjectOperation(source ui.Source, project ui.Project, operation projectsync.OperationDefinition) tea.Cmd {
	return func() tea.Msg {
		client, err := projectsync.NewSourceClient(source.Type, nil)
		if err != nil {
			return operationAppliedMsg{title: operation.Title, projectID: project.ID, project: project.Name, err: err}
		}

		service := projectsync.OperationService{SourceClient: client}
		err = service.Run(context.Background(), projectsync.Source{
			ID:       source.ID,
			Type:     source.Type,
			URL:      source.URL,
			PATToken: source.PATToken,
		}, projectsync.Project{
			ID:         project.ID,
			ProviderID: project.ProjectID,
			Name:       project.Name,
		}, operation.ID)

		return operationAppliedMsg{title: operation.Title, projectID: project.ID, project: project.Name, err: err}
	}
}
