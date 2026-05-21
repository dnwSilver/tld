package screens

import (
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dnwSilver/tld/internal/ui/components"
	"github.com/dnwSilver/tld/internal/ui/uikit"
)

type StacksScreen struct {
	palette uikit.Palette
}

const (
	stackFormLabelWidth = 5
	stackFormValueWidth = 24
)

func NewStacksScreen(palette uikit.Palette) StacksScreen {
	return StacksScreen{
		palette: palette,
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

	content := s.renderContent(contentWidth, contentHeight, stacks, selectedStackID, form, deleteConfirm)
	box := s.renderBox(contentWidth, contentHeight, s.title(len(stacks)), content)

	footer := lipgloss.NewStyle().
		Width(width).
		Background(s.palette.Background).
		Foreground(s.palette.Hint).
		Align(lipgloss.Right).
		Render(uikit.SymbolCopyright + " Kolosov Aleksandr")

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
	iconColumnWidth := s.iconColumnWidth(stacks)
	lines = append(lines, s.tableHeader(width, iconColumnWidth))

	if len(stacks) == 0 {
		empty := lipgloss.NewStyle().
			Background(s.palette.Background).
			Foreground(s.palette.Hint).
			Render("No stacks yet")
		lines = append(lines, s.centerLine(width, empty))
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

func (s StacksScreen) title(stackCount int) string {
	return lipgloss.NewStyle().
		Background(s.palette.Background).
		Foreground(s.palette.Warning).
		Bold(true).
		Render(uikit.SymbolStack + " Stacks [" + uikit.FormatInt(stackCount) + "]")
}

func (s StacksScreen) renderBox(width, height int, title, content string) string {
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

func (s StacksScreen) topBorder(width int, title string) string {
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

func (s StacksScreen) borderFill(width int) string {
	return s.borderPart(strings.Repeat("─", uikit.Max(width, 0)))
}

func (s StacksScreen) borderPart(value string) string {
	return lipgloss.NewStyle().
		Background(s.palette.Background).
		Foreground(s.palette.Primary).
		Render(value)
}

func (s StacksScreen) tableHeader(width int, iconColumnWidth int) string {
	iconColumn := lipgloss.NewStyle().
		Background(s.palette.Background).
		Foreground(s.palette.Hint).
		Width(iconColumnWidth).
		Render("")
	nameColumn := lipgloss.NewStyle().
		Background(s.palette.Background).
		Foreground(s.palette.Hint).
		Bold(true).
		Render("name")
	content := iconColumn + nameColumn

	return content + uikit.BackgroundSpaces(s.palette, width-lipgloss.Width(content))
}

func (s StacksScreen) iconColumnWidth(stacks []uikit.Stack) int {
	width := 2
	for _, stack := range stacks {
		width = uikit.Max(width, lipgloss.Width(stack.Icon))
	}

	return width
}

func (s StacksScreen) renderStackRow(width int, stack uikit.Stack, selected bool, iconColumnWidth int) string {
	rowPalette := s.palette
	background := s.palette.Background
	foreground := s.palette.Text
	if selected {
		background = s.palette.Hover
	}

	color := lipgloss.Color(uikit.NormalizeHexColor(stack.Color))
	iconColumn := lipgloss.NewStyle().
		Background(background).
		Foreground(color).
		Width(iconColumnWidth).
		Render(stack.Icon)
	nameColumn := lipgloss.NewStyle().
		Background(background).
		Foreground(foreground).
		Render(stack.Name)
	content := iconColumn + nameColumn
	rowPalette.Background = background

	return content + uikit.BackgroundSpaces(rowPalette, width-lipgloss.Width(content))
}

func (s StacksScreen) centerLine(width int, content string) string {
	contentWidth := lipgloss.Width(content)
	left := uikit.Max((width-contentWidth)/2, 0)
	right := uikit.Max(width-left-contentWidth, 0)

	return uikit.BackgroundSpaces(s.palette, left) + content + uikit.BackgroundSpaces(s.palette, right)
}

func (s StacksScreen) renderModal(lines []string, width, height int, form uikit.StackForm) []string {
	modal := strings.Split(s.modal(width, form), uikit.SymbolLineBreak)
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

func (s StacksScreen) modal(width int, form uikit.StackForm) string {
	modalWidth := uikit.Min(uikit.Max(width-6, 32), 54)
	contentWidth := uikit.Max(modalWidth-2, 1)
	modal := components.NewModal(s.palette, s.palette.Primary)
	title := modal.Title(uikit.SymbolStack + " Add stack")
	if form.Mode == uikit.StackFormModeEdit {
		title = modal.Title(uikit.SymbolStack + " Edit stack")
	}
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
		Width(stackFormLabelWidth).
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
	valuePadding := uikit.BackgroundSpaces(s.palette, stackFormValueWidth-lipgloss.Width(valuePart))

	return prefixPart + labelPart + separatorPart + valuePart + valuePadding
}

func (s StacksScreen) actionsLine(canSave bool) string {
	saveColor := s.palette.Disable
	if canSave {
		saveColor = s.palette.Primary
	}

	modal := components.NewModal(s.palette, s.palette.Primary)
	cancel := modal.Text(s.palette.Warning, "[Esc] Cancel")
	save := modal.Text(saveColor, "[Enter] Save")

	return cancel + modal.Spaces(3) + save
}

func (s StacksScreen) renderDeleteConfirm(lines []string, width, height int, confirm uikit.DeleteConfirm) []string {
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

func (s StacksScreen) deleteConfirmModal(width int, confirm uikit.DeleteConfirm) string {
	modalWidth := uikit.Min(uikit.Max(width-6, 34), 58)
	contentWidth := uikit.Max(modalWidth-2, 1)
	modal := components.NewModal(s.palette, s.palette.Error)
	question := "Delete stack " + confirm.Name + "?"
	title := modal.Title(uikit.SymbolError + " Delete stack")
	rows := []string{
		modal.CenterLine(contentWidth, modal.Text(s.palette.Hint, question)),
		modal.CenterLine(contentWidth, modal.Text(s.palette.Error, confirm.Error)),
		modal.CenterLine(contentWidth, s.deleteActionsLine()),
	}

	return modal.Render(contentWidth, title, rows)
}

func (s StacksScreen) deleteActionsLine() string {
	modal := components.NewModal(s.palette, s.palette.Error)
	no := modal.Text(s.palette.Primary, "[Esc] No")
	yes := modal.Text(s.palette.Error, "[Enter] Yes")

	return no + modal.Spaces(3) + yes
}
