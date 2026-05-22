package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dnwSilver/tld/internal/ui/uikit"
)

type Hints struct {
	palette uikit.Palette
}

func NewHints(palette uikit.Palette) Hints {
	return Hints{
		palette: palette,
	}
}

func (h Hints) Render(screen uikit.Screen) string {
	rows := []string{
		h.renderRow(
			h.renderHint(h.palette.Hint, uikit.SymbolTooNew, "too new", 10),
			h.renderHint(h.palette.Primary, uikit.SymbolAuth, "Auth", 8),
			h.renderHint(h.palette.Error, uikit.SymbolNpmPackage, "package", 11),
			h.renderHint(h.palette.Hint, uikit.SymbolDashboard, "0 home", 16),
		),
		h.renderRow(
			h.renderHint(h.palette.Primary, uikit.SymbolSoGood, "so good", 10),
			h.renderHint(h.palette.Info, uikit.SymbolNetworkPublic, "WWW", 8),
			h.renderHint(h.palette.Info, uikit.SymbolDockerImage, "image", 11),
			h.renderHint(h.palette.Hint, uikit.SymbolStack, "1 stacks", 16),
		),
		h.renderRow(
			h.renderHint(h.palette.Info, uikit.SymbolBeNice, "be nice", 10),
			h.renderHint(h.palette.Info, uikit.SymbolSEO, "SEO", 8),
			h.renderHint(h.palette.Info, uikit.SymbolSite, "site", 11),
			h.renderHint(h.palette.Hint, uikit.SymbolToggleHead, "toggle head", 16),
		),
		h.renderRow(
			h.renderHint(h.palette.Warning, uikit.SymbolTooOld, "too old", 10),
			h.renderHint(h.palette.Error, uikit.SymbolGitLab, "VCS", 8),
			h.renderHint(h.palette.Primary, uikit.SymbolAPI, "api", 11),
			h.renderHint(h.palette.Hint, uikit.SymbolQuit, "quit", 16),
		),
	}
	if screen == uikit.ScreenStacks || screen == uikit.ScreenNamespaces || screen == uikit.ScreenDependencies {
		rows = []string{
			h.renderRow(
				h.renderHint(h.palette.Hint, uikit.SymbolDashboard, "0 home", 18),
				h.renderHint(h.palette.Primary, uikit.SymbolAdd, "[a] add", 20),
			),
			h.renderRow(
				h.renderHint(h.palette.Hint, uikit.SymbolStack, "1 stacks", 18),
				h.renderHint(h.palette.Info, uikit.SymbolEdit, "[e] edit", 20),
			),
			h.renderRow(
				h.renderHint(h.palette.Hint, uikit.SymbolNamespace, "2 namespaces", 18),
				h.renderHint(h.palette.Error, uikit.SymbolDelete, "[d] delete", 20),
			),
			h.renderRow(
				h.renderHint(h.palette.Hint, uikit.SymbolDependency, "3 deps", 18),
				h.renderHint(h.palette.Hint, uikit.SymbolSelectPrev, "[h] prev", 20),
			),
			h.renderRow(
				h.renderHint(h.palette.Hint, "", "", 18),
				h.renderHint(h.palette.Hint, uikit.SymbolSelectNext, "[j] next", 20),
			),
		}
	}

	rowWidth := lipgloss.Width(rows[0])
	lines := append([]string{uikit.BackgroundSpaces(h.palette, rowWidth+2)}, rows...)

	for index, row := range rows {
		lines[index+1] = uikit.BackgroundSpaces(h.palette, 2) + row
	}

	return strings.Join(lines, uikit.SymbolLineBreak)
}

func (h Hints) renderRow(columns ...string) string {
	return strings.Join(columns, uikit.BackgroundSpaces(h.palette, 3))
}

func (h Hints) renderHint(color lipgloss.Color, icon, text string, width int) string {
	return NewHint(h.palette, color, icon, text, width).Render()
}
