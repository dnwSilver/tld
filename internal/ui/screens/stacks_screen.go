package screens

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dnwSilver/tld/internal/ui/components"
	"github.com/dnwSilver/tld/internal/ui/uikit"
)

type StacksScreen struct {
	palette     uikit.Palette
	screenTitle string
	symbol      string
}

const (
	stackFormLabelWidth = 6
	stackFormValueWidth = 24
)

func NewStacksScreen(palette uikit.Palette) StacksScreen {
	return StacksScreen{
		palette:     palette,
		screenTitle: "Stacks",
		symbol:      uikit.SymbolStack,
	}
}

func NewNamespacesScreen(palette uikit.Palette) StacksScreen {
	return StacksScreen{
		palette:     palette,
		screenTitle: "Namespaces",
		symbol:      uikit.SymbolNamespace,
	}
}

func (s StacksScreen) Render(
	width int,
	height int,
	stacks []uikit.Stack,
	selectedStackID int64,
	policies []uikit.Policy,
	form uikit.StackForm,
	deleteConfirm uikit.DeleteConfirm,
	searchQuery string,
) string {
	boxWidth := uikit.Max(width, 2)
	boxHeight := uikit.Max(height-1, 3)
	contentWidth := uikit.Max(boxWidth-2, 1)
	contentHeight := uikit.Max(boxHeight-2, 1)

	count := len(stacks)
	title := components.ScreenTitle(s.palette, s.symbol, s.screenTitle, &count, searchQuery)
	content := s.renderContent(contentWidth, contentHeight, stacks, selectedStackID, policies, form, deleteConfirm)
	borderColor := s.palette.Primary
	if form.Open || deleteConfirm.Open {
		borderColor = s.palette.Hint
	}
	box := components.NewBox(s.palette, borderColor).Render(contentWidth, contentHeight, title, content)
	footer := components.ScreenFooter(s.palette, width, "Kolosov Aleksandr")

	return lipgloss.JoinVertical(lipgloss.Left, box, footer)
}

func (s StacksScreen) renderContent(
	width int,
	height int,
	stacks []uikit.Stack,
	selectedStackID int64,
	policies []uikit.Policy,
	form uikit.StackForm,
	deleteConfirm uikit.DeleteConfirm,
) string {
	lines := make([]string, 0, height)
	visibleStacks := visibleRowsByID(stacks, selectedStackID, height-1, func(stack uikit.Stack) int64 { return stack.ID })
	iconColumnWidth := components.ColumnWidth(visibleStacks, func(stack uikit.Stack) string { return stack.Icon }, 2)
	lines = append(lines, s.tableHeader(width, iconColumnWidth))

	if len(stacks) == 0 {
		empty := uikit.Text(s.palette, s.palette.Hint, "No "+strings.ToLower(s.screenTitle)+" yet")
		lines = append(lines, uikit.CenterLine(s.palette, width, empty))
	} else {
		for _, stack := range visibleStacks {
			lines = append(lines, s.renderStackRow(width, stack, stack.ID == selectedStackID, iconColumnWidth))
		}
	}

	for len(lines) < height {
		lines = append(lines, uikit.BackgroundSpaces(s.palette, width))
	}

	if len(lines) > height {
		lines = lines[:height]
	}

	if form.Open {
		lines = s.renderModal(lines, width, height, policies, form)
	}
	if deleteConfirm.Open {
		lines = s.renderDeleteConfirm(lines, width, height, deleteConfirm)
	}

	return strings.Join(lines, uikit.SymbolLineBreak)
}

func (s StacksScreen) tableHeader(width int, iconColumnWidth int) string {
	cells := []components.TableCell{
		{Value: "", Width: iconColumnWidth, Foreground: s.palette.Hint},
		{Value: "name", Foreground: s.palette.Hint, Bold: true},
	}
	if s.screenTitle == "Namespaces" {
		cells[1].Width = 24
		cells = append(cells, components.TableCell{Value: "policy", Foreground: s.palette.Hint, Bold: true})
	}

	return components.RenderTableRow(s.palette, s.palette.Background, width, cells)
}

func (s StacksScreen) renderStackRow(width int, stack uikit.Stack, selected bool, iconColumnWidth int) string {
	background := s.palette.Background
	if selected {
		background = s.palette.Hover
	}
	iconColor := lipgloss.Color(uikit.NormalizeHexColor(stack.Color))

	cells := []components.TableCell{
		{Value: stack.Icon, Width: iconColumnWidth, Foreground: iconColor},
		{Value: stack.Name, Foreground: s.palette.Text},
	}
	if s.screenTitle == "Namespaces" {
		cells[1].Width = 24
		cells = append(cells, components.TableCell{Value: stack.PolicyName, Foreground: s.palette.Hint})
	}

	return components.RenderTableRow(s.palette, background, width, cells)
}

func (s StacksScreen) renderModal(lines []string, width, height int, policies []uikit.Policy, form uikit.StackForm) []string {
	modal := components.NewModal(s.palette, s.palette.Primary)
	return modal.Overlay(lines, width, height, s.modal(width, policies, form))
}

func (s StacksScreen) modal(width int, policies []uikit.Policy, form uikit.StackForm) string {
	modalWidth := uikit.Min(uikit.Max(width-6, 32), 54)
	contentWidth := uikit.Max(modalWidth-2, 1)
	modal := components.NewModal(s.palette, s.palette.Primary)

	action := "Add"
	if form.Mode == uikit.StackFormModeEdit {
		action = "Edit"
	}
	title := modal.Title(s.symbol + " " + action + " " + s.singularTitle())

	rows := []string{
		modal.CenterLine(contentWidth, s.inputLine("Icon", form.Icon, form.Focus == uikit.StackFormFieldIcon)),
		modal.CenterLine(contentWidth, s.inputLine("Color", form.Color, form.Focus == uikit.StackFormFieldColor)),
		modal.CenterLine(contentWidth, s.inputLine("Name", form.Name, form.Focus == uikit.StackFormFieldName)),
	}
	if s.screenTitle == "Namespaces" {
		rows = append(rows, modal.CenterLine(contentWidth, s.inputLine("Policy", form.Policy, form.Focus == uikit.StackFormFieldPolicy)))
	}
	rows = append(rows,
		modal.CenterLine(contentWidth, modal.Text(s.palette.Error, form.Error)),
		modal.CenterLine(contentWidth, s.actionsLine(form.CanSave, s.screenTitle == "Namespaces" && form.Focus == uikit.StackFormFieldPolicy)),
	)

	return modal.Render(contentWidth, title, rows)
}

func (s StacksScreen) inputLine(label, value string, focused bool) string {
	return components.RenderFormField(s.palette, components.FormField{
		Label:      label,
		Value:      value,
		Focused:    focused,
		LabelWidth: stackFormLabelWidth,
		ValueWidth: stackFormValueWidth,
	})
}

func (s StacksScreen) actionsLine(canSave bool, policyFocused bool) string {
	saveColor := s.palette.Disable
	if canSave {
		saveColor = s.palette.Primary
	}

	actions := []components.Action{
		{Hint: uikit.KeyCancel.Hint, Color: s.palette.Primary},
		{Hint: uikit.KeySave.Hint, Color: saveColor},
	}
	if policyFocused {
		actions = append(actions, components.Action{Hint: uikit.KeyPolicyPick.Hint, Color: s.palette.Hint})
	}

	return components.RenderActions(s.palette, actions)
}

func (s StacksScreen) renderDeleteConfirm(lines []string, width, height int, confirm uikit.DeleteConfirm) []string {
	modal := components.NewModal(s.palette, s.palette.Primary)
	return modal.Overlay(lines, width, height, s.deleteConfirmModal(width, confirm))
}

func (s StacksScreen) deleteConfirmModal(width int, confirm uikit.DeleteConfirm) string {
	modalWidth := uikit.Min(uikit.Max(width-6, 34), 58)
	contentWidth := uikit.Max(modalWidth-2, 1)
	modal := components.NewModal(s.palette, s.palette.Primary)
	question := "Delete " + s.singularTitle() + " " + confirm.Name + "?"
	title := modal.Title(uikit.SymbolError + " Delete " + s.singularTitle())

	rows := []string{
		modal.CenterLine(contentWidth, modal.Text(s.palette.Hint, question)),
		modal.CenterLine(contentWidth, modal.Text(s.palette.Error, confirm.Error)),
		modal.CenterLine(contentWidth, s.deleteActionsLine()),
	}

	return modal.Render(contentWidth, title, rows)
}

func (s StacksScreen) deleteActionsLine() string {
	return components.RenderActions(s.palette, []components.Action{
		{Hint: uikit.KeyConfirmNo.Hint, Color: s.palette.Primary},
		{Hint: uikit.KeyConfirmYes.Hint, Color: s.palette.Error},
	})
}

func (s StacksScreen) singularTitle() string {
	if s.screenTitle == "Namespaces" {
		return "namespace"
	}

	return "stack"
}
