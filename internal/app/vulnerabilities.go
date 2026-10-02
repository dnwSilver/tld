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
	selectedProjectID int64
	mode              ui.VulnMode
	generation        uint64
	rows              []ui.VulnProjectRow
	items             []ui.VulnerabilityItem
	err               error
}

type vulnSyncMsg struct {
	runID   uint64
	message string
	err     error
	done    bool
	step    bool
	current int
	total   int
}

func (m model) loadVulnerabilities() tea.Cmd {
	projects := activeProjects(m.projects)
	sources := m.sourcesWithCredentials()
	selectedProjectID := m.selectedVulnProjectID
	mode := projectsync.VulnScanMode(m.vulnMode.Title())
	uiMode := m.vulnMode
	var generation uint64
	if m.vulnsGeneration != nil {
		generation = m.vulnsGeneration.Add(1)
	}
	return func() tea.Msg {
		if m.store == nil {
			return vulnsLoadedMsg{selectedProjectID: selectedProjectID, mode: uiMode, generation: generation, rows: []ui.VulnProjectRow{}, items: []ui.VulnerabilityItem{}}
		}

		cache := m.store.Cache()
		refs := make([]projectsync.VulnProjectRef, 0, len(projects))
		reportIndex := make(map[int64]int, len(projects))
		for _, project := range projects {
			source, ok := findByID(sources, project.SourceID, sourceID)
			if !ok {
				continue
			}
			reportIndex[project.ID] = len(refs)
			refs = append(refs, projectsync.VulnProjectRef{
				Source:  projectsync.Source{ID: source.ID, Type: source.Type, URL: source.URL},
				Project: projectsync.Project{ProviderID: project.ProjectID, StackName: project.StackName},
			})
		}
		service := projectsync.VulnScanService{Cache: cache, Mode: mode}
		reports, err := service.LoadProjects(m.ctx, refs)
		if err != nil {
			return vulnsLoadedMsg{selectedProjectID: selectedProjectID, mode: uiMode, generation: generation, err: err}
		}
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
			index, ok := reportIndex[project.ID]
			if !ok {
				rows = append(rows, row)
				continue
			}
			report := reports[index]
			row.Scanned = report.Scanned
			if !report.ScannedAt.IsZero() {
				row.ScannedAt = report.ScannedAt.UTC().Format("2006-01-02 15:04 UTC")
			}
			row.Revision = report.Revision
			row.Coverage = vulnerabilityCoverageSummary(report.Coverage)
			row.LastOutcome = string(report.LastAttempt.Outcome)
			if !report.LastAttempt.At.IsZero() {
				row.LastAttemptAt = report.LastAttempt.At.UTC().Format("2006-01-02 15:04 UTC")
			}
			row.Stale = report.Scanned && report.LastAttempt.Outcome != "" && report.LastAttempt.Outcome != projectsync.VulnScanComplete && report.LastAttempt.At.After(report.ScannedAt)
			row.Counts = toUIVulnCounts(report.Counts)
			rows = append(rows, row)
			if project.ID == selectedProjectID {
				items = toUIVulnerabilityItems(report.Items)
			}
		}

		return vulnsLoadedMsg{selectedProjectID: selectedProjectID, mode: uiMode, generation: generation, rows: rows, items: items}
	}
}

func vulnerabilityCoverageSummary(coverage []projectsync.ScannerCoverage) string {
	if len(coverage) == 0 {
		return "coverage unknown"
	}
	parts := make([]string, 0, len(coverage))
	for _, scanner := range coverage {
		part := scanner.Scanner
		switch {
		case scanner.Packages > 0:
			part += fmt.Sprintf(" %d packages", scanner.Packages)
		case scanner.Targets > 0:
			part += fmt.Sprintf(" %d targets", scanner.Targets)
		default:
			part += " (" + scanner.Method + ")"
		}
		parts = append(parts, part)
	}
	return strings.Join(parts, ", ")
}

func (m model) startVulnsRefresh(projects []ui.Project) (tea.Model, tea.Cmd) {
	if m.vulnsStatus.Running || len(projects) == 0 {
		return m, nil
	}

	ch := make(chan vulnSyncMsg, 16)
	jobCtx, cancel := context.WithCancel(m.ctx)
	m.vulnsCancel = cancel
	m.vulnsRunID++
	m.vulnsSyncCh = ch
	m.vulnsStatus = ui.SettingsStatus{
		Message: "Scanning vulnerabilities...",
		Running: true,
	}

	mode := projectsync.VulnScanMode(m.vulnMode.Title())
	return m, tea.Batch(runVulnsRefresh(jobCtx, m.store, projects, m.sourcesWithCredentials(), mode, ch), waitVulnSync(ch, m.vulnsRunID))
}

func (m model) selectedVulnProjects() []ui.Project {
	if project, ok := findByID(m.projects, m.selectedVulnProjectID, projectID); ok && !project.EndOfLife {
		return []ui.Project{project}
	}
	return nil
}

func runVulnsRefresh(
	ctx context.Context,
	store *storage.Store,
	projects []ui.Project,
	sources []ui.Source,
	mode projectsync.VulnScanMode,
	ch chan<- vulnSyncMsg,
) tea.Cmd {
	return func() tea.Msg {
		defer close(ch)
		if store == nil {
			sendJobMsg(ctx, ch, vulnSyncMsg{message: "store is not ready", err: errors.New("store is not ready"), done: true})
			return nil
		}

		cache := store.Cache()
		runProjectBatch(ctx, projects, "Scans", "Vulnerability scan complete", func(ctx context.Context, project ui.Project, progress func(string)) error {
			if _, err := projectsync.ResolveVulnStrategy(project.StackName, mode); err != nil {
				return err
			}

			source, ok := findByID(sources, project.SourceID, sourceID)
			if !ok {
				return errors.New("source not found")
			}

			client, err := projectsync.NewSourceClient(source.Type, nil)
			if err != nil {
				return err
			}

			service := projectsync.VulnScanService{
				Cache:        cache,
				SourceClient: client,
				Mode:         mode,
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
				StackName:  project.StackName,
			}, progress)
			return err
		}, func(event projectJobEvent) {
			sendJobMsg(ctx, ch, vulnSyncMsg{message: event.message, err: event.err, done: event.done, step: event.step, current: event.current, total: event.total})
		})
		return nil
	}
}

func waitVulnSync(ch <-chan vulnSyncMsg, runID uint64) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return nil
		}
		msg.runID = runID
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
