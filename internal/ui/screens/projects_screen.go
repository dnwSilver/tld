package screens

import (
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
	"github.com/dnwSilver/tld/internal/ui/components"
	"github.com/dnwSilver/tld/internal/ui/uikit"
)

type ProjectsScreen struct {
	palette uikit.Palette
}

const (
	projectFormLabelWidth     = 10
	projectFormValueWidth     = 24
	projectNameColumnMinWidth = 8
	projectNamespaceColumnMin = 8
	projectSourceColumnMin    = 6
	projectTailColumnWidth    = 5
	projectIconSlotWidth      = 2
	projectIconColumnWidth    = projectIconSlotWidth * 2
	projectDepVersionWidth    = 12
)

type projectTableColumns struct {
	name      int
	namespace int
	source    int
}

func projectTableColumnsFor(width, iconColumnWidth int, projects []uikit.Project) projectTableColumns {
	fixed := iconColumnWidth + projectTailColumnWidth*4
	remaining := uikit.Max(width-fixed, projectNameColumnMinWidth+projectNamespaceColumnMin+projectSourceColumnMin)

	nameNeed := uikit.Max(components.ColumnWidth(projects, func(project uikit.Project) string { return project.Name }, projectNameColumnMinWidth), projectNameColumnMinWidth)
	nsNeed := uikit.Max(components.ColumnWidth(projects, func(project uikit.Project) string { return project.NamespaceName }, projectNamespaceColumnMin), projectNamespaceColumnMin)
	srcNeed := uikit.Max(components.ColumnWidth(projects, func(project uikit.Project) string { return project.SourceName }, projectSourceColumnMin), projectSourceColumnMin)
	totalNeed := nameNeed + nsNeed + srcNeed

	cols := projectTableColumns{}
	if totalNeed <= remaining {
		cols.name = nameNeed
		cols.namespace = nsNeed
		cols.source = srcNeed
	} else {
		cols.name = uikit.Max(remaining*2/5, projectNameColumnMinWidth)
		cols.namespace = uikit.Max(remaining*2/5, projectNamespaceColumnMin)
		cols.source = uikit.Max(remaining-cols.name-cols.namespace, projectSourceColumnMin)
	}

	total := iconColumnWidth + cols.name + cols.namespace + cols.source + projectTailColumnWidth*4
	if total > width {
		overflow := total - width
		if cols.source > projectSourceColumnMin {
			shrink := uikit.Min(overflow, cols.source-projectSourceColumnMin)
			cols.source -= shrink
			overflow -= shrink
		}
		if overflow > 0 && cols.namespace > projectNamespaceColumnMin {
			shrink := uikit.Min(overflow, cols.namespace-projectNamespaceColumnMin)
			cols.namespace -= shrink
			overflow -= shrink
		}
		if overflow > 0 && cols.name > projectNameColumnMinWidth {
			shrink := uikit.Min(overflow, cols.name-projectNameColumnMinWidth)
			cols.name -= shrink
		}
	}

	return cols
}

func alignColumn(value string, width int) string {
	for lipgloss.Width(value) > width {
		_, size := utf8.DecodeRuneInString(value)
		if size == 0 {
			break
		}
		value = value[size:]
	}

	return rightAligned(value, width)
}

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
	searchQuery string,
) string {
	boxHeight := uikit.Max(height-1, 3)
	contentHeight := uikit.Max(boxHeight-2, 1)
	modalOpen := form.Open || deleteConfirm.Open
	leftTotalWidth := uikit.Max(width*3/5, 2)
	rightTotalWidth := uikit.Max(width-leftTotalWidth, 2)
	leftContentWidth := uikit.Max(leftTotalWidth-2, 1)
	rightContentWidth := uikit.Max(rightTotalWidth-2, 1)

	count := len(projects)
	leftTitle := components.ScreenTitle(s.palette, uikit.SymbolProject, "Projects", &count, searchQuery)
	leftBorder := s.borderForPane(!modalOpen && focus == uikit.ProjectPaneProjects)
	if modalOpen {
		leftBorder = s.palette.Hint
	}
	left := components.NewBox(s.palette, leftBorder).Render(leftContentWidth, contentHeight, leftTitle, s.renderProjectsContent(leftContentWidth, contentHeight, projects, selectedProjectID))
	rightTitle := components.ScreenTitle(s.palette, uikit.SymbolDependency, "Dependencies", intPtr(len(projectDependencies)), "")
	rightBorder := s.borderForPane(!modalOpen && focus == uikit.ProjectPaneDependencies)
	right := components.NewBox(s.palette, rightBorder).Render(rightContentWidth, contentHeight, rightTitle, s.renderDependencyPanel(rightContentWidth, contentHeight, projectDependencies, selectedProjectDepID, projectLatestRun))
	body := lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	if width < narrowLayoutWidth {
		paneContentWidth := uikit.Max(width-2, 1)
		if focus == uikit.ProjectPaneDependencies && !modalOpen {
			right = components.NewBox(s.palette, s.palette.Primary).Render(paneContentWidth, contentHeight, rightTitle, s.renderDependencyPanel(paneContentWidth, contentHeight, projectDependencies, selectedProjectDepID, projectLatestRun))
			body = right
		} else {
			left = components.NewBox(s.palette, leftBorder).Render(paneContentWidth, contentHeight, leftTitle, s.renderProjectsContent(paneContentWidth, contentHeight, projects, selectedProjectID))
			body = left
		}
	}
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
	visibleProjects := visibleRowsByID(projects, selectedProjectID, height-1, func(project uikit.Project) int64 { return project.ID })
	iconColumnWidth := components.ColumnWidth(visibleProjects, projectIcons, projectIconColumnWidth)
	columns := projectTableColumnsFor(width, iconColumnWidth, visibleProjects)
	lines = append(lines, s.tableHeader(width, iconColumnWidth, columns))

	if len(projects) == 0 {
		empty := uikit.Text(s.palette, s.palette.Hint, "No projects yet")
		lines = append(lines, uikit.CenterLine(s.palette, width, empty))
	} else {
		for _, project := range visibleProjects {
			lines = append(lines, s.renderProjectRow(width, project, project.ID == selectedProjectID, iconColumnWidth, columns))
		}
	}

	return fillLines(s.palette, lines, width, height)
}

func (s ProjectsScreen) tableHeader(width int, iconColumnWidth int, columns projectTableColumns) string {
	return components.RenderTableRow(s.palette, s.palette.Background, width, []components.TableCell{
		{Value: "", Width: iconColumnWidth, Foreground: s.palette.Hint},
		{Value: "name", Width: columns.name, Foreground: s.palette.Hint, Bold: true},
		{Value: "namespace", Width: columns.namespace, Foreground: s.palette.Hint, Bold: true},
		{Value: "source", Width: columns.source, Foreground: s.palette.Hint, Bold: true},
		{Value: alignColumn("deps", projectTailColumnWidth), Width: projectTailColumnWidth, Foreground: s.palette.Hint, Bold: true},
		{Value: alignColumn("pid", projectTailColumnWidth), Width: projectTailColumnWidth, Foreground: s.palette.Hint, Bold: true},
		{Value: alignColumn("frz", projectTailColumnWidth), Width: projectTailColumnWidth, Foreground: s.palette.Hint, Bold: true},
		{Value: alignColumn("eol", projectTailColumnWidth), Width: projectTailColumnWidth, Foreground: s.palette.Hint, Bold: true},
	})
}

func (s ProjectsScreen) renderProjectRow(width int, project uikit.Project, selected bool, iconColumnWidth int, columns projectTableColumns) string {
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
		{Value: uikit.RenderProjectName(s.palette, project.Name, project.Freezing, project.EndOfLife, background, s.palette.Text, columns.name), Width: columns.name, Foreground: s.palette.Text},
		{Value: project.NamespaceName, Width: columns.namespace, Foreground: s.palette.Hint},
		{Value: project.SourceName, Width: columns.source, Foreground: s.palette.Info},
		{Value: alignColumn(uikit.FormatInt(project.DependencyCount), projectTailColumnWidth), Width: projectTailColumnWidth, Foreground: s.palette.Primary},
		{Value: alignColumn(project.ProjectID, projectTailColumnWidth), Width: projectTailColumnWidth, Foreground: s.palette.Info},
		{Value: s.projectFlag(project.Freezing, uikit.SymbolProjectFreeze, s.palette.Info, background), Width: projectTailColumnWidth, Foreground: s.palette.Info},
		{Value: s.projectFlag(project.EndOfLife, uikit.SymbolProjectEOL, s.palette.Hint, background), Width: projectTailColumnWidth, Foreground: s.palette.Hint},
	})
}

func (s ProjectsScreen) projectFlag(active bool, symbol string, color lipgloss.Color, background lipgloss.Color) string {
	value := ""
	if active {
		value = symbol
	}

	return lipgloss.NewStyle().
		Background(background).
		Foreground(color).
		Render(alignColumn(value, projectTailColumnWidth))
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
	if status.Running {
		message += " [Esc] cancel"
	}
	return components.RenderProgress(s.palette, width, uikit.SymbolDependency, message, status.Error, status.Running, status.Current, status.Total)
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
		modal.CenterLine(contentWidth, s.inputLine("Freezing", s.boolLabel(form.Freezing), form.Focus == uikit.ProjectFormFieldFreezing)),
		modal.CenterLine(contentWidth, s.inputLine("EOL", s.boolLabel(form.EndOfLife), form.Focus == uikit.ProjectFormFieldEndOfLife)),
		modal.CenterLine(contentWidth, modal.Text(s.palette.Error, form.Error)),
		modal.CenterLine(contentWidth, s.actionsLine(form.CanSave, form.Focus)),
	}

	return modal.Render(contentWidth, title, rows)
}

func (s ProjectsScreen) boolLabel(value bool) string {
	if value {
		return "yes"
	}

	return "no"
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
