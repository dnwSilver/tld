package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type Hints struct {
	palette Palette
}

func NewHints(palette Palette) Hints {
	return Hints{
		palette: palette,
	}
}

func (h Hints) Render() string {
	rows := []string{
		h.renderRow(
			h.renderLine(h.palette.Hint, symbolTooNew, "too new", 10),
			h.renderLine(h.palette.Primary, symbolAuth, "Auth", 8),
			h.renderLine(h.palette.Error, symbolNpmPackage, "package", 11),
			h.renderLine(h.palette.Hint, symbolSelectPrev, "select prev", 16),
		),
		h.renderRow(
			h.renderLine(h.palette.Primary, symbolSoGood, "so good", 10),
			h.renderLine(h.palette.Info, symbolNetworkPublic, "WWW", 8),
			h.renderLine(h.palette.Info, symbolDockerImage, "image", 11),
			h.renderLine(h.palette.Hint, symbolSelectNext, "select next", 16),
		),
		h.renderRow(
			h.renderLine(h.palette.Info, symbolBeNice, "be nice", 10),
			h.renderLine(h.palette.Info, symbolSEO, "SEO", 8),
			h.renderLine(h.palette.Info, symbolSite, "site", 11),
			h.renderLine(h.palette.Hint, symbolToggleHead, "toggle head", 16),
		),
		h.renderRow(
			h.renderLine(h.palette.Warning, symbolTooOld, "too old", 10),
			h.renderLine(h.palette.Error, symbolGitLab, "VCS", 8),
			h.renderLine(h.palette.Primary, symbolAPI, "api", 11),
			h.renderLine(h.palette.Hint, symbolQuit, "quit", 16),
		),
	}

	rowWidth := lipgloss.Width(rows[0])
	lines := append([]string{backgroundSpaces(h.palette, rowWidth+2)}, rows...)

	for index, row := range rows {
		lines[index+1] = backgroundSpaces(h.palette, 2) + row
	}

	return strings.Join(lines, symbolLineBreak)
}

func (h Hints) renderRow(columns ...string) string {
	return strings.Join(columns, backgroundSpaces(h.palette, 3))
}

func (h Hints) renderLine(color lipgloss.Color, icon, text string, width int) string {
	line := lipgloss.NewStyle().
		Background(h.palette.Background).
		Foreground(color).
		Render(icon) + backgroundSpaces(h.palette, 1) + lipgloss.NewStyle().
		Background(h.palette.Background).
		Foreground(h.palette.Hint).
		Render(text)

	return line + backgroundSpaces(h.palette, width-lipgloss.Width(line))
}
