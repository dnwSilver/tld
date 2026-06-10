package screens

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dnwSilver/tld/internal/ui/components"
	"github.com/dnwSilver/tld/internal/ui/uikit"
)

type VulnerabilitiesScreen struct {
	palette uikit.Palette
}

const (
	vulnProjectColumnWidth = 20
	vulnCountColumnWidth   = 5
	vulnItemTitleWidth     = 28
	vulnItemSeverityWidth  = 8
)

func NewVulnerabilitiesScreen(palette uikit.Palette) VulnerabilitiesScreen {
	return VulnerabilitiesScreen{palette: palette}
}

func (s VulnerabilitiesScreen) Render(
	width int,
	height int,
	rows []uikit.VulnProjectRow,
	selectedProjectID int64,
	items []uikit.VulnerabilityItem,
	selectedItemIndex int,
	focus uikit.VulnPane,
	status uikit.SettingsStatus,
) string {
	boxHeight := uikit.Max(height-1, 3)
	contentHeight := uikit.Max(boxHeight-2, 1)
	leftTotalWidth := uikit.Max(width/3, 2)
	rightTotalWidth := uikit.Max(width-leftTotalWidth, 2)
	leftContentWidth := uikit.Max(leftTotalWidth-2, 1)
	rightContentWidth := uikit.Max(rightTotalWidth-2, 1)

	count := len(rows)
	leftTitle := components.ScreenTitle(s.palette, uikit.SymbolVulnerabilities, "Vulnerabilities", &count)
	leftBorder := s.borderForPane(focus == uikit.VulnPaneProjects)
	left := components.NewBox(s.palette, leftBorder).Render(leftContentWidth, contentHeight, leftTitle, s.renderProjectsContent(leftContentWidth, contentHeight, rows, selectedProjectID))
	rightTitle := components.ScreenTitle(s.palette, uikit.SymbolVulnHigh, "CVE", intPtr(len(items)))
	rightBorder := s.borderForPane(focus == uikit.VulnPaneDetails)
	right := components.NewBox(s.palette, rightBorder).Render(rightContentWidth, contentHeight, rightTitle, s.renderDetailsContent(rightContentWidth, contentHeight, items, selectedItemIndex))
	body := lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	footer := components.ScreenFooter(s.palette, width, "Kolosov Aleksandr")
	if status.Message != "" || status.Error != "" {
		footer = s.progressLine(width, status)
	}

	return lipgloss.JoinVertical(lipgloss.Left, body, footer)
}

func (s VulnerabilitiesScreen) borderForPane(active bool) lipgloss.Color {
	if active {
		return s.palette.Primary
	}
	return s.palette.Hint
}

func (s VulnerabilitiesScreen) renderProjectsContent(width, height int, rows []uikit.VulnProjectRow, selectedProjectID int64) string {
	lines := make([]string, 0, height)
	lines = append(lines, s.tableHeader(width))
	if len(rows) == 0 {
		empty := uikit.Text(s.palette, s.palette.Hint, "No projects yet")
		lines = append(lines, uikit.CenterLine(s.palette, width, empty))
	} else {
		for _, row := range rows {
			lines = append(lines, s.renderProjectRow(width, row, row.ProjectID == selectedProjectID))
		}
	}
	return fillLines(s.palette, lines, width, height)
}

func (s VulnerabilitiesScreen) tableHeader(width int) string {
	return components.RenderTableRow(s.palette, s.palette.Background, width, []components.TableCell{
		{Value: "project", Width: vulnProjectColumnWidth, Foreground: s.palette.Hint, Bold: true},
		{Value: uikit.SymbolVulnCritical, Width: vulnCountColumnWidth, Foreground: s.palette.Critical, Bold: true},
		{Value: uikit.SymbolVulnHigh, Width: vulnCountColumnWidth, Foreground: s.palette.Error, Bold: true},
		{Value: uikit.SymbolVulnMedium, Width: vulnCountColumnWidth, Foreground: s.palette.Warning, Bold: true},
		{Value: uikit.SymbolVulnLow, Width: vulnCountColumnWidth, Foreground: s.palette.Primary, Bold: true},
		{Value: uikit.SymbolVulnNone, Width: vulnCountColumnWidth, Foreground: s.palette.Hint, Bold: true},
	})
}

func (s VulnerabilitiesScreen) renderProjectRow(width int, row uikit.VulnProjectRow, selected bool) string {
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

	counts := row.Counts
	if !row.Scanned {
		counts = uikit.VulnCounts{}
	}

	return components.RenderTableRow(s.palette, background, width, []components.TableCell{
		{Value: projectCell, Width: vulnProjectColumnWidth, Foreground: s.palette.Text},
		s.countTableCell(counts.Critical, row.Scanned, s.palette.Critical),
		s.countTableCell(counts.High, row.Scanned, s.palette.Error),
		s.countTableCell(counts.Medium, row.Scanned, s.palette.Warning),
		s.countTableCell(counts.Low, row.Scanned, s.palette.Primary),
		s.countTableCell(counts.None, row.Scanned, s.palette.Hint),
	})
}

func (s VulnerabilitiesScreen) countTableCell(value int, scanned bool, activeColor lipgloss.Color) components.TableCell {
	if !scanned {
		return components.TableCell{
			Value:      "-",
			Width:      vulnCountColumnWidth,
			Foreground: s.palette.Hint,
		}
	}
	color := activeColor
	if value == 0 {
		color = s.palette.Hint
	}
	return components.TableCell{
		Value:      uikit.FormatInt(value),
		Width:      vulnCountColumnWidth,
		Foreground: color,
		Bold:       true,
	}
}

func (s VulnerabilitiesScreen) renderDetailsContent(width, height int, items []uikit.VulnerabilityItem, selectedIndex int) string {
	lines := make([]string, 0, height)
	titleWidth := uikit.Max(width-vulnItemSeverityWidth, 1)
	lines = append(lines, components.RenderTableRow(s.palette, s.palette.Background, width, []components.TableCell{
		{Value: "CVE", Width: titleWidth, Foreground: s.palette.Hint, Bold: true},
		{Value: "severity", Width: vulnItemSeverityWidth, Foreground: s.palette.Hint, Bold: true},
	}))

	if len(items) == 0 {
		empty := uikit.Text(s.palette, s.palette.Hint, "No CVE yet")
		lines = append(lines, uikit.CenterLine(s.palette, width, empty))
	} else {
		for index, item := range visibleVulnItems(items, selectedIndex, height-1) {
			background := s.palette.Background
			if index == selectedIndex {
				background = s.palette.Hover
			}
			title := item.Title
			if title == "" {
				title = item.Package
			}
			lines = append(lines, components.RenderTableRow(s.palette, background, width, []components.TableCell{
				{Value: title, Width: titleWidth, Foreground: s.palette.Text},
				{Value: s.severitySymbol(item.Severity), Width: vulnItemSeverityWidth, Foreground: s.severityColor(item.Severity)},
			}))
			if index == selectedIndex && item.Description != "" {
				lines = append(lines, s.descriptionLine(width, item))
			}
		}
	}

	return fillLines(s.palette, lines, width, height)
}

func visibleVulnItems(items []uikit.VulnerabilityItem, selectedIndex int, visibleRows int) []uikit.VulnerabilityItem {
	if visibleRows <= 0 || len(items) <= visibleRows {
		return items
	}
	start := selectedIndex - visibleRows + 1
	if start < 0 {
		start = 0
	}
	if start+visibleRows > len(items) {
		start = len(items) - visibleRows
	}
	return items[start : start+visibleRows]
}

func (s VulnerabilitiesScreen) descriptionLine(width int, item uikit.VulnerabilityItem) string {
	text := strings.TrimSpace(item.Description)
	if item.Range != "" {
		text = strings.TrimSpace(item.Range + " · " + text)
	}
	if text == "" {
		return uikit.BackgroundSpaces(s.palette, width)
	}
	line := uikit.Text(s.palette, s.palette.Hint, text)
	if lipgloss.Width(line) > width {
		line = lipgloss.NewStyle().
			Background(s.palette.Background).
			Foreground(s.palette.Hint).
			Width(width).
			Render(text)
	}
	return lipgloss.NewStyle().
		Width(width).
		Background(s.palette.Background).
		Render(line)
}

func (s VulnerabilitiesScreen) severitySymbol(severity string) string {
	switch strings.ToLower(strings.TrimSpace(severity)) {
	case "critical":
		return uikit.SymbolVulnCritical
	case "high":
		return uikit.SymbolVulnHigh
	case "medium", "moderate":
		return uikit.SymbolVulnMedium
	case "low":
		return uikit.SymbolVulnLow
	default:
		return uikit.SymbolVulnNone
	}
}

func (s VulnerabilitiesScreen) severityColor(severity string) lipgloss.Color {
	switch strings.ToLower(strings.TrimSpace(severity)) {
	case "critical":
		return s.palette.Critical
	case "high":
		return s.palette.Error
	case "medium", "moderate":
		return s.palette.Warning
	case "low":
		return s.palette.Primary
	default:
		return s.palette.Hint
	}
}

func (s VulnerabilitiesScreen) progressLine(width int, status uikit.SettingsStatus) string {
	return components.RenderProgress(s.palette, width, uikit.SymbolVulnerabilities, status.Message, status.Error, status.Running, status.Current, status.Total)
}
