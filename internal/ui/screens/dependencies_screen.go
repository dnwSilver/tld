package screens

import (
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dnwSilver/tld/internal/ui/components"
	"github.com/dnwSilver/tld/internal/ui/uikit"
)

type DependenciesScreen struct {
	palette uikit.Palette
}

const (
	dependencyFormLabelWidth = 5
	dependencyFormValueWidth = 24
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

	content := s.renderContent(contentWidth, contentHeight, dependencies, selectedDependencyID, stacks, form, deleteConfirm)
	box := s.renderBox(contentWidth, contentHeight, s.title(len(dependencies)), content)

	footer := lipgloss.NewStyle().
		Width(width).
		Background(s.palette.Background).
		Foreground(s.palette.Hint).
		Align(lipgloss.Right).
		Render(uikit.SymbolCopyright + " Kolosov Aleksandr")

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
	iconColumnWidth := s.iconColumnWidth(dependencies)
	stackColumnWidth := s.stackColumnWidth(dependencies)
	lines = append(lines, s.tableHeader(width, iconColumnWidth, stackColumnWidth))

	if len(dependencies) == 0 {
		empty := lipgloss.NewStyle().
			Background(s.palette.Background).
			Foreground(s.palette.Hint).
			Render("No dependencies yet")
		lines = append(lines, s.centerLine(width, empty))
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

func (s DependenciesScreen) title(dependencyCount int) string {
	return lipgloss.NewStyle().
		Background(s.palette.Background).
		Foreground(s.palette.Warning).
		Bold(true).
		Render(uikit.SymbolDependency + " Dependencies [" + uikit.FormatInt(dependencyCount) + "]")
}

func (s DependenciesScreen) renderBox(width, height int, title, content string) string {
	contentLines := strings.Split(content, uikit.SymbolLineBreak)
	for len(contentLines) < height {
		contentLines = append(contentLines, uikit.BackgroundSpaces(s.palette, width))
	}
	if len(contentLines) > height {
		contentLines = contentLines[:height]
	}

	lines := make([]string, 0, height+2)
	lines = append(lines, s.topBorder(width, title))
	for _, line := range contentLines {
		lines = append(lines, s.borderPart("│")+line+s.borderPart("│"))
	}
	lines = append(lines, s.borderPart("└")+s.borderFill(width)+s.borderPart("┘"))

	return strings.Join(lines, uikit.SymbolLineBreak)
}

func (s DependenciesScreen) topBorder(width int, title string) string {
	titleWidth := lipgloss.Width(title)
	available := uikit.Max(width-titleWidth-2, 0)
	left := int(math.Floor(float64(available) / 2))
	right := uikit.Max(available-left, 0)

	return s.borderPart("┌") +
		s.borderFill(left) +
		s.borderPart(" ") +
		title +
		s.borderPart(" ") +
		s.borderFill(right) +
		s.borderPart("┐")
}

func (s DependenciesScreen) borderFill(width int) string {
	return s.borderPart(strings.Repeat("─", uikit.Max(width, 0)))
}

func (s DependenciesScreen) borderPart(value string) string {
	return lipgloss.NewStyle().
		Background(s.palette.Background).
		Foreground(s.palette.Primary).
		Render(value)
}

func (s DependenciesScreen) tableHeader(width int, iconColumnWidth int, stackColumnWidth int) string {
	iconColumn := lipgloss.NewStyle().
		Background(s.palette.Background).
		Foreground(s.palette.Hint).
		Width(iconColumnWidth).
		Render("")
	nameColumn := lipgloss.NewStyle().
		Background(s.palette.Background).
		Foreground(s.palette.Hint).
		Bold(true).
		Width(24).
		Render("name")
	stackColumn := lipgloss.NewStyle().
		Background(s.palette.Background).
		Foreground(s.palette.Hint).
		Bold(true).
		Width(stackColumnWidth).
		Render("stack")
	content := iconColumn + nameColumn + stackColumn

	return content + uikit.BackgroundSpaces(s.palette, width-lipgloss.Width(content))
}

func (s DependenciesScreen) iconColumnWidth(dependencies []uikit.Dependency) int {
	width := 2
	for _, dependency := range dependencies {
		width = uikit.Max(width, lipgloss.Width(dependency.Icon))
	}

	return width
}

func (s DependenciesScreen) stackColumnWidth(dependencies []uikit.Dependency) int {
	width := 5
	for _, dependency := range dependencies {
		width = uikit.Max(width, lipgloss.Width(dependency.StackName))
	}

	return width
}

func (s DependenciesScreen) renderDependencyRow(width int, dependency uikit.Dependency, selected bool, iconColumnWidth int, stackColumnWidth int) string {
	rowPalette := s.palette
	background := s.palette.Background
	foreground := s.palette.Text
	if selected {
		background = s.palette.Hover
	}

	color := lipgloss.Color(uikit.NormalizeHexColor(dependency.Color))
	iconColumn := lipgloss.NewStyle().
		Background(background).
		Foreground(color).
		Width(iconColumnWidth).
		Render(dependency.Icon)
	nameColumn := lipgloss.NewStyle().
		Background(background).
		Foreground(foreground).
		Width(24).
		Render(dependency.Name)
	stackColumn := lipgloss.NewStyle().
		Background(background).
		Foreground(s.palette.Hint).
		Width(stackColumnWidth).
		Render(dependency.StackName)
	content := iconColumn + nameColumn + stackColumn
	rowPalette.Background = background

	return content + uikit.BackgroundSpaces(rowPalette, width-lipgloss.Width(content))
}

func (s DependenciesScreen) centerLine(width int, content string) string {
	contentWidth := lipgloss.Width(content)
	left := uikit.Max((width-contentWidth)/2, 0)
	right := uikit.Max(width-left-contentWidth, 0)

	return uikit.BackgroundSpaces(s.palette, left) + content + uikit.BackgroundSpaces(s.palette, right)
}

func (s DependenciesScreen) renderModal(lines []string, width, height int, stacks []uikit.Stack, form uikit.DependencyForm) []string {
	modal := strings.Split(s.modal(width, stacks, form), uikit.SymbolLineBreak)
	modalHeight := len(modal)
	top := uikit.Max((height-modalHeight)/2, 0)

	for index, modalLine := range modal {
		row := top + index
		if row >= len(lines) {
			break
		}

		modalWidth := lipgloss.Width(modalLine)
		left := uikit.Max((width-modalWidth)/2, 0)
		right := uikit.Max(width-left-modalWidth, 0)
		lines[row] = uikit.BackgroundSpaces(s.palette, left) + modalLine + uikit.BackgroundSpaces(s.palette, right)
	}

	return lines
}

func (s DependenciesScreen) modal(width int, stacks []uikit.Stack, form uikit.DependencyForm) string {
	modalWidth := uikit.Min(uikit.Max(width-6, 34), 58)
	contentWidth := uikit.Max(modalWidth-2, 1)
	modal := components.NewModal(s.palette, s.palette.Primary)
	title := modal.Title(uikit.SymbolDependency + " Add dependency")
	if form.Mode == uikit.StackFormModeEdit {
		title = modal.Title(uikit.SymbolDependency + " Edit dependency")
	}
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
	prefix := "  "
	if focused {
		prefix = "> "
	}

	prefixPart := lipgloss.NewStyle().
		Background(s.palette.Background).
		Foreground(s.palette.Hint).
		Render(prefix)
	labelPart := lipgloss.NewStyle().
		Background(s.palette.Background).
		Foreground(s.palette.Hint).
		Width(dependencyFormLabelWidth).
		Align(lipgloss.Right).
		Render(label)
	separatorPart := lipgloss.NewStyle().
		Background(s.palette.Background).
		Foreground(s.palette.Hint).
		Render(": ")
	valuePart := lipgloss.NewStyle().
		Background(s.palette.Background).
		Foreground(s.palette.Primary).
		Render(value)
	valuePadding := uikit.BackgroundSpaces(s.palette, dependencyFormValueWidth-lipgloss.Width(valuePart))

	return prefixPart + labelPart + separatorPart + valuePart + valuePadding
}

func (s DependenciesScreen) actionsLine(canSave bool, stackFocused bool) string {
	saveColor := s.palette.Disable
	if canSave {
		saveColor = s.palette.Primary
	}

	modal := components.NewModal(s.palette, s.palette.Primary)
	cancel := modal.Text(s.palette.Warning, "[Esc] Cancel")
	save := modal.Text(saveColor, "[Enter] Save")
	stack := ""
	if stackFocused {
		stack = modal.Spaces(3) + modal.Text(s.palette.Hint, "[h/j] Stack")
	}

	return cancel + modal.Spaces(3) + save + stack
}

func (s DependenciesScreen) renderDeleteConfirm(lines []string, width, height int, confirm uikit.DeleteConfirm) []string {
	modal := strings.Split(s.deleteConfirmModal(width, confirm), uikit.SymbolLineBreak)
	modalHeight := len(modal)
	top := uikit.Max((height-modalHeight)/2, 0)

	for index, modalLine := range modal {
		row := top + index
		if row >= len(lines) {
			break
		}

		modalWidth := lipgloss.Width(modalLine)
		left := uikit.Max((width-modalWidth)/2, 0)
		right := uikit.Max(width-left-modalWidth, 0)
		lines[row] = uikit.BackgroundSpaces(s.palette, left) + modalLine + uikit.BackgroundSpaces(s.palette, right)
	}

	return lines
}

func (s DependenciesScreen) deleteConfirmModal(width int, confirm uikit.DeleteConfirm) string {
	modalWidth := uikit.Min(uikit.Max(width-6, 34), 58)
	contentWidth := uikit.Max(modalWidth-2, 1)
	modal := components.NewModal(s.palette, s.palette.Error)
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
	modal := components.NewModal(s.palette, s.palette.Error)
	no := modal.Text(s.palette.Primary, "[Esc] No")
	yes := modal.Text(s.palette.Error, "[Enter] Yes")

	return no + modal.Spaces(3) + yes
}
