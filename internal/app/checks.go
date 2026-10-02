package app

import (
	"cmp"
	"context"
	"errors"
	"sort"
	"strings"

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
	runID   uint64
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
	sources := m.sourcesWithCredentials()
	return func() tea.Msg {
		if m.store == nil {
			return projectChecksLoadedMsg{rows: []ui.ProjectCheckRow{}}
		}

		cache := m.store.Cache()
		refs := make([]projectsync.ProjectSourceRef, 0, len(projects))
		reportIndex := make(map[int64]int, len(projects))
		for _, project := range projects {
			source, ok := findByID(sources, project.SourceID, sourceID)
			if !ok {
				continue
			}
			reportIndex[project.ID] = len(refs)
			refs = append(refs, projectsync.ProjectSourceRef{
				Source:  projectsync.Source{ID: source.ID, Type: source.Type, URL: source.URL},
				Project: projectsync.Project{ProviderID: project.ProjectID},
			})
		}
		service := projectsync.CheckService{Cache: cache}
		allResults, allVersions, err := service.LoadProjectsWithVersions(m.ctx, refs)
		if err != nil {
			return projectChecksLoadedMsg{err: err}
		}
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
			index, ok := reportIndex[project.ID]
			if !ok {
				for _, check := range projectsync.ProjectChecks {
					row.Results[check.ID] = ui.CheckStateUnknown
				}
				rows = append(rows, row)
				continue
			}

			for checkID, state := range allResults[index] {
				row.Results[checkID] = ui.CheckState(state)
			}
			for checkID, version := range allVersions[index] {
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
	jobCtx, cancel := context.WithCancel(m.ctx)
	m.checksCancel = cancel
	m.checksRunID++
	m.checksSyncCh = ch
	m.checksStatus = ui.SettingsStatus{
		Message: "Running checks...",
		Running: true,
	}

	return m, tea.Batch(runProjectChecksRefresh(jobCtx, m.store, projects, m.sourcesWithCredentials(), ch), waitCheckSync(ch, m.checksRunID))
}

func (m model) selectedCheckProjects() []ui.Project {
	if project, ok := findByID(m.projects, m.selectedCheckProjectID, projectID); ok && !project.EndOfLife {
		return []ui.Project{project}
	}

	return nil
}

func runProjectChecksRefresh(ctx context.Context, store *storage.Store, projects []ui.Project, sources []ui.Source, ch chan<- checkSyncMsg) tea.Cmd {
	return func() tea.Msg {
		defer close(ch)
		if store == nil {
			sendJobMsg(ctx, ch, checkSyncMsg{message: "store is not ready", err: errors.New("store is not ready"), done: true})
			return nil
		}

		cache := store.Cache()
		runProjectBatch(ctx, projects, "Checks", "Checks complete", func(ctx context.Context, project ui.Project, progress func(string)) error {
			source, ok := findByID(sources, project.SourceID, sourceID)
			if !ok {
				return errors.New("source not found")
			}

			client, err := projectsync.NewSourceClient(source.Type, nil)
			if err != nil {
				return err
			}

			service := projectsync.CheckService{
				Cache:        cache,
				SourceClient: client,
			}
			_, err = service.RunProject(ctx, projectsync.Source{
				ID:       source.ID,
				Type:     source.Type,
				URL:      source.URL,
				PATToken: source.PATToken,
			}, projectsync.Project{
				ID:         project.ID,
				ProviderID: project.ProjectID,
				Name:       project.Name,
			}, progress)
			return err
		}, func(event projectJobEvent) {
			sendJobMsg(ctx, ch, checkSyncMsg{message: event.message, err: event.err, done: event.done, step: event.step, current: event.current, total: event.total})
		})
		return nil
	}
}

func waitCheckSync(ch <-chan checkSyncMsg, runID uint64) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return nil
		}
		msg.runID = runID
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

func (m *model) selectPreviousCheckColumn() {
	if m.settingsTableState.SelectedColumn > 0 {
		m.settingsTableState.SelectedColumn--
	}
}

func (m *model) selectNextCheckColumn() {
	if m.settingsTableState.SelectedColumn < len(m.checkColumns) {
		m.settingsTableState.SelectedColumn++
	}
}

func (m *model) sortProjectChecksBySelectedColumn() {
	if m.settingsTableState.SortActive && m.settingsTableState.SortColumn == m.settingsTableState.SelectedColumn {
		m.settingsTableState.SortDescending = !m.settingsTableState.SortDescending
	} else {
		m.settingsTableState.SortActive = true
		m.settingsTableState.SortColumn = m.settingsTableState.SelectedColumn
		m.settingsTableState.SortDescending = false
	}

	m.applyProjectCheckSort()
}

func (m *model) applyProjectCheckSort() {
	if !m.settingsTableState.SortActive || len(m.projectCheckRows) < 2 {
		return
	}

	column := m.settingsTableState.SortColumn
	descending := m.settingsTableState.SortDescending
	sort.SliceStable(m.projectCheckRows, func(i, j int) bool {
		comparison := compareProjectCheckRows(m.projectCheckRows[i], m.projectCheckRows[j], m.checkColumns, column)
		if descending {
			return comparison > 0
		}
		return comparison < 0
	})
}

func compareProjectCheckRows(left, right ui.ProjectCheckRow, columns []ui.ProjectCheck, column int) int {
	if column > 0 && column <= len(columns) {
		checkID := columns[column-1].ID
		if comparison := cmp.Compare(checkStateSortRank(left.Results[checkID]), checkStateSortRank(right.Results[checkID])); comparison != 0 {
			return comparison
		}
		if comparison := compareFoldedStrings(left.Versions[checkID], right.Versions[checkID]); comparison != 0 {
			return comparison
		}
	}

	if comparison := compareFoldedStrings(left.ProjectName, right.ProjectName); comparison != 0 {
		return comparison
	}
	return cmp.Compare(left.ProjectID, right.ProjectID)
}

func checkStateSortRank(state ui.CheckState) int {
	switch state {
	case ui.CheckStateFail:
		return 0
	case ui.CheckStateWarning:
		return 1
	case ui.CheckStateUnknown:
		return 2
	case ui.CheckStatePass:
		return 3
	case ui.CheckStateNotApplicable:
		return 4
	default:
		return 2
	}
}

func compareFoldedStrings(left, right string) int {
	if comparison := strings.Compare(strings.ToLower(left), strings.ToLower(right)); comparison != 0 {
		return comparison
	}
	return strings.Compare(left, right)
}
