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
	result    projectsync.OperationResult
	err       error
}

func settingsOperationItems() []ui.SettingsOperation {
	items := make([]ui.SettingsOperation, 0, len(projectsync.ProjectOperations))
	for _, operation := range projectsync.ProjectOperations {
		items = append(items, ui.SettingsOperation{ID: operation.ID, Title: operation.Title, Change: operation.Change})
	}

	return items
}

func (m model) openOperationsModal() model {
	if _, ok := findByID(m.projectCheckRows, m.selectedCheckProjectID, projectCheckRowID); !ok {
		return m
	}
	m.operationsModalOpen = true
	m.operationsModalIndex = 0
	m.operationsConfirm = false

	return m
}

func (m model) updateOperationsModal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	action, ok := ui.ModalActionForKey(ui.ModalOperations, key)
	if !ok {
		return m, nil
	}
	switch action {
	case ui.ModalActionCancel:
		if m.operationsConfirm {
			m.operationsConfirm = false
		} else {
			m.operationsModalOpen = false
		}
	case ui.ModalActionPrev:
		if m.operationsConfirm {
			break
		}
		if m.operationsModalIndex > 0 {
			m.operationsModalIndex--
		}
	case ui.ModalActionNext:
		if m.operationsConfirm {
			break
		}
		if m.operationsModalIndex < len(projectsync.ProjectOperations)-1 {
			m.operationsModalIndex++
		}
	case ui.ModalActionConfirm:
		if m.operationsConfirm {
			return m.applySelectedOperation()
		}
		m.operationsConfirm = true
	}

	return m, nil
}

func (m model) applySelectedOperation() (tea.Model, tea.Cmd) {
	m.operationsModalOpen = false
	m.operationsConfirm = false
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
	source = m.sourceWithCredential(source)

	m.checksStatus = ui.SettingsStatus{
		Message: "Applying " + operation.Title + " to " + project.Name + "...",
		Running: true,
	}

	return m, applyProjectOperation(m.ctx, source, project, operation)
}

func applyProjectOperation(ctx context.Context, source ui.Source, project ui.Project, operation projectsync.OperationDefinition) tea.Cmd {
	return func() tea.Msg {
		client, err := projectsync.NewSourceClient(source.Type, nil)
		if err != nil {
			return operationAppliedMsg{title: operation.Title, projectID: project.ID, project: project.Name, result: projectsync.OperationResult{Outcome: projectsync.OperationOutcomeFailed}, err: err}
		}

		service := projectsync.OperationService{SourceClient: client}
		result, err := service.RunAndVerify(ctx, projectsync.Source{
			ID:       source.ID,
			Type:     source.Type,
			URL:      source.URL,
			PATToken: source.PATToken,
		}, projectsync.Project{
			ID:         project.ID,
			ProviderID: project.ProjectID,
			Name:       project.Name,
		}, operation.ID)

		return operationAppliedMsg{title: operation.Title, projectID: project.ID, project: project.Name, result: result, err: err}
	}
}
