package screens

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/dnwSilver/tld/internal/ui/components"
	"github.com/dnwSilver/tld/internal/ui/uikit"
)

type SettingsScreen struct {
	palette uikit.Palette
}

const (
	settingsProjectColumnWidth = 24
	settingsCheckColumnWidth   = 8
)

func NewSettingsScreen(palette uikit.Palette) SettingsScreen {
	return SettingsScreen{palette: palette}
}

func (s SettingsScreen) Render(
	width int,
	height int,
	checkColumns []uikit.ProjectCheck,
	rows []uikit.ProjectCheckRow,
	selectedProjectID int64,
	tableState uikit.SettingsTableState,
	status uikit.SettingsStatus,
	searchQuery string,
) string {
	boxHeight := uikit.Max(height-1, 3)
	contentHeight := uikit.Max(boxHeight-2, 1)
	contentWidth := uikit.Max(width-2, 1)
	count := len(rows)
	title := components.ScreenTitle(s.palette, uikit.SymbolSettings, "Settings", &count, searchQuery)
	body := s.renderContent(contentWidth, contentHeight, checkColumns, rows, selectedProjectID, tableState)
	box := components.NewBox(s.palette, s.palette.Primary).Render(contentWidth, contentHeight, title, body)
	footer := components.ScreenFooter(s.palette, width, "Kolosov Aleksandr")
	if status.Message != "" || status.Error != "" {
		footer = s.progressLine(width, status)
	}
	return lipgloss.JoinVertical(lipgloss.Left, box, footer)
}

func (s SettingsScreen) renderContent(
	width int,
	height int,
	checkColumns []uikit.ProjectCheck,
	rows []uikit.ProjectCheckRow,
	selectedProjectID int64,
	tableState uikit.SettingsTableState,
) string {
	lines := make([]string, 0, height)
	visibleColumns, columnOffset := visibleSettingsColumns(width, checkColumns, tableState.SelectedColumn)
	lines = append(lines, s.tableHeader(width, visibleColumns, columnOffset, tableState))
	if len(rows) == 0 {
		empty := uikit.Text(s.palette, s.palette.Hint, "No projects yet")
		lines = append(lines, uikit.CenterLine(s.palette, width, empty))
	} else {
		for _, row := range visibleRowsByID(rows, selectedProjectID, height-1, func(row uikit.ProjectCheckRow) int64 { return row.ProjectID }) {
			lines = append(lines, s.renderRow(width, visibleColumns, checkColumns, row, row.ProjectID == selectedProjectID))
		}
	}
	return fillLines(s.palette, lines, width, height)
}

func visibleSettingsColumns(width int, columns []uikit.ProjectCheck, selectedColumn int) ([]uikit.ProjectCheck, int) {
	capacity := uikit.Max((width-settingsProjectColumnWidth)/settingsCheckColumnWidth, 1)
	if len(columns) <= capacity {
		return columns, 0
	}
	start := selectedColumn - capacity
	if start < 0 {
		start = 0
	}
	if start+capacity > len(columns) {
		start = len(columns) - capacity
	}
	return columns[start : start+capacity], start
}

func (s SettingsScreen) tableHeader(width int, checkColumns []uikit.ProjectCheck, columnOffset int, tableState uikit.SettingsTableState) string {
	cells := []components.TableCell{
		s.headerCell("project", settingsProjectColumnWidth, 0, tableState),
	}
	for index, check := range checkColumns {
		cells = append(cells, s.headerCell(check.Title, settingsCheckColumnWidth, columnOffset+index+1, tableState))
	}
	return components.RenderTableRow(s.palette, s.palette.Background, width, cells)
}

func (s SettingsScreen) headerCell(title string, width int, column int, tableState uikit.SettingsTableState) components.TableCell {
	color := s.palette.Hint
	if column == tableState.SelectedColumn {
		color = s.palette.Primary
	}
	if tableState.SortActive && column == tableState.SortColumn {
		symbol := uikit.SymbolSortAscending
		if tableState.SortDescending {
			symbol = uikit.SymbolSortDescending
		}
		title = symbol + title
	}

	return components.TableCell{
		Value:      title,
		Width:      width,
		Foreground: color,
		Bold:       true,
	}
}

func (s SettingsScreen) renderRow(width int, checkColumns, allCheckColumns []uikit.ProjectCheck, row uikit.ProjectCheckRow, selected bool) string {
	background := s.palette.Background
	if selected {
		background = s.palette.Hover
	}
	iconColor := s.projectIconColor(allCheckColumns, row.Results, lipgloss.Color(uikit.NormalizeHexColor(row.ProjectColor)))
	projectCell := uikit.RenderProjectWithIcon(
		s.palette,
		row.ProjectIcon,
		iconColor,
		row.ProjectName,
		row.ProjectFreezing,
		row.ProjectEndOfLife,
		background,
		s.palette.Text,
		2,
	)

	cells := []components.TableCell{
		{Value: projectCell, Width: settingsProjectColumnWidth, Foreground: s.palette.Text},
	}
	for _, check := range checkColumns {
		state := uikit.CheckStateUnknown
		if row.Results != nil {
			if value, ok := row.Results[check.ID]; ok {
				state = value
			}
		}
		version := ""
		if row.Versions != nil {
			version = row.Versions[check.ID]
		}
		cells = append(cells, components.TableCell{
			Value:      s.checkValue(state, version),
			Width:      settingsCheckColumnWidth,
			Foreground: s.checkColor(state),
		})
	}
	return components.RenderTableRow(s.palette, background, width, cells)
}

func (s SettingsScreen) projectIconColor(checkColumns []uikit.ProjectCheck, results map[string]uikit.CheckState, defaultColor lipgloss.Color) lipgloss.Color {
	if len(checkColumns) == 0 || len(results) == 0 {
		return defaultColor
	}

	hasApplicableCheck := false
	for _, check := range checkColumns {
		state, ok := results[check.ID]
		if !ok {
			return defaultColor
		}
		switch state {
		case uikit.CheckStatePass:
			hasApplicableCheck = true
		case uikit.CheckStateNotApplicable:
			continue
		default:
			return defaultColor
		}
	}
	if hasApplicableCheck {
		return s.palette.Primary
	}

	return defaultColor
}

func (s SettingsScreen) checkValue(state uikit.CheckState, version string) string {
	value := s.checkSymbol(state)
	if version != "" {
		value += " " + version
	}
	return value
}

func (s SettingsScreen) checkSymbol(state uikit.CheckState) string {
	switch state {
	case uikit.CheckStatePass:
		return uikit.SymbolCheckPass
	case uikit.CheckStateWarning:
		return uikit.SymbolCheckWarning
	case uikit.CheckStateFail:
		return uikit.SymbolCheckFail
	case uikit.CheckStateNotApplicable:
		return uikit.SymbolCheckNotApplicable
	default:
		return uikit.SymbolCheckUnknown
	}
}

func (s SettingsScreen) checkColor(state uikit.CheckState) lipgloss.Color {
	switch state {
	case uikit.CheckStatePass:
		return s.palette.Primary
	case uikit.CheckStateWarning:
		return s.palette.Warning
	case uikit.CheckStateFail:
		return s.palette.Error
	default:
		return s.palette.Hint
	}
}

func (s SettingsScreen) progressLine(width int, status uikit.SettingsStatus) string {
	message := status.Message
	if status.Running {
		message += " [Esc] cancel"
	}
	return components.RenderProgress(s.palette, width, uikit.SymbolSettings, message, status.Error, status.Running, status.Current, status.Total)
}
