package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dnwSilver/tld/internal/projectsync"
	"github.com/dnwSilver/tld/internal/storage"
	"github.com/dnwSilver/tld/internal/ui"
)

type releasesLoadedMsg struct {
	rows []ui.ReleaseRow
	err  error
}

type releaseSyncMsg struct {
	message string
	err     error
	done    bool
	step    bool
	current int
	total   int
}

func (m model) loadReleases() tea.Cmd {
	projects := m.projects
	sources := m.sources
	return func() tea.Msg {
		if m.store == nil {
			return releasesLoadedMsg{rows: []ui.ReleaseRow{}}
		}

		cache := m.store.Cache()
		now := time.Now().UTC()
		period := m.releasePeriod
		rows := make([]ui.ReleaseRow, 0, len(projects))
		for _, project := range projects {
			row := ui.ReleaseRow{
				ProjectID:    project.ID,
				ProjectIcon:  project.Icon,
				ProjectName:  project.Name,
				ProjectColor: project.Color,
				Months:       buildReleaseMonths(now, nil, period),
			}
			source, ok := findByID(sources, project.SourceID, sourceID)
			if !ok {
				rows = append(rows, row)
				continue
			}

			service := projectsync.ReleaseService{Cache: cache}
			dates, err := service.LoadProject(context.Background(), projectsync.Source{Type: source.Type}, projectsync.Project{ProviderID: project.ProjectID})
			if err != nil {
				return releasesLoadedMsg{err: err}
			}
			row.HasReleases = len(dates) > 0
			row.Months = buildReleaseMonths(now, dates, period)
			rows = append(rows, row)
		}

		return releasesLoadedMsg{rows: rows}
	}
}

func (m model) startReleasesRefresh(projects []ui.Project) (tea.Model, tea.Cmd) {
	if m.releasesStatus.Running || len(projects) == 0 {
		return m, nil
	}

	ch := make(chan releaseSyncMsg, 16)
	m.releasesSyncCh = ch
	m.releasesStatus = ui.SettingsStatus{
		Message: "Loading releases...",
		Running: true,
	}

	return m, tea.Batch(runReleasesRefresh(m.store, projects, m.sources, ch), waitReleaseSync(ch))
}

func (m model) selectedReleaseProjects() []ui.Project {
	if project, ok := findByID(m.projects, m.selectedReleaseProjectID, projectID); ok {
		return []ui.Project{project}
	}

	return nil
}

func runReleasesRefresh(store *storage.Store, projects []ui.Project, sources []ui.Source, ch chan<- releaseSyncMsg) tea.Cmd {
	return func() tea.Msg {
		defer close(ch)
		if store == nil {
			ch <- releaseSyncMsg{message: "store is not ready", err: errors.New("store is not ready"), done: true}
			return nil
		}

		cache := store.Cache()
		total := len(projects)
		for index, project := range projects {
			source, ok := findByID(sources, project.SourceID, sourceID)
			if !ok {
				ch <- releaseSyncMsg{message: fmt.Sprintf("Skipping %s: source not found", project.Name), current: index, total: total}
				continue
			}

			client, err := projectsync.NewSourceClient(source.Type, nil)
			if err != nil {
				ch <- releaseSyncMsg{message: err.Error(), err: err, done: true}
				return nil
			}

			service := projectsync.ReleaseService{
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
				ch <- releaseSyncMsg{message: project.Name + ": " + message, current: index, total: total}
			})
			if err != nil {
				ch <- releaseSyncMsg{message: err.Error(), err: err, done: true}
				return nil
			}

			ch <- releaseSyncMsg{message: project.Name, step: true, current: index + 1, total: total}
		}

		ch <- releaseSyncMsg{message: "Releases complete", done: true, current: total, total: total}
		return nil
	}
}

func waitReleaseSync(ch <-chan releaseSyncMsg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return nil
		}
		return msg
	}
}

func buildReleaseMonths(now time.Time, dates []time.Time, period ui.ReleasePeriod) []ui.ReleaseMonth {
	count := period.Months()
	months := make([]ui.ReleaseMonth, count)
	starts := make([]time.Time, count)
	base := monthStart(now)
	for index := range months {
		starts[index] = base.AddDate(0, -(count - 1 - index), 0)
		slotCount := monthSlotCount(starts[index], period)
		months[index] = ui.ReleaseMonth{
			Label:     starts[index].Format("Jan 06"),
			SlotCount: slotCount,
			Marks:     make([]bool, slotCount),
		}
	}

	for _, date := range dates {
		if date.IsZero() {
			continue
		}
		date = date.UTC()
		for index := range months {
			next := starts[index].AddDate(0, 1, 0)
			if !date.Before(starts[index]) && date.Before(next) {
				slot := daySlot(date.Day(), period)
				if slot >= 0 && slot < months[index].SlotCount {
					months[index].Marks[slot] = true
				}
				break
			}
		}
	}

	return months
}

func monthSlotCount(value time.Time, period ui.ReleasePeriod) int {
	if period == ui.ReleasePeriodQuarter {
		return daysInMonth(value)
	}

	return 4
}

func daySlot(day int, period ui.ReleasePeriod) int {
	if period == ui.ReleasePeriodQuarter {
		return day - 1
	}
	switch {
	case day <= 7:
		return 0
	case day <= 15:
		return 1
	case day <= 23:
		return 2
	default:
		return 3
	}
}

func daysInMonth(value time.Time) int {
	return time.Date(value.Year(), value.Month()+1, 0, 0, 0, 0, 0, value.Location()).Day()
}

func monthStart(value time.Time) time.Time {
	year, month, _ := value.Date()
	return time.Date(year, month, 1, 0, 0, 0, 0, value.Location())
}

func (m *model) ensureSelectedReleaseProject() {
	m.selectedReleaseProjectID = ensureSelected(m.releaseRows, m.selectedReleaseProjectID, releaseRowID)
}

func releaseRowID(row ui.ReleaseRow) int64 {
	return row.ProjectID
}

func (m *model) selectPreviousReleaseProject() {
	m.selectedReleaseProjectID = selectPrevious(m.releaseRows, m.selectedReleaseProjectID, releaseRowID)
}

func (m *model) selectNextReleaseProject() {
	m.selectedReleaseProjectID = selectNext(m.releaseRows, m.selectedReleaseProjectID, releaseRowID)
}
