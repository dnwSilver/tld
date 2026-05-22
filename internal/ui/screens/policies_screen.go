package screens

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dnwSilver/tld/internal/ui/components"
	"github.com/dnwSilver/tld/internal/ui/uikit"
)

type PoliciesScreen struct {
	palette uikit.Palette
}

const (
	policyFormLabelWidth   = 9
	policyFormValueWidth   = 28
	policyNameColumnWidth  = 22
	policyDepColumnWidth   = 24
	policyValueColumnWidth = 14
)

func NewPoliciesScreen(palette uikit.Palette) PoliciesScreen {
	return PoliciesScreen{palette: palette}
}

func (s PoliciesScreen) Render(
	width int,
	height int,
	policies []uikit.Policy,
	selectedPolicyID int64,
	values []uikit.PolicyValue,
	selectedValueID int64,
	focus uikit.PolicyPane,
	namespaces []uikit.Namespace,
	dependencies []uikit.Dependency,
	policyForm uikit.PolicyForm,
	valueForm uikit.PolicyValueForm,
	deleteConfirm uikit.DeleteConfirm,
) string {
	boxHeight := uikit.Max(height-1, 3)
	contentHeight := uikit.Max(boxHeight-2, 1)
	hasPolicy := selectedPolicyID != 0
	modalOpen := policyForm.Open || valueForm.Open || deleteConfirm.Open
	leftTotalWidth := uikit.Max(width, 2)
	rightTotalWidth := 0
	if hasPolicy {
		leftTotalWidth = uikit.Max(width/2, 2)
		rightTotalWidth = uikit.Max(width-leftTotalWidth, 2)
	}
	leftContentWidth := uikit.Max(leftTotalWidth-2, 1)
	rightContentWidth := uikit.Max(rightTotalWidth-2, 1)

	leftTitle := components.ScreenTitle(s.palette, uikit.SymbolPolicy, "Policies", intPtr(len(policies)))
	leftBorder := s.borderForPane(!modalOpen && focus == uikit.PolicyPanePolicies)
	left := components.NewBox(s.palette, leftBorder).Render(leftContentWidth, contentHeight, leftTitle, s.renderPoliciesContent(leftContentWidth, contentHeight, policies, selectedPolicyID))
	body := left
	if hasPolicy {
		rightTitle := components.ScreenTitle(s.palette, uikit.SymbolDependency, "Pinned deps", intPtr(len(values)))
		rightBorder := s.borderForPane(!modalOpen && focus == uikit.PolicyPaneValues)
		right := components.NewBox(s.palette, rightBorder).Render(rightContentWidth, contentHeight, rightTitle, s.renderValuesContent(rightContentWidth, contentHeight, values, selectedValueID))
		body = lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	}
	lines := strings.Split(body, uikit.SymbolLineBreak)

	if policyForm.Open {
		lines = s.overlay(lines, width, boxHeight, s.policyModal(width, namespaces, policyForm))
	}
	if valueForm.Open {
		lines = s.overlay(lines, width, boxHeight, s.valueModal(width, dependencies, valueForm))
	}
	if deleteConfirm.Open {
		lines = s.overlay(lines, width, boxHeight, s.deleteConfirmModal(width, focus, deleteConfirm))
	}

	footer := components.ScreenFooter(s.palette, width, "Kolosov Aleksandr")
	return lipgloss.JoinVertical(lipgloss.Left, strings.Join(lines, uikit.SymbolLineBreak), footer)
}

func (s PoliciesScreen) borderForPane(active bool) lipgloss.Color {
	if active {
		return s.palette.Primary
	}

	return s.palette.Hint
}

func (s PoliciesScreen) renderPoliciesContent(width, height int, policies []uikit.Policy, selectedPolicyID int64) string {
	lines := make([]string, 0, height)
	lines = append(lines, s.policiesHeader(width))
	if len(policies) == 0 {
		empty := uikit.Text(s.palette, s.palette.Hint, "No policies yet")
		lines = append(lines, uikit.CenterLine(s.palette, width, empty))
	} else {
		for _, policy := range policies {
			lines = append(lines, s.policyRow(width, policy, policy.ID == selectedPolicyID))
		}
	}
	return fillLines(s.palette, lines, width, height)
}

func (s PoliciesScreen) renderValuesContent(width, height int, values []uikit.PolicyValue, selectedValueID int64) string {
	lines := make([]string, 0, height)
	lines = append(lines, s.valuesHeader(width))
	if len(values) == 0 {
		empty := uikit.Text(s.palette, s.palette.Hint, "No pinned deps yet")
		lines = append(lines, uikit.CenterLine(s.palette, width, empty))
	} else {
		for _, value := range values {
			lines = append(lines, s.valueRow(width, value, value.ID == selectedValueID))
		}
	}
	return fillLines(s.palette, lines, width, height)
}

func (s PoliciesScreen) policiesHeader(width int) string {
	return components.RenderTableRow(s.palette, s.palette.Background, width, []components.TableCell{
		{Value: "name", Width: policyNameColumnWidth, Foreground: s.palette.Hint, Bold: true},
	})
}

func (s PoliciesScreen) policyRow(width int, policy uikit.Policy, selected bool) string {
	background := s.palette.Background
	if selected {
		background = s.palette.Hover
	}
	return components.RenderTableRow(s.palette, background, width, []components.TableCell{
		{Value: policy.Name, Width: policyNameColumnWidth, Foreground: s.palette.Text},
	})
}

func (s PoliciesScreen) valuesHeader(width int) string {
	return components.RenderTableRow(s.palette, s.palette.Background, width, []components.TableCell{
		{Value: "dependency", Width: policyDepColumnWidth, Foreground: s.palette.Hint, Bold: true},
		{Value: "version", Width: policyValueColumnWidth, Foreground: s.palette.Hint, Bold: true},
	})
}

func (s PoliciesScreen) valueRow(width int, value uikit.PolicyValue, selected bool) string {
	background := s.palette.Background
	if selected {
		background = s.palette.Hover
	}
	color := lipgloss.Color(uikit.NormalizeHexColor(value.DependencyColor))
	return components.RenderTableRow(s.palette, background, width, []components.TableCell{
		{Value: value.DependencyIcon + " " + value.DependencyName, Width: policyDepColumnWidth, Foreground: color},
		{Value: value.Version, Width: policyValueColumnWidth, Foreground: s.palette.Text},
	})
}

func (s PoliciesScreen) policyModal(width int, namespaces []uikit.Namespace, form uikit.PolicyForm) string {
	modalWidth := uikit.Min(uikit.Max(width-6, 40), 64)
	contentWidth := uikit.Max(modalWidth-2, 1)
	modal := components.NewModal(s.palette, s.palette.Primary)
	action := "Add"
	if form.Mode == uikit.StackFormModeEdit {
		action = "Edit"
	}
	rows := []string{
		modal.CenterLine(contentWidth, s.inputLine("Name", form.Name, form.Focus == uikit.PolicyFormFieldName)),
		modal.CenterLine(contentWidth, modal.Text(s.palette.Error, form.Error)),
		modal.CenterLine(contentWidth, s.actionsLine(form.CanSave, false, "")),
	}
	return modal.Render(contentWidth, modal.Title(uikit.SymbolPolicy+" "+action+" policy"), rows)
}

func (s PoliciesScreen) valueModal(width int, dependencies []uikit.Dependency, form uikit.PolicyValueForm) string {
	modalWidth := uikit.Min(uikit.Max(width-6, 40), 64)
	contentWidth := uikit.Max(modalWidth-2, 1)
	modal := components.NewModal(s.palette, s.palette.Primary)
	action := "Add"
	if form.Mode == uikit.StackFormModeEdit {
		action = "Edit"
	}
	rows := []string{
		modal.CenterLine(contentWidth, s.inputLine("Dep", s.dependencyName(dependencies, form.DependencyID), form.Focus == uikit.PolicyValueFormFieldDependency)),
		modal.CenterLine(contentWidth, s.inputLine("Version", form.Version, form.Focus == uikit.PolicyValueFormFieldVersion)),
		modal.CenterLine(contentWidth, modal.Text(s.palette.Error, form.Error)),
		modal.CenterLine(contentWidth, s.actionsLine(form.CanSave, form.Focus == uikit.PolicyValueFormFieldDependency, uikit.KeyDependencyPick.Hint)),
	}
	return modal.Render(contentWidth, modal.Title(uikit.SymbolDependency+" "+action+" pinned dep"), rows)
}

func (s PoliciesScreen) inputLine(label, value string, focused bool) string {
	return components.RenderFormField(s.palette, components.FormField{
		Label:      label,
		Value:      value,
		Focused:    focused,
		LabelWidth: policyFormLabelWidth,
		ValueWidth: policyFormValueWidth,
	})
}

func (s PoliciesScreen) actionsLine(canSave bool, pickFocused bool, pickHint string) string {
	saveColor := s.palette.Disable
	if canSave {
		saveColor = s.palette.Primary
	}
	actions := []components.Action{
		{Hint: uikit.KeyCancel.Hint, Color: s.palette.Primary},
		{Hint: uikit.KeySave.Hint, Color: saveColor},
	}
	if pickFocused {
		actions = append(actions, components.Action{Hint: pickHint, Color: s.palette.Hint})
	}
	return components.RenderActions(s.palette, actions)
}

func (s PoliciesScreen) deleteConfirmModal(width int, focus uikit.PolicyPane, confirm uikit.DeleteConfirm) string {
	modalWidth := uikit.Min(uikit.Max(width-6, 34), 58)
	contentWidth := uikit.Max(modalWidth-2, 1)
	modal := components.NewModal(s.palette, s.palette.Primary)
	target := "policy"
	if focus == uikit.PolicyPaneValues {
		target = "pinned dep"
	}
	rows := []string{
		modal.CenterLine(contentWidth, modal.Text(s.palette.Hint, "Delete "+target+" "+confirm.Name+"?")),
		modal.CenterLine(contentWidth, modal.Text(s.palette.Error, confirm.Error)),
		modal.CenterLine(contentWidth, components.RenderActions(s.palette, []components.Action{
			{Hint: uikit.KeyConfirmNo.Hint, Color: s.palette.Primary},
			{Hint: uikit.KeyConfirmYes.Hint, Color: s.palette.Error},
		})),
	}
	return modal.Render(contentWidth, modal.Title(uikit.SymbolError+" Delete "+target), rows)
}

func (s PoliciesScreen) overlay(lines []string, width, height int, block string) []string {
	modal := components.NewModal(s.palette, s.palette.Primary)
	return modal.Overlay(lines, width, height, block)
}

func (s PoliciesScreen) namespaceName(namespaces []uikit.Namespace, id int64) string {
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

func (s PoliciesScreen) dependencyName(dependencies []uikit.Dependency, id int64) string {
	for _, dependency := range dependencies {
		if dependency.ID == id {
			return dependency.Name
		}
	}
	if len(dependencies) == 0 {
		return "No deps"
	}
	return ""
}

func fillLines(palette uikit.Palette, lines []string, width, height int) string {
	for len(lines) < height {
		lines = append(lines, uikit.BackgroundSpaces(palette, width))
	}
	if len(lines) > height {
		lines = lines[:height]
	}
	return strings.Join(lines, uikit.SymbolLineBreak)
}

func intPtr(value int) *int {
	return &value
}
