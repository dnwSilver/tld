package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dnwSilver/tld/internal/ui/uikit"
)

type Hints struct {
	palette uikit.Palette
}

type hintSpec struct {
	color lipgloss.Color
	icon  string
	text  string
	width int
}

func NewHints(palette uikit.Palette) Hints {
	return Hints{
		palette: palette,
	}
}

func (h Hints) Render(screen uikit.Screen) string {
	rows := h.defaultRows()
	if isListScreen(screen) {
		rows = h.listRows()
	}

	rowWidth := lipgloss.Width(rows[0])
	lines := make([]string, 0, len(rows)+1)
	lines = append(lines, uikit.BackgroundSpaces(h.palette, rowWidth+2))
	for _, row := range rows {
		lines = append(lines, uikit.BackgroundSpaces(h.palette, 2)+row)
	}

	return strings.Join(lines, uikit.SymbolLineBreak)
}

func isListScreen(screen uikit.Screen) bool {
	return screen == uikit.ScreenStacks ||
		screen == uikit.ScreenNamespaces ||
		screen == uikit.ScreenDependencies ||
		screen == uikit.ScreenProjects ||
		screen == uikit.ScreenSources ||
		screen == uikit.ScreenPolicies
}

func (h Hints) defaultRows() []string {
	return []string{
		h.row(
			hintSpec{h.palette.Hint, uikit.SymbolTooNew, "too new", 10},
			hintSpec{h.palette.Primary, uikit.SymbolAuth, "Auth", 8},
			hintSpec{h.palette.Error, uikit.SymbolNpmPackage, "package", 11},
			h.fromBinding(h.palette.Hint, uikit.KeyHome, 16),
		),
		h.row(
			hintSpec{h.palette.Primary, uikit.SymbolSoGood, "so good", 10},
			hintSpec{h.palette.Info, uikit.SymbolNetworkPublic, "WWW", 8},
			hintSpec{h.palette.Info, uikit.SymbolDockerImage, "image", 11},
			h.fromBinding(h.palette.Hint, uikit.KeyStacks, 16),
		),
		h.row(
			hintSpec{h.palette.Info, uikit.SymbolBeNice, "be nice", 10},
			hintSpec{h.palette.Info, uikit.SymbolSEO, "SEO", 8},
			hintSpec{h.palette.Info, uikit.SymbolSite, "site", 11},
			h.fromBinding(h.palette.Hint, uikit.KeyToggleHead, 16),
		),
		h.row(
			hintSpec{h.palette.Primary, uikit.SymbolTooOld, "too old", 10},
			hintSpec{h.palette.Error, uikit.SymbolGitLab, "VCS", 8},
			hintSpec{h.palette.Primary, uikit.SymbolAPI, "api", 11},
			h.fromBinding(h.palette.Hint, uikit.KeyQuit, 16),
		),
	}
}

func (h Hints) listRows() []string {
	return []string{
		h.row(h.fromBinding(h.palette.Hint, uikit.KeyHome, 18), h.fromBinding(h.palette.Primary, uikit.KeyAdd, 20)),
		h.row(h.fromBinding(h.palette.Hint, uikit.KeyStacks, 18), h.fromBinding(h.palette.Info, uikit.KeyEdit, 20)),
		h.row(h.fromBinding(h.palette.Hint, uikit.KeyNamespaces, 18), h.fromBinding(h.palette.Error, uikit.KeyDelete, 20)),
		h.row(h.fromBinding(h.palette.Hint, uikit.KeyDependencies, 18), h.fromBinding(h.palette.Hint, uikit.KeyPrev, 20)),
		h.row(h.fromBinding(h.palette.Hint, uikit.KeyProjects, 18), h.fromBinding(h.palette.Hint, uikit.KeyNext, 20)),
		h.row(h.fromBinding(h.palette.Hint, uikit.KeySources, 18), hintSpec{h.palette.Hint, "", "", 20}),
		h.row(h.fromBinding(h.palette.Hint, uikit.KeyPolicies, 18), h.fromBinding(h.palette.Primary, uikit.KeyRefreshDeps, 20)),
	}
}

func (h Hints) fromBinding(color lipgloss.Color, binding uikit.Binding, width int) hintSpec {
	return hintSpec{color: color, icon: binding.Symbol, text: binding.Hint, width: width}
}

func (h Hints) row(specs ...hintSpec) string {
	parts := make([]string, 0, len(specs))
	for _, spec := range specs {
		parts = append(parts, NewHint(h.palette, spec.color, spec.icon, spec.text, spec.width).Render())
	}

	return strings.Join(parts, uikit.BackgroundSpaces(h.palette, 3))
}
