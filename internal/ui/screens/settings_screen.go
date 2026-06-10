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
	status uikit.SettingsStatus,
) string {
	boxHeight := uikit.Max(height-1, 3)
	contentHeight := uikit.Max(boxHeight-2, 1)
	contentWidth := uikit.Max(width-2, 1)
	count := len(rows)
	title := components.ScreenTitle(s.palette, uikit.SymbolSettings, "Settings", &count)
	body := s.renderContent(contentWidth, contentHeight, checkColumns, rows, selectedProjectID)
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
) string {
	lines := make([]string, 0, height)
	lines = append(lines, s.tableHeader(width, checkColumns))
	if len(rows) == 0 {
		empty := uikit.Text(s.palette, s.palette.Hint, "No projects yet")
		lines = append(lines, uikit.CenterLine(s.palette, width, empty))
	} else {
		for _, row := range rows {
			lines = append(lines, s.renderRow(width, checkColumns, row, row.ProjectID == selectedProjectID))
		}
	}
	return fillLines(s.palette, lines, width, height)
}

func (s SettingsScreen) tableHeader(width int, checkColumns []uikit.ProjectCheck) string {
	cells := []components.TableCell{
		{Value: "project", Width: settingsProjectColumnWidth, Foreground: s.palette.Hint, Bold: true},
	}
	for _, check := range checkColumns {
		cells = append(cells, components.TableCell{
			Value:      check.Title,
			Width:      settingsCheckColumnWidth,
			Foreground: s.palette.Hint,
			Bold:       true,
		})
	}
	return components.RenderTableRow(s.palette, s.palette.Background, width, cells)
}

func (s SettingsScreen) renderRow(width int, checkColumns []uikit.ProjectCheck, row uikit.ProjectCheckRow, selected bool) string {
	background := s.palette.Background
	if selected {
		background = s.palette.Hover
	}
	iconColor := lipgloss.Color(uikit.NormalizeHexColor(row.ProjectColor))
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
		cells = append(cells, components.TableCell{
			Value:      s.checkSymbol(state),
			Width:      settingsCheckColumnWidth,
			Foreground: s.checkColor(state),
		})
	}
	return components.RenderTableRow(s.palette, background, width, cells)
}

func (s SettingsScreen) checkSymbol(state uikit.CheckState) string {
	switch state {
	case uikit.CheckStatePass:
		return uikit.SymbolCheckPass
	case uikit.CheckStateFail:
		return uikit.SymbolCheckFail
	default:
		return uikit.SymbolCheckUnknown
	}
}

func (s SettingsScreen) checkColor(state uikit.CheckState) lipgloss.Color {
	switch state {
	case uikit.CheckStatePass:
		return s.palette.Primary
	case uikit.CheckStateFail:
		return s.palette.Error
	default:
		return s.palette.Hint
	}
}

func (s SettingsScreen) progressLine(width int, status uikit.SettingsStatus) string {
	return components.RenderProgress(s.palette, width, uikit.SymbolSettings, status.Message, status.Error, status.Running, status.Current, status.Total)
}
