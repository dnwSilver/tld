package screens

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/dnwSilver/tld/internal/ui/components"
	"github.com/dnwSilver/tld/internal/ui/uikit"
)

type DefaultScreen struct {
	palette uikit.Palette
}

const (
	dashboardProjectColumnMinWidth    = 9
	dashboardProjectColumnPadding     = 2
	dashboardCriticalColumnWidth      = 6
	dashboardHighColumnWidth          = 6
	dashboardMajorColumnWidth         = 7
	dashboardMinorPatchColumnWidth    = 10
	dashboardSettingsErrorColumnWidth = 8
	dashboardSettingsWarnColumnWidth  = 9
)

func NewDefaultScreen(palette uikit.Palette) DefaultScreen {
	return DefaultScreen{
		palette: palette,
	}
}

func (s DefaultScreen) Render(
	width int,
	height int,
	stackCount int,
	attentionRows []uikit.DashboardAttentionRow,
	focus uikit.DashboardPane,
	selectedProjectID int64,
) string {
	boxHeight := uikit.Max(height-1, 3)
	contentHeight := uikit.Max(boxHeight-2, 1)
	leftTotalWidth := uikit.Max(width/3, 2)
	rightTotalWidth := uikit.Max(width-leftTotalWidth, 2)
	leftContentWidth := uikit.Max(leftTotalWidth-2, 1)
	rightContentWidth := uikit.Max(rightTotalWidth-2, 1)

	leftTitle := components.ScreenTitle(s.palette, uikit.SymbolDashboard, "Team lead dashboard", nil)
	left := components.NewBox(s.palette, s.borderColor(focus == uikit.DashboardPaneSummary)).
		Render(leftContentWidth, contentHeight, leftTitle, s.renderDashboardContent(leftContentWidth, contentHeight, stackCount))
	count := len(attentionRows)
	rightTitle := uikit.BoldText(
		s.palette,
		s.palette.Warning,
		uikit.SymbolCheckWarning+" Attention ["+uikit.FormatInt(count)+"]",
	)
	right := components.NewBox(s.palette, s.borderColor(focus == uikit.DashboardPaneAttention)).
		Render(rightContentWidth, contentHeight, rightTitle, s.renderAttentionContent(rightContentWidth, contentHeight, attentionRows, selectedProjectID))

	footer := components.ScreenFooter(s.palette, width, "Kolosov Aleksandr")

	return lipgloss.JoinVertical(lipgloss.Left, lipgloss.JoinHorizontal(lipgloss.Top, left, right), footer)
}

func (s DefaultScreen) borderColor(active bool) lipgloss.Color {
	if active {
		return s.palette.Primary
	}
	return s.palette.Hint
}

func (s DefaultScreen) renderDashboardContent(width, height int, stackCount int) string {
	counter := uikit.Text(s.palette, s.palette.Primary, "Stacks: "+uikit.FormatInt(stackCount))
	lines := make([]string, 0, height)
	counterRow := height / 2
	for row := range height {
		if row == counterRow {
			lines = append(lines, uikit.CenterLine(s.palette, width, counter))
			continue
		}
		lines = append(lines, uikit.BackgroundSpaces(s.palette, width))
	}
	return fillLines(s.palette, lines, width, height)
}

func (s DefaultScreen) renderAttentionContent(width, height int, rows []uikit.DashboardAttentionRow, selectedProjectID int64) string {
	columns := dashboardTableColumnsFor(width, rows)
	lines := []string{s.attentionTableHeader(width, columns)}
	visibleRows := visibleDashboardAttentionRows(rows, selectedProjectID, height-1)
	if len(rows) == 0 {
		empty := uikit.Text(s.palette, s.palette.Hint, "No projects require attention")
		lines = append(lines, uikit.CenterLine(s.palette, width, empty))
	} else {
		for _, row := range visibleRows {
			lines = append(lines, s.attentionTableRow(width, columns, row, row.ProjectID == selectedProjectID))
		}
	}

	return fillLines(s.palette, lines, width, height)
}

type dashboardTableColumns struct {
	project int
}

func dashboardTableColumnsFor(width int, rows []uikit.DashboardAttentionRow) dashboardTableColumns {
	fixedWidth := dashboardFixedColumnsWidth()
	available := uikit.Max(width-fixedWidth, 1)
	projectNeed := components.ColumnWidth(rows, func(row uikit.DashboardAttentionRow) string {
		return row.ProjectName
	}, dashboardProjectColumnMinWidth)
	projectWidth := uikit.Min(projectNeed+dashboardProjectColumnPadding, available)
	return dashboardTableColumns{project: projectWidth}
}

func (s DefaultScreen) attentionTableHeader(width int, columns dashboardTableColumns) string {
	return components.RenderTableRow(s.palette, s.palette.Background, width, []components.TableCell{
		{Value: "project", Width: columns.project, Foreground: s.palette.Hint, Bold: true},
		{Value: rightAligned("crit", dashboardCriticalColumnWidth), Width: dashboardCriticalColumnWidth, Foreground: s.palette.Critical, Bold: true},
		{Value: rightAligned("high", dashboardHighColumnWidth), Width: dashboardHighColumnWidth, Foreground: s.palette.Error, Bold: true},
		{Value: rightAligned("major", dashboardMajorColumnWidth), Width: dashboardMajorColumnWidth, Foreground: s.palette.Warning, Bold: true},
		{Value: rightAligned("min/patch", dashboardMinorPatchColumnWidth), Width: dashboardMinorPatchColumnWidth, Foreground: s.palette.Primary, Bold: true},
		{Value: rightAligned("errors", dashboardSettingsErrorColumnWidth), Width: dashboardSettingsErrorColumnWidth, Foreground: s.palette.Error, Bold: true},
		{Value: rightAligned("warnings", dashboardSettingsWarnColumnWidth), Width: dashboardSettingsWarnColumnWidth, Foreground: s.palette.Warning, Bold: true},
	})
}

func (s DefaultScreen) attentionTableRow(width int, columns dashboardTableColumns, row uikit.DashboardAttentionRow, selected bool) string {
	background := s.palette.Background
	if selected {
		background = s.palette.Hover
	}
	projectName := lipgloss.NewStyle().MaxWidth(columns.project).Render(row.ProjectName)
	return components.RenderTableRow(s.palette, background, width, []components.TableCell{
		{Value: projectName, Width: columns.project, Foreground: s.palette.Text},
		s.attentionCountCell(row.Critical, dashboardCriticalColumnWidth, s.palette.Critical),
		s.attentionCountCell(row.High, dashboardHighColumnWidth, s.palette.Error),
		s.attentionCountCell(row.Major, dashboardMajorColumnWidth, s.palette.Warning),
		s.attentionCountCell(row.MinorPatch, dashboardMinorPatchColumnWidth, s.palette.Primary),
		s.attentionCountCell(row.SettingsErrors, dashboardSettingsErrorColumnWidth, s.palette.Error),
		s.attentionCountCell(row.SettingsWarnings, dashboardSettingsWarnColumnWidth, s.palette.Warning),
	})
}

func (s DefaultScreen) attentionCountCell(value, width int, color lipgloss.Color) components.TableCell {
	if value == 0 {
		color = s.palette.Hint
	}
	return components.TableCell{
		Value:      rightAligned(uikit.FormatInt(value), width),
		Width:      width,
		Foreground: color,
		Bold:       value > 0,
	}
}

func dashboardFixedColumnsWidth() int {
	return dashboardCriticalColumnWidth +
		dashboardHighColumnWidth +
		dashboardMajorColumnWidth +
		dashboardMinorPatchColumnWidth +
		dashboardSettingsErrorColumnWidth +
		dashboardSettingsWarnColumnWidth
}

func visibleDashboardAttentionRows(rows []uikit.DashboardAttentionRow, selectedProjectID int64, visibleRows int) []uikit.DashboardAttentionRow {
	if visibleRows <= 0 || len(rows) <= visibleRows {
		return rows
	}

	selectedIndex := 0
	for index, row := range rows {
		if row.ProjectID == selectedProjectID {
			selectedIndex = index
			break
		}
	}
	start := selectedIndex - visibleRows + 1
	if start < 0 {
		start = 0
	}
	if start+visibleRows > len(rows) {
		start = len(rows) - visibleRows
	}
	return rows[start : start+visibleRows]
}
