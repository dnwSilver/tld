package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/dnwSilver/tld/internal/safetext"
	"github.com/dnwSilver/tld/internal/ui/components"
	"github.com/dnwSilver/tld/internal/ui/screens"
	"github.com/dnwSilver/tld/internal/ui/uikit"
)

type Creator struct {
	palette               uikit.Palette
	hints                 components.Hints
	navModal              components.NavModal
	operationsModal       components.OperationsModal
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
		operationsModal:       components.NewOperationsModal(palette),
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

func (c Creator) Render(state RenderState) string {
	width, height, screen := state.Width, state.Height, state.Screen
	base := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Background(c.palette.Background).
		Foreground(c.palette.Primary)

	top := c.renderTopBar(width, screen)
	topHeight := lipgloss.Height(top)
	searchLine := c.renderSearchLine(width, state.Overlay)
	searchHeight := 0
	if searchLine != "" {
		searchHeight = lipgloss.Height(searchLine)
	}
	loadLine := c.renderLoadState(width, state.Overlay.LoadState)
	errorLine := ""
	if state.Overlay.GlobalError != "" {
		plain := safetext.Plain("Error: " + state.Overlay.GlobalError + "  [Esc to dismiss]")
		errorLine = lipgloss.NewStyle().Foreground(c.palette.Error).Background(c.palette.Background).Render(ansi.Cut(plain, 0, width))
	}
	bodyHeight := uikit.Max(height-topHeight-searchHeight-lipgloss.Height(loadLine)-lipgloss.Height(errorLine), 1)
	body := ""
	if screen == uikit.ScreenDefault {
		body = c.defaultScreen.Render(width, bodyHeight, len(state.Stacks.Items), state.Dashboard.TokenRights, state.Dashboard.AttentionRows, state.Dashboard.Focus, state.Dashboard.SelectedProjectID, state.Overlay.SearchQuery)
	}
	if screen == uikit.ScreenStacks {
		body = c.stacksScreen.Render(width, bodyHeight, state.Stacks.Items, state.Stacks.SelectedID, nil, state.Stacks.Form, state.Stacks.Delete, state.Overlay.SearchQuery)
	}
	if screen == uikit.ScreenNamespaces {
		body = c.namespacesScreen.Render(
			width,
			bodyHeight,
			toNamespaceItems(state.Namespaces.Items, state.Policies.Items),
			state.Namespaces.SelectedID,
			state.Policies.Items,
			state.Namespaces.Form,
			state.Namespaces.Delete,
			state.Overlay.SearchQuery,
		)
	}
	if screen == uikit.ScreenDependencies {
		body = c.dependenciesScreen.Render(
			width,
			bodyHeight,
			state.Dependencies.Items,
			state.Dependencies.SelectedID,
			state.Stacks.Items,
			registrySources(state.Sources.Items),
			state.Dependencies.Form,
			state.Dependencies.Delete,
			state.Overlay.SearchQuery,
		)
	}
	if screen == uikit.ScreenProjects {
		body = c.projectsScreen.Render(
			width,
			bodyHeight,
			state.Projects.Items,
			state.Projects.SelectedID,
			state.Projects.Dependencies,
			state.Projects.SelectedDependencyID,
			state.Projects.LatestRun,
			state.Projects.SyncStatus,
			state.Projects.Focus,
			state.Namespaces.Items,
			projectSources(state.Sources.Items),
			state.Stacks.Items,
			state.Projects.Form,
			state.Projects.Delete,
			state.Overlay.SearchQuery,
		)
	}
	if screen == uikit.ScreenSources {
		body = c.sourcesScreen.Render(
			width,
			bodyHeight,
			state.Sources.Items,
			state.Sources.SelectedID,
			state.Sources.Form,
			state.Sources.Delete,
			state.Overlay.SearchQuery,
		)
	}
	if screen == uikit.ScreenPolicies {
		body = c.policiesScreen.Render(
			width,
			bodyHeight,
			state.Policies.Items,
			state.Policies.SelectedID,
			state.Policies.Values,
			state.Policies.SelectedValueID,
			state.Policies.Focus,
			state.Namespaces.Items,
			state.Dependencies.Items,
			state.Stacks.Items,
			registrySources(state.Sources.Items),
			state.Policies.Form,
			state.Policies.ValueForm,
			state.Policies.UpdateForm,
			state.Policies.UpdateStatus,
			state.Policies.Delete,
			state.Overlay.SearchQuery,
		)
	}
	if screen == uikit.ScreenView {
		body = c.viewScreen.Render(width, bodyHeight, state.Stacks.Items, state.View.View.StackID, state.View.View, state.View.SelectedProjectID, state.View.ColumnOffset, state.Projects.SyncStatus, state.Overlay.SearchQuery)
	}
	if screen == uikit.ScreenSettings {
		body = c.settingsScreen.Render(width, bodyHeight, state.Settings.Columns, state.Settings.Rows, state.Settings.SelectedProjectID, state.Settings.TableState, state.Settings.Status, state.Overlay.SearchQuery)
	}
	if screen == uikit.ScreenReleases {
		body = c.releasesScreen.Render(width, bodyHeight, state.Releases.Rows, state.Releases.SelectedProjectID, state.Releases.Period, state.Releases.Status, state.Overlay.SearchQuery)
	}
	if screen == uikit.ScreenVulnerabilities {
		body = c.vulnerabilitiesScreen.Render(width, bodyHeight, state.Vulnerabilities.Rows, state.Vulnerabilities.SelectedProjectID, state.Vulnerabilities.Items, state.Vulnerabilities.SelectedItemIndex, state.Vulnerabilities.Focus, state.Vulnerabilities.Mode, state.Vulnerabilities.Status, state.Overlay.SearchQuery)
	}

	parts := []string{top}
	if searchLine != "" {
		parts = append(parts, searchLine)
	}
	if loadLine != "" {
		parts = append(parts, loadLine)
	}
	if errorLine != "" {
		parts = append(parts, errorLine)
	}
	parts = append(parts, body)
	result := base.Render(lipgloss.JoinVertical(lipgloss.Left, parts...))
	if !state.Overlay.NavigationOpen && !state.Overlay.OperationsOpen {
		return result
	}

	lines := strings.Split(result, uikit.SymbolLineBreak)
	modal := components.NewModal(c.palette, c.palette.Primary)
	if state.Overlay.NavigationOpen {
		block := c.navModal.Render(width, height, state.Overlay.NavigationIndex, screen)
		lines = modal.Overlay(lines, width, height, block)
	}
	if state.Overlay.OperationsOpen {
		projectName := checkProjectName(state.Settings.Rows, state.Settings.SelectedProjectID)
		preview := ""
		if state.Overlay.OperationsConfirm {
			for _, project := range state.Projects.Items {
				if project.ID == state.Settings.SelectedProjectID {
					for _, source := range state.Sources.Items {
						if source.ID == project.SourceID {
							preview = source.URL + " / " + project.ProjectID
							break
						}
					}
				}
			}
		}
		block := c.operationsModal.Render(width, height, state.Overlay.Operations, state.Overlay.OperationsIndex, projectName, preview, state.Overlay.OperationsConfirm)
		lines = modal.Overlay(lines, width, height, block)
	}

	return strings.Join(lines, uikit.SymbolLineBreak)
}

func (c Creator) renderLoadState(width int, state uikit.LoadState) string {
	message := ""
	color := c.palette.Hint
	switch state.Phase {
	case uikit.LoadPhaseLoading:
		message = "Loading..."
	case uikit.LoadPhaseStale:
		message = "STALE: " + state.Error + "  [Ctrl+R to retry]"
		color = c.palette.Warning
	case uikit.LoadPhaseError:
		message = "Load failed: " + state.Error + "  [Ctrl+R to retry]"
		color = c.palette.Error
	}
	if message == "" {
		return ""
	}
	return lipgloss.NewStyle().Foreground(color).Background(c.palette.Background).Render(
		ansi.Cut(safetext.Plain(message), 0, width),
	)
}

func checkProjectName(rows []uikit.ProjectCheckRow, selectedProjectID int64) string {
	for _, row := range rows {
		if row.ProjectID == selectedProjectID {
			return row.ProjectName
		}
	}

	return ""
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
