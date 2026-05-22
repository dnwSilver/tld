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
	stackFormLabelWidth = 5
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
	form uikit.StackForm,
	deleteConfirm uikit.DeleteConfirm,
) string {
	boxWidth := uikit.Max(width, 2)
	boxHeight := uikit.Max(height-1, 3)
	contentWidth := uikit.Max(boxWidth-2, 1)
	contentHeight := uikit.Max(boxHeight-2, 1)

	count := len(stacks)
	title := components.ScreenTitle(s.palette, s.symbol, s.screenTitle, &count)
	content := s.renderContent(contentWidth, contentHeight, stacks, selectedStackID, form, deleteConfirm)
	box := components.NewBox(s.palette, s.palette.Hint).Render(contentWidth, contentHeight, title, content)
	footer := components.ScreenFooter(s.palette, width, "Kolosov Aleksandr")

	return lipgloss.JoinVertical(lipgloss.Left, box, footer)
}

func (s StacksScreen) renderContent(
	width int,
	height int,
	stacks []uikit.Stack,
	selectedStackID int64,
	form uikit.StackForm,
	deleteConfirm uikit.DeleteConfirm,
) string {
	lines := make([]string, 0, height)
	iconColumnWidth := components.ColumnWidth(stacks, func(stack uikit.Stack) string { return stack.Icon }, 2)
	lines = append(lines, s.tableHeader(width, iconColumnWidth))

	if len(stacks) == 0 {
		empty := uikit.Text(s.palette, s.palette.Hint, "No "+strings.ToLower(s.screenTitle)+" yet")
		lines = append(lines, uikit.CenterLine(s.palette, width, empty))
	} else {
		for _, stack := range stacks {
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
		lines = s.renderModal(lines, width, height, form)
	}
	if deleteConfirm.Open {
		lines = s.renderDeleteConfirm(lines, width, height, deleteConfirm)
	}

	return strings.Join(lines, uikit.SymbolLineBreak)
}

func (s StacksScreen) tableHeader(width int, iconColumnWidth int) string {
	return components.RenderTableRow(s.palette, s.palette.Background, width, []components.TableCell{
		{Value: "", Width: iconColumnWidth, Foreground: s.palette.Hint},
		{Value: "name", Foreground: s.palette.Hint, Bold: true},
	})
}

func (s StacksScreen) renderStackRow(width int, stack uikit.Stack, selected bool, iconColumnWidth int) string {
	background := s.palette.Background
	if selected {
		background = s.palette.Hover
	}
	iconColor := lipgloss.Color(uikit.NormalizeHexColor(stack.Color))

	return components.RenderTableRow(s.palette, background, width, []components.TableCell{
		{Value: stack.Icon, Width: iconColumnWidth, Foreground: iconColor},
		{Value: stack.Name, Foreground: s.palette.Text},
	})
}

func (s StacksScreen) renderModal(lines []string, width, height int, form uikit.StackForm) []string {
	modal := components.NewModal(s.palette, s.palette.Hint)
	return modal.Overlay(lines, width, height, s.modal(width, form))
}

func (s StacksScreen) modal(width int, form uikit.StackForm) string {
	modalWidth := uikit.Min(uikit.Max(width-6, 32), 54)
	contentWidth := uikit.Max(modalWidth-2, 1)
	modal := components.NewModal(s.palette, s.palette.Hint)

	action := "Add"
	if form.Mode == uikit.StackFormModeEdit {
		action = "Edit"
	}
	title := modal.Title(s.symbol + " " + action + " " + s.singularTitle())

	rows := []string{
		modal.CenterLine(contentWidth, s.inputLine("Icon", form.Icon, form.Focus == uikit.StackFormFieldIcon)),
		modal.CenterLine(contentWidth, s.inputLine("Color", form.Color, form.Focus == uikit.StackFormFieldColor)),
		modal.CenterLine(contentWidth, s.inputLine("Name", form.Name, form.Focus == uikit.StackFormFieldName)),
		modal.CenterLine(contentWidth, modal.Text(s.palette.Error, form.Error)),
		modal.CenterLine(contentWidth, s.actionsLine(form.CanSave)),
	}

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

func (s StacksScreen) actionsLine(canSave bool) string {
	saveColor := s.palette.Disable
	if canSave {
		saveColor = s.palette.Warning
	}

	return components.RenderActions(s.palette, []components.Action{
		{Hint: uikit.KeyCancel.Hint, Color: s.palette.Primary},
		{Hint: uikit.KeySave.Hint, Color: saveColor},
	})
}

func (s StacksScreen) renderDeleteConfirm(lines []string, width, height int, confirm uikit.DeleteConfirm) []string {
	modal := components.NewModal(s.palette, s.palette.Hint)
	return modal.Overlay(lines, width, height, s.deleteConfirmModal(width, confirm))
}

func (s StacksScreen) deleteConfirmModal(width int, confirm uikit.DeleteConfirm) string {
	modalWidth := uikit.Min(uikit.Max(width-6, 34), 58)
	contentWidth := uikit.Max(modalWidth-2, 1)
	modal := components.NewModal(s.palette, s.palette.Hint)
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
