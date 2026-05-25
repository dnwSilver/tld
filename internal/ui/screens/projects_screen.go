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
	projectNameColumnWidth      = 18
	projectNamespaceColumnWidth = 16
	projectSourceColumnWidth    = 16
	projectStackColumnWidth     = 14
)

func NewProjectsScreen(palette uikit.Palette) ProjectsScreen {
	return ProjectsScreen{palette: palette}
}

func (s ProjectsScreen) Render(
	width int,
	height int,
	projects []uikit.Project,
	selectedProjectID int64,
	namespaces []uikit.Namespace,
	sources []uikit.Source,
	stacks []uikit.Stack,
	form uikit.ProjectForm,
	deleteConfirm uikit.DeleteConfirm,
) string {
	boxWidth := uikit.Max(width, 2)
	boxHeight := uikit.Max(height-1, 3)
	contentWidth := uikit.Max(boxWidth-2, 1)
	contentHeight := uikit.Max(boxHeight-2, 1)

	count := len(projects)
	title := components.ScreenTitle(s.palette, uikit.SymbolProject, "Projects", &count)
	content := s.renderContent(contentWidth, contentHeight, projects, selectedProjectID, namespaces, sources, stacks, form, deleteConfirm)
	borderColor := s.palette.Primary
	if form.Open || deleteConfirm.Open {
		borderColor = s.palette.Hint
	}
	box := components.NewBox(s.palette, borderColor).Render(contentWidth, contentHeight, title, content)
	footer := components.ScreenFooter(s.palette, width, "Kolosov Aleksandr")

	return lipgloss.JoinVertical(lipgloss.Left, box, footer)
}

func (s ProjectsScreen) renderContent(
	width int,
	height int,
	projects []uikit.Project,
	selectedProjectID int64,
	namespaces []uikit.Namespace,
	sources []uikit.Source,
	stacks []uikit.Stack,
	form uikit.ProjectForm,
	deleteConfirm uikit.DeleteConfirm,
) string {
	lines := make([]string, 0, height)
	iconColumnWidth := components.ColumnWidth(projects, func(project uikit.Project) string { return project.Icon }, 2)
	lines = append(lines, s.tableHeader(width, iconColumnWidth))

	if len(projects) == 0 {
		empty := uikit.Text(s.palette, s.palette.Hint, "No projects yet")
		lines = append(lines, uikit.CenterLine(s.palette, width, empty))
	} else {
		for _, project := range projects {
			lines = append(lines, s.renderProjectRow(width, project, project.ID == selectedProjectID, iconColumnWidth))
		}
	}

	for len(lines) < height {
		lines = append(lines, uikit.BackgroundSpaces(s.palette, width))
	}
	if len(lines) > height {
		lines = lines[:height]
	}

	if form.Open {
		lines = s.renderModal(lines, width, height, namespaces, sources, stacks, form)
	}
	if deleteConfirm.Open {
		lines = s.renderDeleteConfirm(lines, width, height, deleteConfirm)
	}

	return strings.Join(lines, uikit.SymbolLineBreak)
}

func (s ProjectsScreen) tableHeader(width int, iconColumnWidth int) string {
	return components.RenderTableRow(s.palette, s.palette.Background, width, []components.TableCell{
		{Value: "", Width: iconColumnWidth, Foreground: s.palette.Hint},
		{Value: "name", Width: projectNameColumnWidth, Foreground: s.palette.Hint, Bold: true},
		{Value: "namespace", Width: projectNamespaceColumnWidth, Foreground: s.palette.Hint, Bold: true},
		{Value: "source", Width: projectSourceColumnWidth, Foreground: s.palette.Hint, Bold: true},
		{Value: "stack", Width: projectStackColumnWidth, Foreground: s.palette.Hint, Bold: true},
		{Value: "project_id", Width: projectIDColumnWidth, Foreground: s.palette.Hint, Bold: true},
	})
}

func (s ProjectsScreen) renderProjectRow(width int, project uikit.Project, selected bool, iconColumnWidth int) string {
	background := s.palette.Background
	if selected {
		background = s.palette.Hover
	}
	iconColor := lipgloss.Color(uikit.NormalizeHexColor(project.Color))

	return components.RenderTableRow(s.palette, background, width, []components.TableCell{
		{Value: project.Icon, Width: iconColumnWidth, Foreground: iconColor},
		{Value: project.Name, Width: projectNameColumnWidth, Foreground: s.palette.Text},
		{Value: project.NamespaceName, Width: projectNamespaceColumnWidth, Foreground: s.palette.Hint},
		{Value: project.SourceName, Width: projectSourceColumnWidth, Foreground: s.palette.Info},
		{Value: project.StackName, Width: projectStackColumnWidth, Foreground: s.palette.Hint},
		{Value: project.ProjectID, Width: projectIDColumnWidth, Foreground: s.palette.Info},
	})
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
