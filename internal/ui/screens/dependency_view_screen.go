package screens

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dnwSilver/tld/internal/ui/components"
	"github.com/dnwSilver/tld/internal/ui/uikit"
)

const (
	viewProjectColumnWidth = 18
	viewVersionColumnWidth = 14
	viewProjectIconWidth   = 2
)

var viewVersionNumberPattern = regexp.MustCompile(`\d+`)

type DependencyViewScreen struct {
	palette uikit.Palette
}

func NewDependencyViewScreen(palette uikit.Palette) DependencyViewScreen {
	return DependencyViewScreen{palette: palette}
}

func (s DependencyViewScreen) Render(width int, height int, stacks []uikit.Stack, activeStackID int64, view uikit.DependencyView, selectedProjectID int64, columnOffset int) string {
	boxWidth := uikit.Max(width, 2)
	boxHeight := uikit.Max(height-1, 3)
	contentWidth := uikit.Max(boxWidth-2, 1)
	contentHeight := uikit.Max(boxHeight-2, 1)

	title := components.ScreenTitle(s.palette, uikit.SymbolDependency, "View", intPtr(len(view.Rows)))
	content := s.renderContent(contentWidth, contentHeight, view, selectedProjectID, columnOffset)
	box := components.NewTabbedPanel(s.palette, s.palette.Primary).Render(contentWidth, contentHeight, title, stackTabs(stacks, activeStackID), content)
	footer := components.ScreenFooter(s.palette, width, "Kolosov Aleksandr")

	return lipgloss.JoinVertical(lipgloss.Left, box, footer)
}

func (s DependencyViewScreen) renderContent(width int, height int, view uikit.DependencyView, selectedProjectID int64, columnOffset int) string {
	lines := make([]string, 0, height)
	columns := visibleViewColumns(width, view.Columns, columnOffset)
	lines = append(lines, s.header(width, columns))
	if len(view.Rows) == 0 {
		lines = append(lines, uikit.CenterLine(s.palette, width, uikit.Text(s.palette, s.palette.Hint, "No projects for stack")))
	} else {
		for _, row := range view.Rows {
			lines = append(lines, s.row(width, columns, row, row.ProjectID == selectedProjectID))
		}
	}

	return fillLines(s.palette, lines, width, height)
}

func visibleViewColumns(width int, columns []uikit.DependencyViewColumn, offset int) []uikit.DependencyViewColumn {
	if len(columns) == 0 {
		return columns
	}
	visibleCount := uikit.Max((width-viewProjectColumnWidth)/viewVersionColumnWidth, 1)
	if visibleCount >= len(columns) {
		return columns
	}
	if offset < 0 {
		offset = 0
	}
	maxOffset := len(columns) - visibleCount
	if offset > maxOffset {
		offset = maxOffset
	}

	return columns[offset : offset+visibleCount]
}

func stackTabs(stacks []uikit.Stack, activeStackID int64) []components.Tab {
	tabs := make([]components.Tab, 0, len(stacks))
	for _, stack := range stacks {
		tabs = append(tabs, components.Tab{
			ID:     stack.ID,
			Icon:   stack.Icon,
			Label:  stack.Name,
			Active: stack.ID == activeStackID,
		})
	}

	return tabs
}

func (s DependencyViewScreen) header(width int, columns []uikit.DependencyViewColumn) string {
	cells := []components.TableCell{
		{Value: "project", Width: viewProjectColumnWidth, Foreground: s.palette.Hint, Bold: true},
	}
	for _, column := range columns {
		cells = append(cells, components.TableCell{
			Value:      s.headerCell(column),
			Width:      viewVersionColumnWidth,
			Foreground: s.palette.Text,
			Bold:       true,
		})
	}

	return components.RenderTableRow(s.palette, s.palette.Background, width, cells)
}

func (s DependencyViewScreen) row(width int, columns []uikit.DependencyViewColumn, row uikit.DependencyViewRow, selected bool) string {
	background := s.palette.Background
	if selected {
		background = s.palette.Hover
	}
	cells := []components.TableCell{
		{Value: s.projectCell(row, background), Width: viewProjectColumnWidth, Foreground: s.palette.Text},
	}
	for _, column := range columns {
		version := strings.TrimSpace(row.Versions[column.DependencyID])
		cells = append(cells, components.TableCell{
			Value:      rightAligned(s.versionText(version), viewVersionColumnWidth),
			Width:      viewVersionColumnWidth,
			Foreground: s.versionColor(version, column.PolicyVersion),
		})
	}

	return components.RenderTableRow(s.palette, background, width, cells)
}

func (s DependencyViewScreen) versionText(version string) string {
	if strings.TrimSpace(version) == "" {
		return uikit.SymbolNotFound
	}

	return version
}

func (s DependencyViewScreen) projectCell(row uikit.DependencyViewRow, background lipgloss.Color) string {
	icon := lipgloss.NewStyle().
		Background(background).
		Foreground(lipgloss.Color(uikit.NormalizeHexColor(row.ProjectColor))).
		Width(viewProjectIconWidth).
		Render(row.ProjectIcon)

	return icon + row.ProjectName
}

func (s DependencyViewScreen) headerCell(column uikit.DependencyViewColumn) string {
	icon := lipgloss.NewStyle().
		Background(s.palette.Background).
		Foreground(lipgloss.Color(uikit.NormalizeHexColor(column.Color))).
		Width(2).
		Render(column.Icon)
	version := lipgloss.NewStyle().
		Background(s.palette.Background).
		Foreground(s.palette.Text).
		Render(strings.TrimSpace(column.PolicyVersion))
	content := icon + version

	return lipgloss.NewStyle().
		Background(s.palette.Background).
		Width(viewVersionColumnWidth).
		Align(lipgloss.Right).
		Render(content)
}

func rightAligned(value string, width int) string {
	valueWidth := lipgloss.Width(value)
	if valueWidth >= width {
		return value
	}

	return strings.Repeat(" ", width-valueWidth) + value
}

func (s DependencyViewScreen) versionColor(actual string, policy string) lipgloss.Color {
	actual = strings.TrimSpace(actual)
	policy = strings.TrimSpace(policy)
	if actual == "" || policy == "" {
		return s.palette.Hint
	}
	if actual == policy {
		return s.palette.Hint
	}

	actualMajor, actualParts, actualOK := parseViewVersion(actual)
	policyMajor, policyParts, policyOK := parseViewVersion(policy)
	if !actualOK || !policyOK {
		return s.palette.Text
	}
	if actualMajor > policyMajor {
		return s.palette.Info
	}
	if policyMajor-actualMajor > 1 {
		return s.palette.Warning
	}
	if compareViewVersions(actualParts, policyParts) < 0 {
		return s.palette.Primary
	}
	if compareViewVersions(actualParts, policyParts) == 0 {
		return s.palette.Hint
	}

	return s.palette.Info
}

func parseViewVersion(version string) (int, []int, bool) {
	matches := viewVersionNumberPattern.FindAllString(version, -1)
	if len(matches) == 0 {
		return 0, nil, false
	}

	parts := make([]int, 0, len(matches))
	for _, match := range matches {
		value, err := strconv.Atoi(match)
		if err != nil {
			return 0, nil, false
		}
		parts = append(parts, value)
	}

	return parts[0], parts, true
}

func compareViewVersions(left []int, right []int) int {
	maxLength := uikit.Max(len(left), len(right))
	for index := range maxLength {
		leftValue := 0
		if index < len(left) {
			leftValue = left[index]
		}
		rightValue := 0
		if index < len(right) {
			rightValue = right[index]
		}
		if leftValue < rightValue {
			return -1
		}
		if leftValue > rightValue {
			return 1
		}
	}

	return 0
}
