package screens

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dnwSilver/tld/internal/ui/components"
	"github.com/dnwSilver/tld/internal/ui/uikit"
)

type ProjectsScreen struct {
	palette uikit.Palette
}

const (
	projectFormLabelWidth       = 10
	projectFormValueWidth       = 24
	projectIDColumnWidth        = 10
	projectNameColumnWidth      = 20
	projectNamespaceColumnWidth = 15
	projectSourceColumnWidth    = 13
	projectDepsColumnWidth      = 6
	projectGapColumnWidth       = 2
	projectIconSlotWidth        = 2
	projectIconColumnWidth      = projectIconSlotWidth * 2
	projectDepVersionWidth      = 12
)

func NewProjectsScreen(palette uikit.Palette) ProjectsScreen {
	return ProjectsScreen{palette: palette}
}

func (s ProjectsScreen) Render(
	width int,
	height int,
	projects []uikit.Project,
	selectedProjectID int64,
	projectDependencies []uikit.ProjectDependency,
	selectedProjectDepID int64,
	projectLatestRun uikit.ProjectDependencyRun,
	projectSyncStatus uikit.ProjectSyncStatus,
	focus uikit.ProjectPane,
	namespaces []uikit.Namespace,
	sources []uikit.Source,
	stacks []uikit.Stack,
	form uikit.ProjectForm,
	deleteConfirm uikit.DeleteConfirm,
) string {
	boxHeight := uikit.Max(height-1, 3)
	contentHeight := uikit.Max(boxHeight-2, 1)
	modalOpen := form.Open || deleteConfirm.Open
	leftTotalWidth := uikit.Max(width*3/5, 2)
	rightTotalWidth := uikit.Max(width-leftTotalWidth, 2)
	leftContentWidth := uikit.Max(leftTotalWidth-2, 1)
	rightContentWidth := uikit.Max(rightTotalWidth-2, 1)

	count := len(projects)
	leftTitle := components.ScreenTitle(s.palette, uikit.SymbolProject, "Projects", &count)
	leftBorder := s.borderForPane(!modalOpen && focus == uikit.ProjectPaneProjects)
	if modalOpen {
		leftBorder = s.palette.Hint
	}
	left := components.NewBox(s.palette, leftBorder).Render(leftContentWidth, contentHeight, leftTitle, s.renderProjectsContent(leftContentWidth, contentHeight, projects, selectedProjectID))
	rightTitle := components.ScreenTitle(s.palette, uikit.SymbolDependency, "Dependencies", intPtr(len(projectDependencies)))
	rightBorder := s.borderForPane(!modalOpen && focus == uikit.ProjectPaneDependencies)
	right := components.NewBox(s.palette, rightBorder).Render(rightContentWidth, contentHeight, rightTitle, s.renderDependencyPanel(rightContentWidth, contentHeight, projectDependencies, selectedProjectDepID, projectLatestRun))
	body := lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	lines := strings.Split(body, uikit.SymbolLineBreak)
	if form.Open {
		lines = s.renderModal(lines, width, boxHeight, namespaces, sources, stacks, form)
	}
	if deleteConfirm.Open {
		lines = s.renderDeleteConfirm(lines, width, boxHeight, deleteConfirm)
	}
	footer := components.ScreenFooter(s.palette, width, "Kolosov Aleksandr")
	if projectSyncStatus.Message != "" || projectSyncStatus.Error != "" {
		footer = s.progressLine(width, projectSyncStatus)
	}

	return lipgloss.JoinVertical(lipgloss.Left, strings.Join(lines, uikit.SymbolLineBreak), footer)
}

func (s ProjectsScreen) borderForPane(active bool) lipgloss.Color {
	if active {
		return s.palette.Primary
	}

	return s.palette.Hint
}

func (s ProjectsScreen) renderProjectsContent(
	width int,
	height int,
	projects []uikit.Project,
	selectedProjectID int64,
) string {
	lines := make([]string, 0, height)
	iconColumnWidth := components.ColumnWidth(projects, projectIcons, projectIconColumnWidth)
	lines = append(lines, s.tableHeader(width, iconColumnWidth))

	if len(projects) == 0 {
		empty := uikit.Text(s.palette, s.palette.Hint, "No projects yet")
		lines = append(lines, uikit.CenterLine(s.palette, width, empty))
	} else {
		for _, project := range projects {
			lines = append(lines, s.renderProjectRow(width, project, project.ID == selectedProjectID, iconColumnWidth))
		}
	}

	return fillLines(s.palette, lines, width, height)
}

func (s ProjectsScreen) tableHeader(width int, iconColumnWidth int) string {
	return components.RenderTableRow(s.palette, s.palette.Background, width, []components.TableCell{
		{Value: "", Width: iconColumnWidth, Foreground: s.palette.Hint},
		{Value: "name", Width: projectNameColumnWidth, Foreground: s.palette.Hint, Bold: true},
		{Value: "namespace", Width: projectNamespaceColumnWidth, Foreground: s.palette.Hint, Bold: true},
		{Value: "source", Width: projectSourceColumnWidth, Foreground: s.palette.Hint, Bold: true},
		{Value: "deps", Width: projectDepsColumnWidth, Foreground: s.palette.Hint, Bold: true},
		{Value: "", Width: projectGapColumnWidth, Foreground: s.palette.Hint},
		{Value: "project_id", Width: projectIDColumnWidth, Foreground: s.palette.Hint, Bold: true},
	})
}

func (s ProjectsScreen) renderProjectRow(width int, project uikit.Project, selected bool, iconColumnWidth int) string {
	background := s.palette.Background
	if selected {
		background = s.palette.Hover
	}
	iconColor := lipgloss.Color(uikit.NormalizeHexColor(project.Color))
	stackColor := lipgloss.Color(uikit.NormalizeHexColor(project.StackColor))
	iconCell := lipgloss.NewStyle().
		Background(background).
		Foreground(iconColor).
		Width(projectIconSlotWidth).
		Render(project.Icon) + lipgloss.NewStyle().
		Background(background).
		Foreground(stackColor).
		Width(projectIconSlotWidth).
		Render(project.StackIcon)

	return components.RenderTableRow(s.palette, background, width, []components.TableCell{
		{Value: iconCell, Width: iconColumnWidth, Foreground: s.palette.Text},
		{Value: project.Name, Width: projectNameColumnWidth, Foreground: s.palette.Text},
		{Value: project.NamespaceName, Width: projectNamespaceColumnWidth, Foreground: s.palette.Hint},
		{Value: project.SourceName, Width: projectSourceColumnWidth, Foreground: s.palette.Info},
		{Value: rightAligned(uikit.FormatInt(project.DependencyCount), projectDepsColumnWidth), Width: projectDepsColumnWidth, Foreground: s.palette.Primary},
		{Value: "", Width: projectGapColumnWidth, Foreground: s.palette.Hint},
		{Value: project.ProjectID, Width: projectIDColumnWidth, Foreground: s.palette.Info},
	})
}

func projectIcons(project uikit.Project) string {
	return project.Icon + project.StackIcon
}

func (s ProjectsScreen) renderDependencyPanel(width int, height int, dependencies []uikit.ProjectDependency, selectedID int64, latestRun uikit.ProjectDependencyRun) string {
	lines := make([]string, 0, height)
	dependencyWidth := uikit.Max(width-projectDepVersionWidth, 1)
	lines = append(lines, components.RenderTableRow(s.palette, s.palette.Background, width, []components.TableCell{
		{Value: "dependency", Width: dependencyWidth, Foreground: s.palette.Hint, Bold: true},
		{Value: "version", Width: projectDepVersionWidth, Foreground: s.palette.Hint, Bold: true},
	}))

	if len(dependencies) == 0 {
		empty := uikit.Text(s.palette, s.palette.Hint, "No dependencies yet")
		lines = append(lines, uikit.CenterLine(s.palette, width, empty))
	} else {
		for _, dependency := range visibleProjectDependencies(dependencies, selectedID, height-1) {
			background := s.palette.Background
			if dependency.ID == selectedID {
				background = s.palette.Hover
			}
			lines = append(lines, components.RenderTableRow(s.palette, background, width, []components.TableCell{
				{Value: dependency.Name, Width: dependencyWidth, Foreground: s.palette.Text},
				{Value: dependency.Version, Width: projectDepVersionWidth, Foreground: s.palette.Text},
			}))
		}
	}

	for len(lines) < height {
		lines = append(lines, uikit.BackgroundSpaces(s.palette, width))
	}
	if len(lines) > height {
		lines = lines[:height]
	}

	return strings.Join(lines, uikit.SymbolLineBreak)
}

func visibleProjectDependencies(dependencies []uikit.ProjectDependency, selectedID int64, visibleRows int) []uikit.ProjectDependency {
	if visibleRows <= 0 || len(dependencies) <= visibleRows {
		return dependencies
	}

	selectedIndex := 0
	for index, dependency := range dependencies {
		if dependency.ID == selectedID {
			selectedIndex = index
			break
		}
	}

	start := selectedIndex - visibleRows + 1
	if start < 0 {
		start = 0
	}
	if start+visibleRows > len(dependencies) {
		start = len(dependencies) - visibleRows
	}

	return dependencies[start : start+visibleRows]
}

func (s ProjectsScreen) dependencyPanelTitle(width int, count int, latestRun uikit.ProjectDependencyRun) string {
	title := "Dependencies [" + uikit.FormatInt(count) + "]"
	if latestRun.HasValue {
		title += " · updated " + latestRun.UpdatedAt + " · " + latestRun.CommitShortSHA
	}

	return components.RenderTableRow(s.palette, s.palette.Background, width, []components.TableCell{
		{Value: title, Width: width, Foreground: s.palette.Primary, Bold: true},
	})
}

func (s ProjectsScreen) progressLine(width int, status uikit.ProjectSyncStatus) string {
	message := status.Message
	color := s.palette.Info
	if status.Error != "" {
		message = status.Error
		color = s.palette.Error
	}
	if status.Running {
		message = uikit.SymbolDependency + " " + message
	}

	return uikit.CenterLine(s.palette, width, uikit.Text(s.palette, color, message))
}

func (s ProjectsScreen) renderModal(lines []string, width, height int, namespaces []uikit.Namespace, sources []uikit.Source, stacks []uikit.Stack, form uikit.ProjectForm) []string {
	modal := components.NewModal(s.palette, s.palette.Primary)
	return modal.Overlay(lines, width, height, s.modal(width, namespaces, sources, stacks, form))
}

func (s ProjectsScreen) modal(width int, namespaces []uikit.Namespace, sources []uikit.Source, stacks []uikit.Stack, form uikit.ProjectForm) string {
	modalWidth := uikit.Min(uikit.Max(width-6, 42), 64)
	contentWidth := uikit.Max(modalWidth-2, 1)
	modal := components.NewModal(s.palette, s.palette.Primary)

	action := "Add"
	if form.Mode == uikit.StackFormModeEdit {
		action = "Edit"
	}
	title := modal.Title(uikit.SymbolProject + " " + action + " project")

	rows := []string{
		modal.CenterLine(contentWidth, s.inputLine("Icon", form.Icon, form.Focus == uikit.ProjectFormFieldIcon)),
		modal.CenterLine(contentWidth, s.inputLine("Color", form.Color, form.Focus == uikit.ProjectFormFieldColor)),
		modal.CenterLine(contentWidth, s.inputLine("Name", form.Name, form.Focus == uikit.ProjectFormFieldName)),
		modal.CenterLine(contentWidth, s.inputLine("Namespace", s.namespaceName(namespaces, form.NamespaceID), form.Focus == uikit.ProjectFormFieldNamespace)),
		modal.CenterLine(contentWidth, s.inputLine("Source", s.sourceName(sources, form.SourceID), form.Focus == uikit.ProjectFormFieldSource)),
		modal.CenterLine(contentWidth, s.inputLine("Stack", s.stackName(stacks, form.StackID), form.Focus == uikit.ProjectFormFieldStack)),
		modal.CenterLine(contentWidth, s.inputLine("Project ID", form.ProjectID, form.Focus == uikit.ProjectFormFieldProjectID)),
		modal.CenterLine(contentWidth, modal.Text(s.palette.Error, form.Error)),
		modal.CenterLine(contentWidth, s.actionsLine(form.CanSave, form.Focus)),
	}

	return modal.Render(contentWidth, title, rows)
}

func (s ProjectsScreen) inputLine(label, value string, focused bool) string {
	return components.RenderFormField(s.palette, components.FormField{
		Label:      label,
		Value:      value,
		Focused:    focused,
		LabelWidth: projectFormLabelWidth,
		ValueWidth: projectFormValueWidth,
	})
}

func (s ProjectsScreen) actionsLine(canSave bool, focus uikit.ProjectFormField) string {
	saveColor := s.palette.Disable
	if canSave {
		saveColor = s.palette.Primary
	}

	actions := []components.Action{
		{Hint: uikit.KeyCancel.Hint, Color: s.palette.Primary},
		{Hint: uikit.KeySave.Hint, Color: saveColor},
	}
	switch focus {
	case uikit.ProjectFormFieldNamespace:
		actions = append(actions, components.Action{Hint: uikit.KeyNamespacePick.Hint, Color: s.palette.Hint})
	case uikit.ProjectFormFieldSource:
		actions = append(actions, components.Action{Hint: uikit.KeySourcePick.Hint, Color: s.palette.Hint})
	case uikit.ProjectFormFieldStack:
		actions = append(actions, components.Action{Hint: uikit.KeyStackPick.Hint, Color: s.palette.Hint})
	}

	return components.RenderActions(s.palette, actions)
}

func (s ProjectsScreen) namespaceName(namespaces []uikit.Namespace, id int64) string {
	for _, namespace := range namespaces {
		if namespace.ID == id {
			return namespace.Name
		}
	}
	if len(namespaces) == 0 {
		return "No namespaces"
	}
	return ""
}

func (s ProjectsScreen) sourceName(sources []uikit.Source, id int64) string {
	for _, source := range sources {
		if source.ID == id {
			return source.Name
		}
	}
	if len(sources) == 0 {
		return "No sources"
	}
	return ""
}

func (s ProjectsScreen) stackName(stacks []uikit.Stack, id int64) string {
	for _, stack := range stacks {
		if stack.ID == id {
			return stack.Name
		}
	}
	if len(stacks) == 0 {
		return "No stacks"
	}
	return ""
}

func (s ProjectsScreen) renderDeleteConfirm(lines []string, width, height int, confirm uikit.DeleteConfirm) []string {
	modal := components.NewModal(s.palette, s.palette.Primary)
	return modal.Overlay(lines, width, height, s.deleteConfirmModal(width, confirm))
}

func (s ProjectsScreen) deleteConfirmModal(width int, confirm uikit.DeleteConfirm) string {
	modalWidth := uikit.Min(uikit.Max(width-6, 34), 58)
	contentWidth := uikit.Max(modalWidth-2, 1)
	modal := components.NewModal(s.palette, s.palette.Primary)
	question := "Delete project " + confirm.Name + "?"
	title := modal.Title(uikit.SymbolError + " Delete project")

	rows := []string{
		modal.CenterLine(contentWidth, modal.Text(s.palette.Hint, question)),
		modal.CenterLine(contentWidth, modal.Text(s.palette.Error, confirm.Error)),
		modal.CenterLine(contentWidth, s.deleteActionsLine()),
	}

	return modal.Render(contentWidth, title, rows)
}

func (s ProjectsScreen) deleteActionsLine() string {
	return components.RenderActions(s.palette, []components.Action{
		{Hint: uikit.KeyConfirmNo.Hint, Color: s.palette.Primary},
		{Hint: uikit.KeyConfirmYes.Hint, Color: s.palette.Error},
	})
}
