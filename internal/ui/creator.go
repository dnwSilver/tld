package ui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/dnwSilver/tld/internal/ui/components"
	"github.com/dnwSilver/tld/internal/ui/screens"
	"github.com/dnwSilver/tld/internal/ui/uikit"
)

type Creator struct {
	palette       uikit.Palette
	hints         components.Hints
	logo          components.Logo
	defaultScreen screens.DefaultScreen
	stacksScreen  screens.StacksScreen
}

func NewCreator() Creator {
	palette := uikit.NewPalette()

	return Creator{
		palette:       palette,
		hints:         components.NewHints(palette),
		logo:          components.NewLogo(palette),
		defaultScreen: screens.NewDefaultScreen(palette),
		stacksScreen:  screens.NewStacksScreen(palette),
	}
}

func (c Creator) Render(
	width int,
	height int,
	screen uikit.Screen,
	stacks []uikit.Stack,
	selectedStackID int64,
	stackForm uikit.StackForm,
	deleteConfirm uikit.DeleteConfirm,
) string {
	base := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Background(c.palette.Background).
		Foreground(c.palette.Primary)

	top := c.renderTopBar(width, screen)
	topHeight := lipgloss.Height(top)
	bodyHeight := uikit.Max(height-topHeight, 1)
	body := c.defaultScreen.Render(width, bodyHeight, len(stacks))
	if screen == uikit.ScreenStacks {
		body = c.stacksScreen.Render(width, bodyHeight, stacks, selectedStackID, stackForm, deleteConfirm)
	}

	return base.Render(lipgloss.JoinVertical(lipgloss.Left, top, body))
}

func (c Creator) renderTopBar(width int, screen uikit.Screen) string {
	hints := c.hints.Render(screen)
	logo := c.logo.Render()
	topHeight := uikit.Max(lipgloss.Height(hints), lipgloss.Height(logo))
	hints = uikit.PadBlockHeight(c.palette, hints, topHeight)
	logo = uikit.PadBlockHeight(c.palette, logo, topHeight)
	gap := uikit.BackgroundBlock(c.palette, uikit.Max(width-lipgloss.Width(hints)-lipgloss.Width(logo), 0), topHeight)

	return lipgloss.NewStyle().
		Width(width).
		Background(c.palette.Background).
		Render(lipgloss.JoinHorizontal(lipgloss.Top, hints, gap, logo))
}
