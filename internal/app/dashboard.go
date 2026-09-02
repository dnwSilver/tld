package app

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dnwSilver/tld/internal/projectsync"
	"github.com/dnwSilver/tld/internal/storage"
	"github.com/dnwSilver/tld/internal/ui"
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

		ctx := context.Background()
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
		vulnerabilities := make(map[int64]projectsync.VulnCounts, len(projects))
		checks := make(map[int64]projectsync.ProjectCheckResults, len(projects))
		for _, project := range projects {
			if project.EndOfLife {
				continue
			}
			source, ok := sourcesByID[project.SourceID]
			if !ok {
				continue
			}
			syncSource := projectsync.Source{Type: source.Type}
			syncProject := projectsync.Project{
				ProviderID: project.ProjectID,
				Name:       project.Name,
				StackName:  project.StackName,
			}

			report, err := (projectsync.VulnScanService{
				Cache: cache,
				Mode:  projectsync.VulnScanModeProd,
			}).LoadProject(ctx, syncSource, syncProject)
			if err != nil {
				return dashboardAttentionLoadedMsg{err: fmt.Errorf("load vulnerabilities for %s: %w", project.Name, err)}
			}
			vulnerabilities[project.ID] = report.Counts

			results, err := (projectsync.CheckService{Cache: cache}).LoadProject(ctx, syncSource, syncProject)
			if err != nil {
				return dashboardAttentionLoadedMsg{err: fmt.Errorf("load settings for %s: %w", project.Name, err)}
			}
			checks[project.ID] = results
		}

		return dashboardAttentionLoadedMsg{
			rows: buildDashboardAttentionRows(projects, vulnerabilities, checks, comparisons),
		}
	}
}

// loadTokenRights probes the first active GitLab project: the Maintainer
// role is a per-project relation, so one representative project is enough
// for a team-wide token.
func (m model) loadTokenRights() tea.Cmd {
	store := m.store
	return func() tea.Msg {
		if store == nil {
			return tokenRightsLoadedMsg{}
		}

		ctx := context.Background()
		sources, err := store.Sources().List(ctx)
		if err != nil {
			return tokenRightsLoadedMsg{err: err}
		}
		projects, err := store.Projects().List(ctx)
		if err != nil {
			return tokenRightsLoadedMsg{err: err}
		}

		source, project, ok := firstGitLabProject(sources, projects)
		if !ok {
			return tokenRightsLoadedMsg{}
		}

		client, err := projectsync.NewSourceClient(source.Type, nil)
		if err != nil {
			return tokenRightsLoadedMsg{err: err}
		}
		rightsClient, ok := client.(projectsync.MaintainerRightsSourceClient)
		if !ok {
			return tokenRightsLoadedMsg{}
		}

		maintainer, err := rightsClient.HasMaintainerRights(ctx, projectsync.Source{
			ID:       source.ID,
			Type:     source.Type,
			URL:      source.URL,
			PATToken: source.PATToken,
		}, projectsync.Project{
			ID:         project.ID,
			ProviderID: project.ProjectID,
			Name:       project.Name,
		})
		if err != nil {
			return tokenRightsLoadedMsg{err: err}
		}

		return tokenRightsLoadedMsg{rights: ui.TokenRights{Checked: true, Maintainer: maintainer}}
	}
}

func firstGitLabProject(sources []storage.Source, projects []storage.Project) (storage.Source, storage.Project, bool) {
	sourcesByID := make(map[int64]storage.Source, len(sources))
	for _, source := range sources {
		sourcesByID[source.ID] = source
	}

	for _, project := range projects {
		if project.EndOfLife {
			continue
		}
		source, ok := sourcesByID[project.SourceID]
		if !ok {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(source.Type), projectsync.SourceTypeGitLab) {
			return source, project, true
		}
	}

	return storage.Source{}, storage.Project{}, false
}

func buildDashboardAttentionRows(
	projects []storage.Project,
	vulnerabilities map[int64]projectsync.VulnCounts,
	checks map[int64]projectsync.ProjectCheckResults,
	comparisons []storage.PolicyVersionComparison,
) []ui.DashboardAttentionRow {
	rowsByProject := make(map[int64]*ui.DashboardAttentionRow, len(projects))
	for _, project := range projects {
		if project.EndOfLife {
			continue
		}
		rowsByProject[project.ID] = &ui.DashboardAttentionRow{
			ProjectID:   project.ID,
			ProjectName: project.Name,
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
		if row.Critical == 0 &&
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

var dashboardVersionNumberPattern = regexp.MustCompile(`\d+`)

func dependencyVersionLag(actual, policy string) versionLag {
	actualParts, actualOK := dashboardVersionParts(actual)
	policyParts, policyOK := dashboardVersionParts(policy)
	if !actualOK || !policyOK {
		return versionLagNone
	}
	if actualParts[0] < policyParts[0] {
		return versionLagMajor
	}
	if actualParts[0] == policyParts[0] && compareDashboardVersions(actualParts, policyParts) < 0 {
		return versionLagMinorPatch
	}
	return versionLagNone
}

func dashboardVersionParts(version string) ([]int, bool) {
	matches := dashboardVersionNumberPattern.FindAllString(strings.TrimSpace(version), -1)
	if len(matches) == 0 {
		return nil, false
	}
	parts := make([]int, 0, len(matches))
	for _, match := range matches {
		value, err := strconv.Atoi(match)
		if err != nil {
			return nil, false
		}
		parts = append(parts, value)
	}
	return parts, true
}

func compareDashboardVersions(left, right []int) int {
	length := len(left)
	if len(right) > length {
		length = len(right)
	}
	for index := range length {
		var leftValue, rightValue int
		if index < len(left) {
			leftValue = left[index]
		}
		if index < len(right) {
			rightValue = right[index]
		}
		if leftValue < rightValue {
			return -1
		}
		if leftValue > rightValue {
			return 1
		}
	}
	return 0
}
