package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dnwSilver/tld/internal/ui/uikit"
)

type Hints struct {
	palette uikit.Palette
}

const maxHintRows = 5

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
	if screen == uikit.ScreenVulnerabilities {
		rows = h.vulnerabilityRows()
	}
	if screen == uikit.ScreenPolicies {
		rows = h.policyRows()
	}
	if screen == uikit.ScreenSettings {
		rows = h.settingsRows()
	}

	rowWidth := maxRowWidth(rows)
	columnRows := uikit.Max(maxHintRows-1, 1)
	columns := make([][]string, 0, (len(rows)+columnRows-1)/columnRows)
	for start := 0; start < len(rows); start += columnRows {
		end := uikit.Min(start+columnRows, len(rows))
		column := []string{uikit.BackgroundSpaces(h.palette, rowWidth+2)}
		for _, row := range rows[start:end] {
			column = append(column, h.paddedRow(row, rowWidth))
		}
		for len(column) < maxHintRows {
			column = append(column, uikit.BackgroundSpaces(h.palette, rowWidth+2))
		}
		columns = append(columns, column)
	}

	lines := make([]string, 0, maxHintRows)
	gap := uikit.BackgroundSpaces(h.palette, 3)
	for rowIndex := 0; rowIndex < maxHintRows; rowIndex++ {
		parts := make([]string, 0, len(columns))
		for _, column := range columns {
			parts = append(parts, column[rowIndex])
		}
		lines = append(lines, strings.Join(parts, gap))
	}

	return strings.Join(lines, uikit.SymbolLineBreak)
}

func isListScreen(screen uikit.Screen) bool {
	return screen == uikit.ScreenStacks ||
		screen == uikit.ScreenNamespaces ||
		screen == uikit.ScreenDependencies ||
		screen == uikit.ScreenProjects ||
		screen == uikit.ScreenSources ||
		screen == uikit.ScreenPolicies ||
		screen == uikit.ScreenView ||
		screen == uikit.ScreenSettings ||
		screen == uikit.ScreenReleases ||
		screen == uikit.ScreenVulnerabilities
}

func (h Hints) defaultRows() []string {
	return []string{
		h.row(
			hintSpec{h.palette.Hint, uikit.SymbolTooNew, "too new", 10},
			hintSpec{h.palette.Primary, uikit.SymbolAuth, "Auth", 8},
			hintSpec{h.palette.Error, uikit.SymbolNpmPackage, "package", 11},
			h.fromBinding(h.palette.Hint, uikit.KeyToggleHead, 16),
		),
		h.row(
			hintSpec{h.palette.Primary, uikit.SymbolSoGood, "so good", 10},
			hintSpec{h.palette.Info, uikit.SymbolNetworkPublic, "WWW", 8},
			hintSpec{h.palette.Info, uikit.SymbolDockerImage, "image", 11},
			hintSpec{h.palette.Hint, "", "", 16},
		),
		h.row(
			hintSpec{h.palette.Info, uikit.SymbolBeNice, "be nice", 10},
			hintSpec{h.palette.Info, uikit.SymbolSEO, "SEO", 8},
			hintSpec{h.palette.Info, uikit.SymbolSite, "site", 11},
			hintSpec{h.palette.Hint, "", "", 16},
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
		h.row(h.fromBinding(h.palette.Primary, uikit.KeyAdd, 20), h.fromBinding(h.palette.Info, uikit.KeyEdit, 20)),
		h.row(h.fromBinding(h.palette.Error, uikit.KeyDelete, 20), h.fromBinding(h.palette.Hint, uikit.KeyPrev, 20)),
		h.row(h.fromBinding(h.palette.Info, uikit.KeyClone, 20), h.fromBinding(h.palette.Primary, uikit.KeyRefreshDeps, 20)),
		h.row(h.fromBinding(h.palette.Primary, uikit.KeyRefreshAll, 20), h.fromBinding(h.palette.Hint, uikit.KeyToggleHead, 20)),
	}
}

func (h Hints) vulnerabilityRows() []string {
	return []string{
		h.row(h.fromBinding(h.palette.Hint, uikit.KeyPrev, 20), h.fromBinding(h.palette.Hint, uikit.KeyNext, 20)),
		h.row(h.fromBinding(h.palette.Primary, uikit.KeyRefreshDeps, 20), h.fromBinding(h.palette.Primary, uikit.KeyRefreshAll, 20)),
		h.row(h.fromBinding(h.palette.Info, uikit.KeyVulnMode, 20), hintSpec{h.palette.Hint, uikit.SymbolSelectNext, "[Tab] pane", 20}),
		h.row(h.fromBinding(h.palette.Hint, uikit.KeyToggleHead, 20), h.fromBinding(h.palette.Hint, uikit.KeyQuit, 20)),
	}
}

func (h Hints) settingsRows() []string {
	return []string{
		h.row(h.fromBinding(h.palette.Hint, uikit.KeyPrev, 20), h.fromBinding(h.palette.Hint, uikit.KeyNext, 20)),
		h.row(h.fromBinding(h.palette.Primary, uikit.KeyRefreshDeps, 20), h.fromBinding(h.palette.Primary, uikit.KeyRefreshAll, 20)),
		h.row(h.fromBinding(h.palette.Info, uikit.KeyOperations, 20), h.fromBinding(h.palette.Hint, uikit.KeyToggleHead, 20)),
		h.row(h.fromBinding(h.palette.Hint, uikit.KeyQuit, 20), hintSpec{h.palette.Hint, "", "", 20}),
	}
}

func (h Hints) policyRows() []string {
	return []string{
		h.row(h.fromBinding(h.palette.Primary, uikit.KeyAdd, 20), h.fromBinding(h.palette.Info, uikit.KeyEdit, 20)),
		h.row(h.fromBinding(h.palette.Error, uikit.KeyDelete, 20), h.fromBinding(h.palette.Hint, uikit.KeyPrev, 20)),
		h.row(h.fromBinding(h.palette.Info, uikit.KeyClone, 20), h.fromBinding(h.palette.Primary, uikit.KeyUpdatePins, 20)),
		h.row(h.fromBinding(h.palette.Hint, uikit.KeyToggleHead, 20), h.fromBinding(h.palette.Hint, uikit.KeyQuit, 20)),
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

func (h Hints) paddedRow(row string, rowWidth int) string {
	return uikit.BackgroundSpaces(h.palette, 2) + row + uikit.BackgroundSpaces(h.palette, rowWidth-lipgloss.Width(row))
}

func maxRowWidth(rows []string) int {
	width := 0
	for _, row := range rows {
		width = uikit.Max(width, lipgloss.Width(row))
	}

	return width
}
