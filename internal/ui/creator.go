package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dnwSilver/tld/internal/ui/components"
	"github.com/dnwSilver/tld/internal/ui/screens"
	"github.com/dnwSilver/tld/internal/ui/uikit"
)

type Creator struct {
	palette               uikit.Palette
	hints                 components.Hints
	navModal              components.NavModal
	logo                  components.Logo
	defaultScreen         screens.DefaultScreen
	stacksScreen          screens.StacksScreen
	namespacesScreen      screens.StacksScreen
	dependenciesScreen    screens.DependenciesScreen
	projectsScreen        screens.ProjectsScreen
	sourcesScreen         screens.SourcesScreen
	policiesScreen        screens.PoliciesScreen
	viewScreen            screens.DependencyViewScreen
	settingsScreen        screens.SettingsScreen
	releasesScreen        screens.ReleasesScreen
	vulnerabilitiesScreen screens.VulnerabilitiesScreen
}

func NewCreator() Creator {
	palette := uikit.NewPalette()

	return Creator{
		palette:               palette,
		hints:                 components.NewHints(palette),
		navModal:              components.NewNavModal(palette),
		logo:                  components.NewLogo(palette),
		defaultScreen:         screens.NewDefaultScreen(palette),
		stacksScreen:          screens.NewStacksScreen(palette),
		namespacesScreen:      screens.NewNamespacesScreen(palette),
		dependenciesScreen:    screens.NewDependenciesScreen(palette),
		projectsScreen:        screens.NewProjectsScreen(palette),
		sourcesScreen:         screens.NewSourcesScreen(palette),
		policiesScreen:        screens.NewPoliciesScreen(palette),
		viewScreen:            screens.NewDependencyViewScreen(palette),
		settingsScreen:        screens.NewSettingsScreen(palette),
		releasesScreen:        screens.NewReleasesScreen(palette),
		vulnerabilitiesScreen: screens.NewVulnerabilitiesScreen(palette),
	}
}

func (c Creator) Render(
	width int,
	height int,
	screen uikit.Screen,
	stacks []uikit.Stack,
	dashboardAttentionRows []uikit.DashboardAttentionRow,
	dashboardFocus uikit.DashboardPane,
	selectedAttentionProjectID int64,
	selectedStackID int64,
	namespaces []uikit.Namespace,
	selectedNamespaceID int64,
	dependencies []uikit.Dependency,
	selectedDependencyID int64,
	projects []uikit.Project,
	selectedProjectID int64,
	projectDependencies []uikit.ProjectDependency,
	selectedProjectDepID int64,
	projectLatestRun uikit.ProjectDependencyRun,
	projectSyncStatus uikit.ProjectSyncStatus,
	projectFocus uikit.ProjectPane,
	sources []uikit.Source,
	selectedSourceID int64,
	policies []uikit.Policy,
	selectedPolicyID int64,
	policyValues []uikit.PolicyValue,
	selectedPolicyValueID int64,
	policyFocus uikit.PolicyPane,
	policyUpdateStatus uikit.SettingsStatus,
	dependencyView uikit.DependencyView,
	selectedViewProjectID int64,
	viewColumnOffset int,
	checkColumns []uikit.ProjectCheck,
	projectCheckRows []uikit.ProjectCheckRow,
	selectedCheckProjectID int64,
	settingsStatus uikit.SettingsStatus,
	releaseRows []uikit.ReleaseRow,
	selectedReleaseProjectID int64,
	releasePeriod uikit.ReleasePeriod,
	releasesStatus uikit.SettingsStatus,
	vulnRows []uikit.VulnProjectRow,
	selectedVulnProjectID int64,
	vulnItems []uikit.VulnerabilityItem,
	selectedVulnItemIndex int,
	vulnFocus uikit.VulnPane,
	vulnMode uikit.VulnMode,
	vulnsStatus uikit.SettingsStatus,
	stackForm uikit.StackForm,
	namespaceForm uikit.StackForm,
	dependencyForm uikit.DependencyForm,
	projectForm uikit.ProjectForm,
	sourceForm uikit.SourceForm,
	policyForm uikit.PolicyForm,
	policyValueForm uikit.PolicyValueForm,
	policyUpdateForm uikit.PolicyUpdateForm,
	deleteConfirm uikit.DeleteConfirm,
	namespaceDeleteConfirm uikit.DeleteConfirm,
	dependencyDeleteConfirm uikit.DeleteConfirm,
	projectDeleteConfirm uikit.DeleteConfirm,
	sourceDeleteConfirm uikit.DeleteConfirm,
	policyDeleteConfirm uikit.DeleteConfirm,
	navModalOpen bool,
	navModalIndex int,
) string {
	base := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Background(c.palette.Background).
		Foreground(c.palette.Primary)

	top := c.renderTopBar(width, screen)
	topHeight := lipgloss.Height(top)
	bodyHeight := uikit.Max(height-topHeight, 1)
	body := c.defaultScreen.Render(width, bodyHeight, len(stacks), dashboardAttentionRows, dashboardFocus, selectedAttentionProjectID)
	if screen == uikit.ScreenStacks {
		body = c.stacksScreen.Render(width, bodyHeight, stacks, selectedStackID, nil, stackForm, deleteConfirm)
	}
	if screen == uikit.ScreenNamespaces {
		body = c.namespacesScreen.Render(
			width,
			bodyHeight,
			toNamespaceItems(namespaces, policies),
			selectedNamespaceID,
			policies,
			namespaceForm,
			namespaceDeleteConfirm,
		)
	}
	if screen == uikit.ScreenDependencies {
		body = c.dependenciesScreen.Render(
			width,
			bodyHeight,
			dependencies,
			selectedDependencyID,
			stacks,
			registrySources(sources),
			dependencyForm,
			dependencyDeleteConfirm,
		)
	}
	if screen == uikit.ScreenProjects {
		body = c.projectsScreen.Render(
			width,
			bodyHeight,
			projects,
			selectedProjectID,
			projectDependencies,
			selectedProjectDepID,
			projectLatestRun,
			projectSyncStatus,
			projectFocus,
			namespaces,
			projectSources(sources),
			stacks,
			projectForm,
			projectDeleteConfirm,
		)
	}
	if screen == uikit.ScreenSources {
		body = c.sourcesScreen.Render(
			width,
			bodyHeight,
			sources,
			selectedSourceID,
			sourceForm,
			sourceDeleteConfirm,
		)
	}
	if screen == uikit.ScreenPolicies {
		body = c.policiesScreen.Render(
			width,
			bodyHeight,
			policies,
			selectedPolicyID,
			policyValues,
			selectedPolicyValueID,
			policyFocus,
			namespaces,
			dependencies,
			stacks,
			registrySources(sources),
			policyForm,
			policyValueForm,
			policyUpdateForm,
			policyUpdateStatus,
			policyDeleteConfirm,
		)
	}
	if screen == uikit.ScreenView {
		body = c.viewScreen.Render(width, bodyHeight, stacks, dependencyView.StackID, dependencyView, selectedViewProjectID, viewColumnOffset, projectSyncStatus)
	}
	if screen == uikit.ScreenSettings {
		body = c.settingsScreen.Render(width, bodyHeight, checkColumns, projectCheckRows, selectedCheckProjectID, settingsStatus)
	}
	if screen == uikit.ScreenReleases {
		body = c.releasesScreen.Render(width, bodyHeight, releaseRows, selectedReleaseProjectID, releasePeriod, releasesStatus)
	}
	if screen == uikit.ScreenVulnerabilities {
		body = c.vulnerabilitiesScreen.Render(width, bodyHeight, vulnRows, selectedVulnProjectID, vulnItems, selectedVulnItemIndex, vulnFocus, vulnMode, vulnsStatus)
	}

	result := base.Render(lipgloss.JoinVertical(lipgloss.Left, top, body))
	if !navModalOpen {
		return result
	}

	lines := strings.Split(result, uikit.SymbolLineBreak)
	block := c.navModal.Render(width, navModalIndex, screen)
	modal := components.NewModal(c.palette, c.palette.Primary)
	lines = modal.Overlay(lines, width, height, block)

	return strings.Join(lines, uikit.SymbolLineBreak)
}

func registrySources(sources []uikit.Source) []uikit.Source {
	result := make([]uikit.Source, 0, len(sources))
	for _, source := range sources {
		if source.Type == "registry" {
			result = append(result, source)
		}
	}
	return result
}

func projectSources(sources []uikit.Source) []uikit.Source {
	result := make([]uikit.Source, 0, len(sources))
	for _, source := range sources {
		if source.Type != "registry" {
			result = append(result, source)
		}
	}
	return result
}

func toNamespaceItems(namespaces []uikit.Namespace, policies []uikit.Policy) []uikit.Stack {
	items := make([]uikit.Stack, 0, len(namespaces))
	for _, namespace := range namespaces {
		items = append(items, uikit.Stack{
			ID:         namespace.ID,
			Icon:       namespace.Icon,
			Name:       namespace.Name,
			Color:      namespace.Color,
			PolicyName: policyNameForNamespace(policies, namespace.PolicyID),
		})
	}

	return items
}

func policyNameForNamespace(policies []uikit.Policy, policyID int64) string {
	for _, policy := range policies {
		if policy.ID == policyID {
			return policy.Name + "[" + uikit.FormatInt(policy.DependencyCount) + "]"
		}
	}

	return ""
}

func (c Creator) renderTopBar(width int, screen uikit.Screen) string {
	hints := c.hints.Render(screen)
	logo := c.logo.Render()
	topHeight := uikit.Max(lipgloss.Height(hints), lipgloss.Height(logo))
	hints = uikit.PadBlockHeight(c.palette, hints, topHeight)
	logo = uikit.PadBlockHeight(c.palette, logo, topHeight)
	gapWidth := uikit.Max(width-lipgloss.Width(hints)-lipgloss.Width(logo), 0)
	gap := lipgloss.NewStyle().
		Width(gapWidth).
		Height(topHeight).
		Background(c.palette.Background).
		Render("")

	return lipgloss.NewStyle().
		Width(width).
		Background(c.palette.Background).
		Render(lipgloss.JoinHorizontal(lipgloss.Top, hints, gap, logo))
}
