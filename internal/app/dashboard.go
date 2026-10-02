package app

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dnwSilver/tld/internal/projectsync"
	"github.com/dnwSilver/tld/internal/storage"
	"github.com/dnwSilver/tld/internal/ui"
	"github.com/dnwSilver/tld/internal/version"
)

type dashboardAttentionLoadedMsg struct {
	rows []ui.DashboardAttentionRow
	err  error
}

type tokenRightsLoadedMsg struct {
	rights ui.TokenRights
	err    error
}

type dashboardState struct {
	attentionRows     []ui.DashboardAttentionRow
	tokenRights       ui.TokenRights
	focus             ui.DashboardPane
	selectedProjectID int64
}

func (m model) loadDashboardAttention() tea.Cmd {
	store := m.store
	return func() tea.Msg {
		if store == nil {
			return dashboardAttentionLoadedMsg{rows: []ui.DashboardAttentionRow{}}
		}

		ctx := m.ctx
		projects, err := store.Projects().List(ctx)
		if err != nil {
			return dashboardAttentionLoadedMsg{err: err}
		}
		sources, err := store.Sources().List(ctx)
		if err != nil {
			return dashboardAttentionLoadedMsg{err: err}
		}
		comparisons, err := store.ProjectDependencies().ListPolicyVersionComparisons(ctx)
		if err != nil {
			return dashboardAttentionLoadedMsg{err: err}
		}

		sourcesByID := make(map[int64]storage.Source, len(sources))
		for _, source := range sources {
			sourcesByID[source.ID] = source
		}

		cache := store.Cache()
		refs := make([]projectsync.ProjectSourceRef, 0, len(projects))
		reportIndex := make(map[int64]int, len(projects))
		for _, project := range projects {
			if project.EndOfLife {
				continue
			}
			source, ok := sourcesByID[project.SourceID]
			if !ok {
				continue
			}
			reportIndex[project.ID] = len(refs)
			refs = append(refs, projectsync.ProjectSourceRef{
				Source: projectsync.Source{ID: source.ID, Type: source.Type, URL: source.URL},
				Project: projectsync.Project{
					ProviderID: project.ProjectID,
					Name:       project.Name,
					StackName:  project.StackName,
				},
			})
		}
		vulnReports, err := (projectsync.VulnScanService{Cache: cache, Mode: projectsync.VulnScanModeProd}).LoadProjects(ctx, refs)
		if err != nil {
			return dashboardAttentionLoadedMsg{err: fmt.Errorf("load vulnerability snapshots: %w", err)}
		}
		checkReports, _, err := (projectsync.CheckService{Cache: cache}).LoadProjectsWithVersions(ctx, refs)
		if err != nil {
			return dashboardAttentionLoadedMsg{err: fmt.Errorf("load check snapshots: %w", err)}
		}
		vulnerabilities := make(map[int64]projectsync.VulnCounts, len(projects))
		vulnCoverage := make(map[int64]bool, len(projects))
		checks := make(map[int64]projectsync.ProjectCheckResults, len(projects))
		checkCoverage := make(map[int64]bool, len(projects))
		for _, project := range projects {
			if project.EndOfLife {
				continue
			}
			index, ok := reportIndex[project.ID]
			if !ok {
				continue
			}
			report := vulnReports[index]
			vulnerabilities[project.ID] = report.Counts
			vulnCoverage[project.ID] = report.Scanned

			results := checkReports[index]
			checks[project.ID] = results
			complete := len(results) == len(projectsync.ProjectChecks)
			for _, state := range results {
				if state == projectsync.CheckStateUnknown {
					complete = false
					break
				}
			}
			checkCoverage[project.ID] = complete
		}

		return dashboardAttentionLoadedMsg{
			rows: buildDashboardAttentionRowsWithCoverage(projects, vulnerabilities, checks, comparisons, vulnCoverage, checkCoverage),
		}
	}
}

func (m model) loadTokenRights() tea.Cmd {
	store := m.store
	return func() tea.Msg {
		if store == nil {
			return tokenRightsLoadedMsg{}
		}

		ctx := m.ctx
		sources, err := store.Sources().List(ctx)
		if err != nil {
			return tokenRightsLoadedMsg{err: err}
		}
		projects, err := store.Projects().List(ctx)
		if err != nil {
			return tokenRightsLoadedMsg{err: err}
		}

		type target struct {
			index   int
			source  storage.Source
			project storage.Project
		}
		sourcesByID := make(map[int64]storage.Source, len(sources))
		for _, source := range sources {
			sourcesByID[source.ID] = source
		}
		targets := make([]target, 0)
		for _, project := range projects {
			source, ok := sourcesByID[project.SourceID]
			if project.EndOfLife || !ok || !strings.EqualFold(strings.TrimSpace(source.Type), projectsync.SourceTypeGitLab) {
				continue
			}
			targets = append(targets, target{index: len(targets), source: source, project: project})
		}
		rights := make([]ui.ProjectRight, len(targets))
		jobs := make(chan target)
		workers := min(4, len(targets))
		var wg sync.WaitGroup
		wg.Add(workers)
		for range workers {
			go func() {
				defer wg.Done()
				for item := range jobs {
					result := ui.ProjectRight{ProjectID: item.project.ID, ProjectName: item.project.Name}
					client, err := projectsync.NewSourceClient(item.source.Type, nil)
					if err == nil {
						if rightsClient, ok := client.(projectsync.MaintainerRightsSourceClient); ok {
							result.Maintainer, err = rightsClient.HasMaintainerRights(ctx, projectsync.Source{
								ID: item.source.ID, Type: item.source.Type, URL: item.source.URL, PATToken: item.source.PATToken,
							}, projectsync.Project{ID: item.project.ID, ProviderID: item.project.ProjectID, Name: item.project.Name})
							result.Checked = err == nil
						}
					}
					rights[item.index] = result
				}
			}()
		}
		for _, item := range targets {
			jobs <- item
		}
		close(jobs)
		wg.Wait()
		return tokenRightsLoadedMsg{rights: ui.TokenRights{Projects: rights}}
	}
}

func buildDashboardAttentionRows(
	projects []storage.Project,
	vulnerabilities map[int64]projectsync.VulnCounts,
	checks map[int64]projectsync.ProjectCheckResults,
	comparisons []storage.PolicyVersionComparison,
) []ui.DashboardAttentionRow {
	return buildDashboardAttentionRowsWithCoverage(projects, vulnerabilities, checks, comparisons, nil, nil)
}

func buildDashboardAttentionRowsWithCoverage(
	projects []storage.Project,
	vulnerabilities map[int64]projectsync.VulnCounts,
	checks map[int64]projectsync.ProjectCheckResults,
	comparisons []storage.PolicyVersionComparison,
	vulnCoverage map[int64]bool,
	checkCoverage map[int64]bool,
) []ui.DashboardAttentionRow {
	rowsByProject := make(map[int64]*ui.DashboardAttentionRow, len(projects))
	for _, project := range projects {
		if project.EndOfLife {
			continue
		}
		rowsByProject[project.ID] = &ui.DashboardAttentionRow{
			ProjectID:   project.ID,
			ProjectName: project.Name,
			DataUnknown: (vulnCoverage != nil && !vulnCoverage[project.ID]) || (checkCoverage != nil && !checkCoverage[project.ID]),
			Critical:    vulnerabilities[project.ID].Critical,
			High:        vulnerabilities[project.ID].High,
		}
		for _, state := range checks[project.ID] {
			switch state {
			case projectsync.CheckStateFail:
				rowsByProject[project.ID].SettingsErrors++
			case projectsync.CheckStateWarning:
				rowsByProject[project.ID].SettingsWarnings++
			}
		}
	}

	for _, comparison := range comparisons {
		row := rowsByProject[comparison.ProjectID]
		if row == nil {
			continue
		}
		switch dependencyVersionLag(comparison.Actual, comparison.Policy) {
		case versionLagMajor:
			row.Major++
		case versionLagMinorPatch:
			row.MinorPatch++
		}
	}

	rows := make([]ui.DashboardAttentionRow, 0, len(rowsByProject))
	for _, row := range rowsByProject {
		if !row.DataUnknown && row.Critical == 0 &&
			row.High == 0 &&
			row.Major == 0 &&
			row.MinorPatch == 0 &&
			row.SettingsErrors == 0 &&
			row.SettingsWarnings == 0 {
			continue
		}
		rows = append(rows, *row)
	}
	sort.SliceStable(rows, func(i, j int) bool {
		left := [...]int{rows[i].Critical, rows[i].High, rows[i].Major, rows[i].MinorPatch, rows[i].SettingsErrors, rows[i].SettingsWarnings}
		right := [...]int{rows[j].Critical, rows[j].High, rows[j].Major, rows[j].MinorPatch, rows[j].SettingsErrors, rows[j].SettingsWarnings}
		for index := range left {
			if left[index] != right[index] {
				return left[index] > right[index]
			}
		}
		return strings.ToLower(rows[i].ProjectName) < strings.ToLower(rows[j].ProjectName)
	})

	return rows
}

type versionLag int

const (
	versionLagNone versionLag = iota
	versionLagMajor
	versionLagMinorPatch
)

func dependencyVersionLag(actual, policy string) versionLag {
	comparison, actualMajor, policyMajor, ok := version.CompareExact(actual, policy)
	if !ok {
		return versionLagNone
	}
	if actualMajor < policyMajor {
		return versionLagMajor
	}
	if actualMajor == policyMajor && comparison < 0 {
		return versionLagMinorPatch
	}
	return versionLagNone
}
