package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type Logo struct {
	palette Palette
}

func NewLogo(palette Palette) Logo {
	return Logo{
		palette: palette,
	}
}

func (l Logo) Render() string {
	mainStyle := lipgloss.NewStyle().
		Background(l.palette.Background).
		Foreground(l.palette.Primary)
	hintStyle := lipgloss.NewStyle().
		Background(l.palette.Background).
		Foreground(l.palette.Hint)
	output := strings.Join([]string{
		"  " + mainStyle.Render("╭━━━━┳╮") + hintStyle.Render("╱╱") + mainStyle.Render("╭━━━╮"),
		"  " + mainStyle.Render("┃╭╮╭╮┃┃") + hintStyle.Render("╱╱") + mainStyle.Render("╰╮╭╮┃"),
		"  " + mainStyle.Render("╰╯┃┃╰┫┃") + hintStyle.Render("╱╱╱") + mainStyle.Render("┃┃┃┃"),
		hintStyle.Render("  ╱╱") + mainStyle.Render("┃┃") + hintStyle.Render("╱") + mainStyle.Render("┃┃") + hintStyle.Render("╱") + mainStyle.Render("╭╮┃┃┃┃"),
		hintStyle.Render(" ╱╱╱") + mainStyle.Render("┃┃") + hintStyle.Render("╱") + mainStyle.Render("┃╰━╯┣╯╰╯┃"),
		hintStyle.Render("╱╱╱╱") + mainStyle.Render("╰╯") + hintStyle.Render("╱") + mainStyle.Render("╰━━━┻━━━╯"),
	}, symbolLineBreak)

	return lipgloss.NewStyle().
		Background(l.palette.Background).
		Padding(1, 2, 0, 0).
		Align(lipgloss.Right).
		Render(output)
}
