package screens

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dnwSilver/tld/internal/ui/components"
	"github.com/dnwSilver/tld/internal/ui/uikit"
)

type ReleasesScreen struct {
	palette uikit.Palette
}

const (
	releasesProjectColumnWidth = 24
	releasesSeparator          = "│"
)

func NewReleasesScreen(palette uikit.Palette) ReleasesScreen {
	return ReleasesScreen{palette: palette}
}

func (s ReleasesScreen) Render(
	width int,
	height int,
	rows []uikit.ReleaseRow,
	selectedProjectID int64,
	period uikit.ReleasePeriod,
	status uikit.SettingsStatus,
) string {
	boxHeight := uikit.Max(height-1, 3)
	contentHeight := uikit.Max(boxHeight-2, 1)
	contentWidth := uikit.Max(width-2, 1)
	count := len(rows)
	title := components.ScreenTitle(s.palette, uikit.SymbolReleases, "Releases · "+period.Title(), &count)
	body := s.renderContent(contentWidth, contentHeight, rows, selectedProjectID, period)
	box := components.NewBox(s.palette, s.palette.Primary).Render(contentWidth, contentHeight, title, body)
	footer := components.ScreenFooter(s.palette, width, "Kolosov Aleksandr")
	if status.Message != "" || status.Error != "" {
		footer = s.progressLine(width, status)
	}

	return lipgloss.JoinVertical(lipgloss.Left, box, footer)
}

func (s ReleasesScreen) renderContent(
	width int,
	height int,
	rows []uikit.ReleaseRow,
	selectedProjectID int64,
	period uikit.ReleasePeriod,
) string {
	monthWidths := s.monthWidths(width, period.Months())
	lines := make([]string, 0, height)
	lines = append(lines, s.tableHeader(width, rows, monthWidths))
	if len(rows) == 0 {
		empty := uikit.Text(s.palette, s.palette.Hint, "No projects yet")
		lines = append(lines, uikit.CenterLine(s.palette, width, empty))
	} else {
		for _, row := range rows {
			lines = append(lines, s.renderRow(width, row, monthWidths, row.ProjectID == selectedProjectID))
		}
	}

	return fillLines(s.palette, lines, width, height)
}

func (s ReleasesScreen) monthWidths(width int, monthCount int) []int {
	widths := make([]int, monthCount)
	available := uikit.Max(width-releasesProjectColumnWidth-monthCount, monthCount)
	base := available / monthCount
	remainder := available % monthCount
	for index := range widths {
		widths[index] = base
		if index >= monthCount-remainder {
			widths[index]++
		}
	}

	return widths
}

func (s ReleasesScreen) tableHeader(width int, rows []uikit.ReleaseRow, monthWidths []int) string {
	labels := s.headerMonths(rows, len(monthWidths))
	cells := []components.TableCell{
		{Value: "project", Width: releasesProjectColumnWidth, Foreground: s.palette.Hint, Bold: true},
	}
	for index, monthWidth := range monthWidths {
		cells = append(cells, s.separatorCell())
		cells = append(cells, components.TableCell{
			Value:      centerText(labels[index], monthWidth),
			Width:      monthWidth,
			Foreground: s.palette.Hint,
			Bold:       true,
		})
	}

	return components.RenderTableRow(s.palette, s.palette.Background, width, cells)
}

func (s ReleasesScreen) headerMonths(rows []uikit.ReleaseRow, monthCount int) []string {
	labels := make([]string, monthCount)
	if len(rows) == 0 {
		return labels
	}
	for index, month := range rows[0].Months {
		if index < len(labels) {
			labels[index] = month.Label
		}
	}

	return labels
}

func (s ReleasesScreen) renderRow(width int, row uikit.ReleaseRow, monthWidths []int, selected bool) string {
	background := s.palette.Background
	if selected {
		background = s.palette.Hover
	}
	iconColor := lipgloss.Color(uikit.NormalizeHexColor(row.ProjectColor))
	projectColor := s.palette.Text
	if !row.HasReleases {
		iconColor = s.palette.Hint
		projectColor = s.palette.Hint
	}
	projectCell := lipgloss.NewStyle().
		Background(background).
		Foreground(iconColor).
		Width(2).
		Render(row.ProjectIcon) + lipgloss.NewStyle().
		Background(background).
		Foreground(projectColor).
		Render(row.ProjectName)

	cells := []components.TableCell{
		{Value: projectCell, Width: releasesProjectColumnWidth, Foreground: projectColor},
	}
	for index, monthWidth := range monthWidths {
		cells = append(cells, s.separatorCell())
		value := uikit.BackgroundSpaces(s.palette, monthWidth)
		if row.HasReleases && index < len(row.Months) {
			value = s.monthCell(row.Months[index], monthWidth)
		}
		cells = append(cells, components.TableCell{
			Value:      value,
			Width:      monthWidth,
			Foreground: iconColor,
		})
	}

	return components.RenderTableRow(s.palette, background, width, cells)
}

func (s ReleasesScreen) separatorCell() components.TableCell {
	return components.TableCell{
		Value:      releasesSeparator,
		Width:      1,
		Foreground: s.palette.Disable,
	}
}

func (s ReleasesScreen) monthCell(month uikit.ReleaseMonth, monthWidth int) string {
	if monthWidth <= 0 || month.SlotCount <= 0 {
		return uikit.BackgroundSpaces(s.palette, monthWidth)
	}
	chars := make([]string, monthWidth)
	for index := range chars {
		chars[index] = " "
	}
	for slot, hasRelease := range month.Marks {
		if !hasRelease {
			continue
		}
		position := slot * monthWidth / month.SlotCount
		if position >= monthWidth {
			position = monthWidth - 1
		}
		chars[position] = uikit.SymbolRocket
	}

	return strings.Join(chars, "")
}

func (s ReleasesScreen) progressLine(width int, status uikit.SettingsStatus) string {
	return components.RenderProgress(s.palette, width, uikit.SymbolReleases, status.Message, status.Error, status.Running, status.Current, status.Total)
}

func centerText(value string, width int) string {
	if width <= 0 {
		return ""
	}
	valueWidth := lipgloss.Width(value)
	if valueWidth >= width {
		return value
	}
	left := (width - valueWidth) / 2
	right := width - valueWidth - left

	return strings.Repeat(" ", left) + value + strings.Repeat(" ", right)
}
