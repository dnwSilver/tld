package app

import (
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dnwSilver/tld/internal/ui"
)

type searchMatch struct {
	id int64
}

func (m model) searchMatches() []searchMatch {
	query := strings.ToLower(strings.TrimSpace(m.searchQuery))
	matches := make([]searchMatch, 0)
	add := func(id int64, label string) {
		if query == "" || strings.Contains(strings.ToLower(label), query) {
			matches = append(matches, searchMatch{id: id})
		}
	}
	switch m.screen {
	case ui.ScreenDefault:
		for _, row := range m.dashboard.attentionRows {
			add(row.ProjectID, row.ProjectName)
		}
	case ui.ScreenStacks:
		for _, row := range m.stacks {
			add(row.ID, row.Name)
		}
	case ui.ScreenNamespaces:
		for _, row := range m.namespaces {
			add(row.ID, row.Name)
		}
	case ui.ScreenDependencies:
		for _, row := range m.dependencies {
			add(row.ID, row.Name)
		}
	case ui.ScreenProjects:
		for _, row := range m.projects {
			add(row.ID, row.Name)
		}
	case ui.ScreenSources:
		for _, row := range m.sources {
			add(row.ID, row.Name)
		}
	case ui.ScreenPolicies:
		for _, row := range m.policies {
			add(row.ID, row.Name)
		}
	case ui.ScreenView:
		for _, row := range m.dependencyView.Rows {
			add(row.ProjectID, row.ProjectName)
		}
	case ui.ScreenSettings:
		for _, row := range m.projectCheckRows {
			add(row.ProjectID, row.ProjectName)
		}
	case ui.ScreenReleases:
		for _, row := range m.releaseRows {
			add(row.ProjectID, row.ProjectName)
		}
	case ui.ScreenVulnerabilities:
		for _, row := range m.vulnRows {
			add(row.ProjectID, row.ProjectName)
		}
	}
	return matches
}

func (m model) updateSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.searchOpen = false
		m.searchQuery = ""
		return m.reconcileSearchSelection()
	case "enter":
		m.searchOpen = false
		return m, nil
	case "up", "ctrl+p":
		return m.stepSearchSelection(-1)
	case "down", "ctrl+n":
		return m.stepSearchSelection(1)
	case "backspace", "ctrl+h":
		runes := []rune(m.searchQuery)
		if len(runes) > 0 {
			m.searchQuery = string(runes[:len(runes)-1])
		}
		return m.reconcileSearchSelection()
	case "ctrl+u":
		m.searchQuery = ""
		return m.reconcileSearchSelection()
	}
	if len(msg.Runes) > 0 && len([]rune(m.searchQuery)) < 120 {
		for _, r := range msg.Runes {
			if unicode.IsControl(r) {
				return m, nil
			}
		}
		m.searchQuery += string(msg.Runes)
		return m.reconcileSearchSelection()
	}
	return m, nil
}

func (m model) reconcileSearchSelection() (tea.Model, tea.Cmd) {
	matches := m.searchMatches()
	selected := m.currentListSelection()
	for _, match := range matches {
		if match.id == selected {
			return m, nil
		}
	}
	if len(matches) == 0 {
		return m.selectSearchMatch(0)
	}
	return m.selectSearchMatch(matches[0].id)
}

func (m model) stepSearchSelection(delta int) (tea.Model, tea.Cmd) {
	matches := m.searchMatches()
	if len(matches) == 0 {
		return m, nil
	}
	current := m.currentListSelection()
	index := -1
	for i, match := range matches {
		if match.id == current {
			index = i
			break
		}
	}
	if index < 0 {
		index = 0
	} else {
		index = max(0, min(index+delta, len(matches)-1))
	}
	if matches[index].id == current {
		return m, nil
	}
	return m.selectSearchMatch(matches[index].id)
}

func (m model) searchTargetPaneActive() bool {
	return !((m.screen == ui.ScreenProjects && m.projectFocus == ui.ProjectPaneDependencies) ||
		(m.screen == ui.ScreenPolicies && m.policyFocus == ui.PolicyPaneValues) ||
		(m.screen == ui.ScreenVulnerabilities && m.vulnFocus == ui.VulnPaneDetails) ||
		(m.screen == ui.ScreenDefault && m.dashboard.focus == ui.DashboardPaneSummary))
}

func filterSearchRows[T any](rows []T, query string, name func(T) string) []T {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return rows
	}
	filtered := make([]T, 0, len(rows))
	for _, row := range rows {
		if strings.Contains(strings.ToLower(name(row)), query) {
			filtered = append(filtered, row)
		}
	}
	return filtered
}

func (m model) filterSearchRenderState(state *ui.RenderState) {
	query := m.searchQuery
	switch m.screen {
	case ui.ScreenDefault:
		state.Dashboard.AttentionRows = filterSearchRows(state.Dashboard.AttentionRows, query, func(r ui.DashboardAttentionRow) string { return r.ProjectName })
	case ui.ScreenStacks:
		state.Stacks.Items = filterSearchRows(state.Stacks.Items, query, func(r ui.Stack) string { return r.Name })
	case ui.ScreenNamespaces:
		state.Namespaces.Items = filterSearchRows(state.Namespaces.Items, query, func(r ui.Namespace) string { return r.Name })
	case ui.ScreenDependencies:
		state.Dependencies.Items = filterSearchRows(state.Dependencies.Items, query, func(r ui.Dependency) string { return r.Name })
	case ui.ScreenProjects:
		state.Projects.Items = filterSearchRows(state.Projects.Items, query, func(r ui.Project) string { return r.Name })
	case ui.ScreenSources:
		state.Sources.Items = filterSearchRows(state.Sources.Items, query, func(r ui.Source) string { return r.Name })
	case ui.ScreenPolicies:
		state.Policies.Items = filterSearchRows(state.Policies.Items, query, func(r ui.Policy) string { return r.Name })
	case ui.ScreenView:
		state.View.View.Rows = filterSearchRows(state.View.View.Rows, query, func(r ui.DependencyViewRow) string { return r.ProjectName })
	case ui.ScreenSettings:
		state.Settings.Rows = filterSearchRows(state.Settings.Rows, query, func(r ui.ProjectCheckRow) string { return r.ProjectName })
	case ui.ScreenReleases:
		state.Releases.Rows = filterSearchRows(state.Releases.Rows, query, func(r ui.ReleaseRow) string { return r.ProjectName })
	case ui.ScreenVulnerabilities:
		state.Vulnerabilities.Rows = filterSearchRows(state.Vulnerabilities.Rows, query, func(r ui.VulnProjectRow) string { return r.ProjectName })
	}
}

func (m model) jumpList(key string) (tea.Model, tea.Cmd) {
	if (m.screen == ui.ScreenProjects && m.projectFocus == ui.ProjectPaneDependencies) ||
		(m.screen == ui.ScreenPolicies && m.policyFocus == ui.PolicyPaneValues) ||
		(m.screen == ui.ScreenVulnerabilities && m.vulnFocus == ui.VulnPaneDetails) {
		return m, nil
	}
	matches := m.searchMatches()
	if len(matches) == 0 {
		return m, nil
	}
	selected := m.currentListSelection()
	index := 0
	for i, match := range matches {
		if match.id == selected {
			index = i
			break
		}
	}
	pageSize := max(m.height-12, 1)
	switch key {
	case "home":
		index = 0
	case "end":
		index = len(matches) - 1
	case "pgup":
		index = max(index-pageSize, 0)
	case "pgdown":
		index = min(index+pageSize, len(matches)-1)
	}
	return m.selectSearchMatch(matches[index].id)
}

func (m model) currentListSelection() int64 {
	switch m.screen {
	case ui.ScreenDefault:
		return m.dashboard.selectedProjectID
	case ui.ScreenStacks:
		return m.selectedStackID
	case ui.ScreenNamespaces:
		return m.selectedNamespaceID
	case ui.ScreenDependencies:
		return m.selectedDependencyID
	case ui.ScreenProjects:
		return m.selectedProjectID
	case ui.ScreenSources:
		return m.selectedSourceID
	case ui.ScreenPolicies:
		return m.selectedPolicyID
	case ui.ScreenView:
		return m.selectedViewProjectID
	case ui.ScreenSettings:
		return m.selectedCheckProjectID
	case ui.ScreenReleases:
		return m.selectedReleaseProjectID
	case ui.ScreenVulnerabilities:
		return m.selectedVulnProjectID
	default:
		return 0
	}
}

func (m model) selectSearchMatch(id int64) (tea.Model, tea.Cmd) {
	if id == 0 {
		switch m.screen {
		case ui.ScreenDefault:
			m.dashboard.selectedProjectID = 0
		case ui.ScreenStacks:
			m.selectedStackID = 0
		case ui.ScreenNamespaces:
			m.selectedNamespaceID = 0
		case ui.ScreenDependencies:
			m.selectedDependencyID = 0
		case ui.ScreenProjects:
			m.selectedProjectID = 0
			m.projectDependencies = nil
		case ui.ScreenSources:
			m.selectedSourceID = 0
		case ui.ScreenPolicies:
			m.selectedPolicyID = 0
			m.policyValues = nil
		case ui.ScreenView:
			m.selectedViewProjectID = 0
		case ui.ScreenSettings:
			m.selectedCheckProjectID = 0
		case ui.ScreenReleases:
			m.selectedReleaseProjectID = 0
		case ui.ScreenVulnerabilities:
			m.selectedVulnProjectID = 0
			m.vulnItems = nil
		}
		return m, nil
	}
	switch m.screen {
	case ui.ScreenDefault:
		m.dashboard.focus = ui.DashboardPaneAttention
		m.dashboard.selectedProjectID = id
	case ui.ScreenStacks:
		m.selectedStackID = id
	case ui.ScreenNamespaces:
		m.selectedNamespaceID = id
	case ui.ScreenDependencies:
		m.selectedDependencyID = id
	case ui.ScreenProjects:
		m.projectFocus = ui.ProjectPaneProjects
		m.selectedProjectID = id
		return m, m.loadProjectDependencies()
	case ui.ScreenSources:
		m.selectedSourceID = id
	case ui.ScreenPolicies:
		m.policyFocus = ui.PolicyPanePolicies
		m.selectedPolicyID = id
		return m, m.loadPolicyValues()
	case ui.ScreenView:
		m.selectedViewProjectID = id
	case ui.ScreenSettings:
		m.selectedCheckProjectID = id
	case ui.ScreenReleases:
		m.selectedReleaseProjectID = id
	case ui.ScreenVulnerabilities:
		m.vulnFocus = ui.VulnPaneProjects
		m.selectedVulnProjectID = id
		return m, m.loadVulnerabilities()
	}
	return m, nil
}
