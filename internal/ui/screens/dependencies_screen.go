package screens

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dnwSilver/tld/internal/ui/components"
	"github.com/dnwSilver/tld/internal/ui/uikit"
)

type DependenciesScreen struct {
	palette uikit.Palette
}

const (
	dependencyFormLabelWidth  = 5
	dependencyFormValueWidth  = 24
	dependencyNameColumnWidth = 24
)

func NewDependenciesScreen(palette uikit.Palette) DependenciesScreen {
	return DependenciesScreen{
		palette: palette,
	}
}

func (s DependenciesScreen) Render(
	width int,
	height int,
	dependencies []uikit.Dependency,
	selectedDependencyID int64,
	stacks []uikit.Stack,
	form uikit.DependencyForm,
	deleteConfirm uikit.DeleteConfirm,
) string {
	boxWidth := uikit.Max(width, 2)
	boxHeight := uikit.Max(height-1, 3)
	contentWidth := uikit.Max(boxWidth-2, 1)
	contentHeight := uikit.Max(boxHeight-2, 1)

	count := len(dependencies)
	title := components.ScreenTitle(s.palette, uikit.SymbolDependency, "Dependencies", &count)
	content := s.renderContent(contentWidth, contentHeight, dependencies, selectedDependencyID, stacks, form, deleteConfirm)
	borderColor := s.palette.Primary
	if form.Open || deleteConfirm.Open {
		borderColor = s.palette.Hint
	}
	box := components.NewBox(s.palette, borderColor).Render(contentWidth, contentHeight, title, content)
	footer := components.ScreenFooter(s.palette, width, "Kolosov Aleksandr")

	return lipgloss.JoinVertical(lipgloss.Left, box, footer)
}

func (s DependenciesScreen) renderContent(
	width int,
	height int,
	dependencies []uikit.Dependency,
	selectedDependencyID int64,
	stacks []uikit.Stack,
	form uikit.DependencyForm,
	deleteConfirm uikit.DeleteConfirm,
) string {
	lines := make([]string, 0, height)
	iconColumnWidth := components.ColumnWidth(dependencies, func(d uikit.Dependency) string { return d.Icon }, 2)
	stackColumnWidth := components.ColumnWidth(dependencies, func(d uikit.Dependency) string { return d.StackName }, 5)
	lines = append(lines, s.tableHeader(width, iconColumnWidth, stackColumnWidth))

	if len(dependencies) == 0 {
		empty := uikit.Text(s.palette, s.palette.Hint, "No dependencies yet")
		lines = append(lines, uikit.CenterLine(s.palette, width, empty))
	} else {
		for _, dependency := range dependencies {
			lines = append(lines, s.renderDependencyRow(width, dependency, dependency.ID == selectedDependencyID, iconColumnWidth, stackColumnWidth))
		}
	}

	for len(lines) < height {
		lines = append(lines, uikit.BackgroundSpaces(s.palette, width))
	}

	if len(lines) > height {
		lines = lines[:height]
	}

	if form.Open {
		lines = s.renderModal(lines, width, height, stacks, form)
	}
	if deleteConfirm.Open {
		lines = s.renderDeleteConfirm(lines, width, height, deleteConfirm)
	}

	return strings.Join(lines, uikit.SymbolLineBreak)
}

func (s DependenciesScreen) tableHeader(width int, iconColumnWidth int, stackColumnWidth int) string {
	return components.RenderTableRow(s.palette, s.palette.Background, width, []components.TableCell{
		{Value: "", Width: iconColumnWidth, Foreground: s.palette.Hint},
		{Value: "name", Width: dependencyNameColumnWidth, Foreground: s.palette.Hint, Bold: true},
		{Value: "stack", Width: stackColumnWidth, Foreground: s.palette.Hint, Bold: true},
	})
}

func (s DependenciesScreen) renderDependencyRow(width int, dependency uikit.Dependency, selected bool, iconColumnWidth int, stackColumnWidth int) string {
	background := s.palette.Background
	if selected {
		background = s.palette.Hover
	}
	iconColor := lipgloss.Color(uikit.NormalizeHexColor(dependency.Color))

	return components.RenderTableRow(s.palette, background, width, []components.TableCell{
		{Value: dependency.Icon, Width: iconColumnWidth, Foreground: iconColor},
		{Value: dependency.Name, Width: dependencyNameColumnWidth, Foreground: s.palette.Text},
		{Value: dependency.StackName, Width: stackColumnWidth, Foreground: s.palette.Hint},
	})
}

func (s DependenciesScreen) renderModal(lines []string, width, height int, stacks []uikit.Stack, form uikit.DependencyForm) []string {
	modal := components.NewModal(s.palette, s.palette.Primary)
	return modal.Overlay(lines, width, height, s.modal(width, stacks, form))
}

func (s DependenciesScreen) modal(width int, stacks []uikit.Stack, form uikit.DependencyForm) string {
	modalWidth := uikit.Min(uikit.Max(width-6, 34), 58)
	contentWidth := uikit.Max(modalWidth-2, 1)
	modal := components.NewModal(s.palette, s.palette.Primary)

	action := "Add"
	if form.Mode == uikit.StackFormModeEdit {
		action = "Edit"
	}
	title := modal.Title(uikit.SymbolDependency + " " + action + " dependency")

	rows := []string{
		modal.CenterLine(contentWidth, s.inputLine("Icon", form.Icon, form.Focus == uikit.DependencyFormFieldIcon)),
		modal.CenterLine(contentWidth, s.inputLine("Color", form.Color, form.Focus == uikit.DependencyFormFieldColor)),
		modal.CenterLine(contentWidth, s.inputLine("Name", form.Name, form.Focus == uikit.DependencyFormFieldName)),
		modal.CenterLine(contentWidth, s.inputLine("Stack", s.stackName(stacks, form.StackID), form.Focus == uikit.DependencyFormFieldStack)),
		modal.CenterLine(contentWidth, modal.Text(s.palette.Error, form.Error)),
		modal.CenterLine(contentWidth, s.actionsLine(form.CanSave, form.Focus == uikit.DependencyFormFieldStack)),
	}

	return modal.Render(contentWidth, title, rows)
}

func (s DependenciesScreen) stackName(stacks []uikit.Stack, stackID int64) string {
	for _, stack := range stacks {
		if stack.ID == stackID {
			return stack.Name
		}
	}
	if len(stacks) == 0 {
		return "No stacks"
	}

	return ""
}

func (s DependenciesScreen) inputLine(label, value string, focused bool) string {
	return components.RenderFormField(s.palette, components.FormField{
		Label:      label,
		Value:      value,
		Focused:    focused,
		LabelWidth: dependencyFormLabelWidth,
		ValueWidth: dependencyFormValueWidth,
	})
}

func (s DependenciesScreen) actionsLine(canSave bool, stackFocused bool) string {
	saveColor := s.palette.Disable
	if canSave {
		saveColor = s.palette.Primary
	}

	actions := []components.Action{
		{Hint: uikit.KeyCancel.Hint, Color: s.palette.Primary},
		{Hint: uikit.KeySave.Hint, Color: saveColor},
	}
	if stackFocused {
		actions = append(actions, components.Action{Hint: uikit.KeyStackPick.Hint, Color: s.palette.Hint})
	}

	return components.RenderActions(s.palette, actions)
}

func (s DependenciesScreen) renderDeleteConfirm(lines []string, width, height int, confirm uikit.DeleteConfirm) []string {
	modal := components.NewModal(s.palette, s.palette.Primary)
	return modal.Overlay(lines, width, height, s.deleteConfirmModal(width, confirm))
}

func (s DependenciesScreen) deleteConfirmModal(width int, confirm uikit.DeleteConfirm) string {
	modalWidth := uikit.Min(uikit.Max(width-6, 34), 58)
	contentWidth := uikit.Max(modalWidth-2, 1)
	modal := components.NewModal(s.palette, s.palette.Primary)
	question := "Delete dependency " + confirm.Name + "?"
	title := modal.Title(uikit.SymbolError + " Delete dependency")

	rows := []string{
		modal.CenterLine(contentWidth, modal.Text(s.palette.Hint, question)),
		modal.CenterLine(contentWidth, modal.Text(s.palette.Error, confirm.Error)),
		modal.CenterLine(contentWidth, s.deleteActionsLine()),
	}

	return modal.Render(contentWidth, title, rows)
}

func (s DependenciesScreen) deleteActionsLine() string {
	return components.RenderActions(s.palette, []components.Action{
		{Hint: uikit.KeyConfirmNo.Hint, Color: s.palette.Primary},
		{Hint: uikit.KeyConfirmYes.Hint, Color: s.palette.Error},
	})
}
