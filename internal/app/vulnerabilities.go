package app

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dnwSilver/tld/internal/projectsync"
	"github.com/dnwSilver/tld/internal/storage"
	"github.com/dnwSilver/tld/internal/ui"
)

type vulnsLoadedMsg struct {
	rows  []ui.VulnProjectRow
	items []ui.VulnerabilityItem
	err   error
}

type vulnSyncMsg struct {
	message string
	err     error
	done    bool
	step    bool
	current int
	total   int
}

func (m model) loadVulnerabilities() tea.Cmd {
	projects := activeProjects(m.projects)
	sources := m.sources
	selectedProjectID := m.selectedVulnProjectID
	return func() tea.Msg {
		if m.store == nil {
			return vulnsLoadedMsg{rows: []ui.VulnProjectRow{}, items: []ui.VulnerabilityItem{}}
		}

		cache := m.store.Cache()
		rows := make([]ui.VulnProjectRow, 0, len(projects))
		var items []ui.VulnerabilityItem
		for _, project := range projects {
			row := ui.VulnProjectRow{
				ProjectID:        project.ID,
				ProjectIcon:      project.Icon,
				ProjectName:      project.Name,
				ProjectColor:     project.Color,
				ProjectFreezing:  project.Freezing,
				ProjectEndOfLife: project.EndOfLife,
			}
			source, ok := findByID(sources, project.SourceID, sourceID)
			if !ok {
				rows = append(rows, row)
				continue
			}

			service := projectsync.VulnScanService{Cache: cache}
			report, err := service.LoadProject(context.Background(), projectsync.Source{Type: source.Type}, projectsync.Project{
				ProviderID: project.ProjectID,
				StackName:  project.StackName,
			})
			if err != nil {
				return vulnsLoadedMsg{err: err}
			}
			row.Scanned = report.Scanned
			row.Counts = toUIVulnCounts(report.Counts)
			rows = append(rows, row)
			if project.ID == selectedProjectID {
				items = toUIVulnerabilityItems(report.Items)
			}
		}

		return vulnsLoadedMsg{rows: rows, items: items}
	}
}

func (m model) startVulnsRefresh(projects []ui.Project) (tea.Model, tea.Cmd) {
	if m.vulnsStatus.Running || len(projects) == 0 {
		return m, nil
	}

	ch := make(chan vulnSyncMsg, 16)
	m.vulnsSyncCh = ch
	m.vulnsStatus = ui.SettingsStatus{
		Message: "Scanning vulnerabilities...",
		Running: true,
	}

	return m, tea.Batch(runVulnsRefresh(m.store, projects, m.sources, ch), waitVulnSync(ch))
}

func (m model) selectedVulnProjects() []ui.Project {
	if project, ok := findByID(m.projects, m.selectedVulnProjectID, projectID); ok && !project.EndOfLife {
		return []ui.Project{project}
	}
	return nil
}

func runVulnsRefresh(store *storage.Store, projects []ui.Project, sources []ui.Source, ch chan<- vulnSyncMsg) tea.Cmd {
	return func() tea.Msg {
		defer close(ch)
		if store == nil {
			ch <- vulnSyncMsg{message: "store is not ready", err: errors.New("store is not ready"), done: true}
			return nil
		}

		cache := store.Cache()
		total := len(projects)
		for index, project := range projects {
			if _, err := projectsync.ResolveVulnStrategy(project.StackName); err != nil {
				ch <- vulnSyncMsg{message: fmt.Sprintf("Skipping %s: %s", project.Name, err.Error()), current: index, total: total}
				continue
			}

			source, ok := findByID(sources, project.SourceID, sourceID)
			if !ok {
				ch <- vulnSyncMsg{message: fmt.Sprintf("Skipping %s: source not found", project.Name), current: index, total: total}
				continue
			}

			client, err := projectsync.NewSourceClient(source.Type, nil)
			if err != nil {
				ch <- vulnSyncMsg{message: err.Error(), err: err, done: true}
				return nil
			}

			service := projectsync.VulnScanService{
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
				StackName:  project.StackName,
			}, func(message string) {
				ch <- vulnSyncMsg{message: project.Name + ": " + message, current: index, total: total}
			})
			if err != nil {
				ch <- vulnSyncMsg{message: err.Error(), err: err, done: true}
				return nil
			}

			ch <- vulnSyncMsg{message: project.Name, step: true, current: index + 1, total: total}
		}

		ch <- vulnSyncMsg{message: "Vulnerability scan complete", done: true, current: total, total: total}
		return nil
	}
}

func waitVulnSync(ch <-chan vulnSyncMsg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return nil
		}
		return msg
	}
}

func toUIVulnCounts(counts projectsync.VulnCounts) ui.VulnCounts {
	return ui.VulnCounts{
		Critical: counts.Critical,
		High:     counts.High,
		Medium:   counts.Medium,
		Low:      counts.Low,
		None:     counts.None,
	}
}

func toUIVulnerabilityItems(items []projectsync.Vulnerability) []ui.VulnerabilityItem {
	result := make([]ui.VulnerabilityItem, 0, len(items))
	for _, item := range items {
		result = append(result, ui.VulnerabilityItem{
			Package:     item.Package,
			Severity:    string(item.Severity),
			Title:       item.Title,
			Description: item.Description,
			Range:       item.Range,
		})
	}
	sort.Slice(result, func(i, j int) bool {
		left := vulnSeverityRank(result[i].Severity)
		right := vulnSeverityRank(result[j].Severity)
		if left != right {
			return left < right
		}
		titleI := result[i].Title
		if titleI == "" {
			titleI = result[i].Package
		}
		titleJ := result[j].Title
		if titleJ == "" {
			titleJ = result[j].Package
		}
		return titleI < titleJ
	})
	return result
}

func vulnSeverityRank(severity string) int {
	switch strings.ToLower(strings.TrimSpace(severity)) {
	case "critical":
		return 0
	case "high":
		return 1
	case "medium", "moderate":
		return 2
	case "low":
		return 3
	default:
		return 4
	}
}

func (m *model) ensureSelectedVulnProject() {
	m.selectedVulnProjectID = ensureSelected(m.vulnRows, m.selectedVulnProjectID, vulnRowID)
}

func vulnRowID(row ui.VulnProjectRow) int64 {
	return row.ProjectID
}

func (m *model) selectPreviousVulnProject() {
	m.selectedVulnProjectID = selectPrevious(m.vulnRows, m.selectedVulnProjectID, vulnRowID)
}

func (m *model) selectNextVulnProject() {
	m.selectedVulnProjectID = selectNext(m.vulnRows, m.selectedVulnProjectID, vulnRowID)
}

func (m *model) selectPreviousVulnItem() {
	if len(m.vulnItems) == 0 {
		m.selectedVulnItemIndex = 0
		return
	}
	if m.selectedVulnItemIndex > 0 {
		m.selectedVulnItemIndex--
	}
}

func (m *model) selectNextVulnItem() {
	if len(m.vulnItems) == 0 {
		m.selectedVulnItemIndex = 0
		return
	}
	if m.selectedVulnItemIndex < len(m.vulnItems)-1 {
		m.selectedVulnItemIndex++
	}
}

func (m *model) toggleVulnPane() {
	if m.screen != ui.ScreenVulnerabilities {
		return
	}
	if m.vulnFocus == ui.VulnPaneProjects {
		m.vulnFocus = ui.VulnPaneDetails
		if m.selectedVulnItemIndex >= len(m.vulnItems) {
			m.selectedVulnItemIndex = 0
		}
		return
	}
	m.vulnFocus = ui.VulnPaneProjects
}
