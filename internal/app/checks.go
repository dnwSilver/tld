package app

import (
	"context"
	"errors"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dnwSilver/tld/internal/projectsync"
	"github.com/dnwSilver/tld/internal/storage"
	"github.com/dnwSilver/tld/internal/ui"
)

type projectChecksLoadedMsg struct {
	rows []ui.ProjectCheckRow
	err  error
}

type checkSyncMsg struct {
	message string
	err     error
	done    bool
	step    bool
	current int
	total   int
}

func defaultCheckColumns() []ui.ProjectCheck {
	columns := make([]ui.ProjectCheck, 0, len(projectsync.ProjectChecks))
	for _, check := range projectsync.ProjectChecks {
		columns = append(columns, ui.ProjectCheck{ID: check.ID, Title: check.Title})
	}
	return columns
}

func (m model) loadProjectChecks() tea.Cmd {
	projects := activeProjects(m.projects)
	sources := m.sources
	return func() tea.Msg {
		if m.store == nil {
			return projectChecksLoadedMsg{rows: []ui.ProjectCheckRow{}}
		}

		cache := m.store.Cache()
		rows := make([]ui.ProjectCheckRow, 0, len(projects))
		for _, project := range projects {
			row := ui.ProjectCheckRow{
				ProjectID:        project.ID,
				ProjectIcon:      project.Icon,
				ProjectName:      project.Name,
				ProjectColor:     project.Color,
				ProjectFreezing:  project.Freezing,
				ProjectEndOfLife: project.EndOfLife,
				Results:          make(map[string]ui.CheckState, len(projectsync.ProjectChecks)),
				Versions:         make(map[string]string, len(projectsync.ProjectChecks)),
			}
			source, ok := findByID(sources, project.SourceID, sourceID)
			if !ok {
				for _, check := range projectsync.ProjectChecks {
					row.Results[check.ID] = ui.CheckStateUnknown
				}
				rows = append(rows, row)
				continue
			}

			service := projectsync.CheckService{Cache: cache}
			results, versions, err := service.LoadProjectWithVersions(context.Background(), projectsync.Source{Type: source.Type}, projectsync.Project{ProviderID: project.ProjectID})
			if err != nil {
				return projectChecksLoadedMsg{err: err}
			}
			for checkID, state := range results {
				row.Results[checkID] = ui.CheckState(state)
			}
			for checkID, version := range versions {
				row.Versions[checkID] = version
			}
			rows = append(rows, row)
		}

		return projectChecksLoadedMsg{rows: rows}
	}
}

func (m model) startProjectChecksRefresh(projects []ui.Project) (tea.Model, tea.Cmd) {
	if m.checksStatus.Running || len(projects) == 0 {
		return m, nil
	}

	ch := make(chan checkSyncMsg, 16)
	m.checksSyncCh = ch
	m.checksStatus = ui.SettingsStatus{
		Message: "Running checks...",
		Running: true,
	}

	return m, tea.Batch(runProjectChecksRefresh(m.store, projects, m.sources, ch), waitCheckSync(ch))
}

func (m model) selectedCheckProjects() []ui.Project {
	if project, ok := findByID(m.projects, m.selectedCheckProjectID, projectID); ok && !project.EndOfLife {
		return []ui.Project{project}
	}

	return nil
}

func runProjectChecksRefresh(store *storage.Store, projects []ui.Project, sources []ui.Source, ch chan<- checkSyncMsg) tea.Cmd {
	return func() tea.Msg {
		defer close(ch)
		if store == nil {
			ch <- checkSyncMsg{message: "store is not ready", err: errors.New("store is not ready"), done: true}
			return nil
		}

		cache := store.Cache()
		total := len(projects)
		for index, project := range projects {
			source, ok := findByID(sources, project.SourceID, sourceID)
			if !ok {
				ch <- checkSyncMsg{message: fmt.Sprintf("Skipping %s: source not found", project.Name), current: index, total: total}
				continue
			}

			client, err := projectsync.NewSourceClient(source.Type, nil)
			if err != nil {
				ch <- checkSyncMsg{message: err.Error(), err: err, done: true}
				return nil
			}

			service := projectsync.CheckService{
				Cache:        cache,
				SourceClient: client,
			}
			_, err = service.RunProject(context.Background(), projectsync.Source{
				ID:       source.ID,
				Type:     source.Type,
				URL:      source.URL,
				PATToken: source.PATToken,
			}, projectsync.Project{
				ID:         project.ID,
				ProviderID: project.ProjectID,
				Name:       project.Name,
			}, func(message string) {
				ch <- checkSyncMsg{message: project.Name + ": " + message, current: index, total: total}
			})
			if err != nil {
				ch <- checkSyncMsg{message: err.Error(), err: err, done: true}
				return nil
			}

			ch <- checkSyncMsg{message: project.Name, step: true, current: index + 1, total: total}
		}

		ch <- checkSyncMsg{message: "Checks complete", done: true, current: total, total: total}
		return nil
	}
}

func waitCheckSync(ch <-chan checkSyncMsg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return nil
		}
		return msg
	}
}

func (m *model) ensureSelectedCheckProject() {
	m.selectedCheckProjectID = ensureSelected(m.projectCheckRows, m.selectedCheckProjectID, projectCheckRowID)
}

func projectCheckRowID(row ui.ProjectCheckRow) int64 {
	return row.ProjectID
}

func (m *model) selectPreviousCheckProject() {
	m.selectedCheckProjectID = selectPrevious(m.projectCheckRows, m.selectedCheckProjectID, projectCheckRowID)
}

func (m *model) selectNextCheckProject() {
	m.selectedCheckProjectID = selectNext(m.projectCheckRows, m.selectedCheckProjectID, projectCheckRowID)
}
