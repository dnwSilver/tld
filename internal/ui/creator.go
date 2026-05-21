package ui

import (
	"github.com/charmbracelet/lipgloss"
)

type Creator struct {
	palette Palette
	hints   Hints
	logo    Logo
	screen  DefaultScreen
}

func NewCreator() Creator {
	palette := NewPalette()

	return Creator{
		palette: palette,
		hints:   NewHints(palette),
		logo:    NewLogo(palette),
		screen:  NewDefaultScreen(palette),
	}
}

func (c Creator) Render(width, height int) string {
	base := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Background(c.palette.Background).
		Foreground(c.palette.Primary)

	top := c.renderTopBar(width)
	topHeight := lipgloss.Height(top)
	bodyHeight := max(height-topHeight, 1)
	body := c.screen.Render(width, bodyHeight)

	return base.Render(lipgloss.JoinVertical(lipgloss.Left, top, body))
}

func (c Creator) renderTopBar(width int) string {
	hints := c.hints.Render()
	logo := c.logo.Render()
	topHeight := max(lipgloss.Height(hints), lipgloss.Height(logo))
	hints = padBlockHeight(c.palette, hints, topHeight)
	logo = padBlockHeight(c.palette, logo, topHeight)
	gap := backgroundBlock(c.palette, max(width-lipgloss.Width(hints)-lipgloss.Width(logo), 0), topHeight)

	return lipgloss.NewStyle().
		Width(width).
		Background(c.palette.Background).
		Render(lipgloss.JoinHorizontal(lipgloss.Top, hints, gap, logo))
}
