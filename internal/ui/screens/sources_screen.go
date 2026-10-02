package screens

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dnwSilver/tld/internal/ui/components"
	"github.com/dnwSilver/tld/internal/ui/uikit"
)

type SourcesScreen struct {
	palette uikit.Palette
}

const (
	sourceFormLabelWidth     = 5
	sourceFormValueWidth     = 28
	sourceNameColumnWidth    = 22
	sourceURLColumnMinWidth  = 34
	sourceTypeColumnMinWidth = 10
	sourceColumnGap          = 2
)

func NewSourcesScreen(palette uikit.Palette) SourcesScreen {
	return SourcesScreen{
		palette: palette,
	}
}

func (s SourcesScreen) Render(
	width int,
	height int,
	sources []uikit.Source,
	selectedSourceID int64,
	form uikit.SourceForm,
	deleteConfirm uikit.DeleteConfirm,
	searchQuery string,
) string {
	boxWidth := uikit.Max(width, 2)
	boxHeight := uikit.Max(height-1, 3)
	contentWidth := uikit.Max(boxWidth-2, 1)
	contentHeight := uikit.Max(boxHeight-2, 1)

	count := len(sources)
	title := components.ScreenTitle(s.palette, uikit.SymbolSource, "Sources", &count, searchQuery)
	content := s.renderContent(contentWidth, contentHeight, sources, selectedSourceID, form, deleteConfirm)
	borderColor := s.palette.Primary
	if form.Open || deleteConfirm.Open {
		borderColor = s.palette.Hint
	}
	box := components.NewBox(s.palette, borderColor).Render(contentWidth, contentHeight, title, content)
	footer := components.ScreenFooter(s.palette, width, "Kolosov Aleksandr")

	return lipgloss.JoinVertical(lipgloss.Left, box, footer)
}

func (s SourcesScreen) renderContent(
	width int,
	height int,
	sources []uikit.Source,
	selectedSourceID int64,
	form uikit.SourceForm,
	deleteConfirm uikit.DeleteConfirm,
) string {
	lines := make([]string, 0, height)
	visibleSources := visibleRowsByID(sources, selectedSourceID, height-1, func(source uikit.Source) int64 { return source.ID })
	typeColumnWidth := components.ColumnWidth(visibleSources, func(source uikit.Source) string { return s.sourceType(source) }, sourceTypeColumnMinWidth)
	urlColumnWidth := components.ColumnWidth(visibleSources, func(source uikit.Source) string { return source.URL }, sourceURLColumnMinWidth)
	lines = append(lines, s.tableHeader(width, typeColumnWidth, urlColumnWidth))

	if len(sources) == 0 {
		empty := uikit.Text(s.palette, s.palette.Hint, "No sources yet")
		lines = append(lines, uikit.CenterLine(s.palette, width, empty))
	} else {
		for _, source := range visibleSources {
			lines = append(lines, s.renderSourceRow(width, source, source.ID == selectedSourceID, typeColumnWidth, urlColumnWidth))
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

func (s SourcesScreen) tableHeader(width int, typeColumnWidth int, urlColumnWidth int) string {
	return components.RenderTableRow(s.palette, s.palette.Background, width, []components.TableCell{
		{Value: "name", Width: sourceNameColumnWidth, Foreground: s.palette.Hint, Bold: true},
		{Value: "", Width: sourceColumnGap, Foreground: s.palette.Hint},
		{Value: "type", Width: typeColumnWidth, Foreground: s.palette.Hint, Bold: true},
		{Value: "", Width: sourceColumnGap, Foreground: s.palette.Hint},
		{Value: "url", Width: urlColumnWidth, Foreground: s.palette.Hint, Bold: true},
	})
}

func (s SourcesScreen) renderSourceRow(width int, source uikit.Source, selected bool, typeColumnWidth int, urlColumnWidth int) string {
	background := s.palette.Background
	if selected {
		background = s.palette.Hover
	}

	return components.RenderTableRow(s.palette, background, width, []components.TableCell{
		{Value: source.Name, Width: sourceNameColumnWidth, Foreground: s.palette.Text},
		{Value: "", Width: sourceColumnGap, Foreground: s.palette.Text},
		{Value: s.sourceType(source), Width: typeColumnWidth, Foreground: s.palette.Info},
		{Value: "", Width: sourceColumnGap, Foreground: s.palette.Text},
		{Value: source.URL, Width: urlColumnWidth, Foreground: s.palette.Hint},
	})
}

func (s SourcesScreen) renderModal(lines []string, width, height int, form uikit.SourceForm) []string {
	modal := components.NewModal(s.palette, s.palette.Primary)
	return modal.Overlay(lines, width, height, s.modal(width, form))
}

func (s SourcesScreen) modal(width int, form uikit.SourceForm) string {
	modalWidth := uikit.Min(uikit.Max(width-6, 38), 62)
	contentWidth := uikit.Max(modalWidth-2, 1)
	modal := components.NewModal(s.palette, s.palette.Primary)

	action := "Add"
	if form.Mode == uikit.StackFormModeEdit {
		action = "Edit"
	}
	title := modal.Title(uikit.SymbolSource + " " + action + " source")

	rows := []string{
		modal.CenterLine(contentWidth, s.inputLine("Name", form.Name, form.Focus == uikit.SourceFormFieldName)),
		modal.CenterLine(contentWidth, s.inputLine("URL", form.URL, form.Focus == uikit.SourceFormFieldURL)),
		modal.CenterLine(contentWidth, s.inputLine("PAT", maskSecret(form.PATToken), form.Focus == uikit.SourceFormFieldPATToken)),
		modal.CenterLine(contentWidth, s.inputLine("Type", form.Type, form.Focus == uikit.SourceFormFieldType)),
		modal.CenterLine(contentWidth, modal.Text(s.palette.Error, form.Error)),
		modal.CenterLine(contentWidth, s.actionsLine(form.CanSave, form.Focus == uikit.SourceFormFieldType, false)),
	}
	if form.Type == "registry" {
		rows = []string{
			modal.CenterLine(contentWidth, s.inputLine("Name", form.Name, form.Focus == uikit.SourceFormFieldName)),
			modal.CenterLine(contentWidth, s.inputLine("URL", form.URL, form.Focus == uikit.SourceFormFieldURL)),
			modal.CenterLine(contentWidth, s.inputLine("PAT", maskSecret(form.PATToken), form.Focus == uikit.SourceFormFieldPATToken)),
			modal.CenterLine(contentWidth, s.inputLine("Type", form.Type, form.Focus == uikit.SourceFormFieldType)),
			modal.CenterLine(contentWidth, s.inputLine("Kind", form.RegistryKind, form.Focus == uikit.SourceFormFieldRegistryKind)),
			modal.CenterLine(contentWidth, modal.Text(s.palette.Error, form.Error)),
			modal.CenterLine(contentWidth, s.actionsLine(form.CanSave, form.Focus == uikit.SourceFormFieldType, form.Focus == uikit.SourceFormFieldRegistryKind)),
		}
	}

	return modal.Render(contentWidth, title, rows)
}

func (s SourcesScreen) inputLine(label, value string, focused bool) string {
	return components.RenderFormField(s.palette, components.FormField{
		Label:      label,
		Value:      value,
		Focused:    focused,
		LabelWidth: sourceFormLabelWidth,
		ValueWidth: sourceFormValueWidth,
	})
}

func (s SourcesScreen) actionsLine(canSave bool, typeFocused bool, kindFocused bool) string {
	saveColor := s.palette.Disable
	if canSave {
		saveColor = s.palette.Primary
	}

	actions := []components.Action{
		{Hint: uikit.KeyCancel.Hint, Color: s.palette.Primary},
		{Hint: uikit.KeySave.Hint, Color: saveColor},
	}
	if typeFocused {
		actions = append(actions, components.Action{Hint: uikit.KeySourceTypePick.Hint, Color: s.palette.Hint})
	}
	if kindFocused {
		actions = append(actions, components.Action{Hint: uikit.KeyRegistryKindPick.Hint, Color: s.palette.Hint})
	}

	return components.RenderActions(s.palette, actions)
}

func (s SourcesScreen) sourceType(source uikit.Source) string {
	if source.Type == "registry" && source.RegistryKind != "" {
		return source.Type + "/" + source.RegistryKind
	}
	return source.Type
}

func (s SourcesScreen) renderDeleteConfirm(lines []string, width, height int, confirm uikit.DeleteConfirm) []string {
	modal := components.NewModal(s.palette, s.palette.Primary)
	return modal.Overlay(lines, width, height, s.deleteConfirmModal(width, confirm))
}

func (s SourcesScreen) deleteConfirmModal(width int, confirm uikit.DeleteConfirm) string {
	modalWidth := uikit.Min(uikit.Max(width-6, 34), 58)
	contentWidth := uikit.Max(modalWidth-2, 1)
	modal := components.NewModal(s.palette, s.palette.Primary)
	question := "Delete source " + confirm.Name + "?"
	title := modal.Title(uikit.SymbolError + " Delete source")

	rows := []string{
		modal.CenterLine(contentWidth, modal.Text(s.palette.Hint, question)),
		modal.CenterLine(contentWidth, modal.Text(s.palette.Error, confirm.Error)),
		modal.CenterLine(contentWidth, s.deleteActionsLine()),
	}

	return modal.Render(contentWidth, title, rows)
}

func (s SourcesScreen) deleteActionsLine() string {
	return components.RenderActions(s.palette, []components.Action{
		{Hint: uikit.KeyConfirmNo.Hint, Color: s.palette.Primary},
		{Hint: uikit.KeyConfirmYes.Hint, Color: s.palette.Error},
	})
}

func maskSecret(value string) string {
	return strings.Repeat("*", len([]rune(value)))
}
