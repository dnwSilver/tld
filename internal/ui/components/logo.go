package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dnwSilver/tld/internal/ui/uikit"
)

type Logo struct {
	palette uikit.Palette
}

func NewLogo(palette uikit.Palette) Logo {
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
	}, uikit.SymbolLineBreak)

	return lipgloss.NewStyle().
		Background(l.palette.Background).
		Padding(1, 2, 0, 0).
		Align(lipgloss.Right).
		Render(output)
}
