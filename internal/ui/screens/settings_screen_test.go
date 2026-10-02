package screens

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/dnwSilver/tld/internal/ui/uikit"
)

func TestSettingsScreenUsesDashForNotApplicableCheck(t *testing.T) {
	screen := NewSettingsScreen(uikit.Palette{})
	if got := screen.checkSymbol(uikit.CheckStateNotApplicable); got != "—" {
		t.Fatalf("not-applicable symbol = %q, want dash", got)
	}
}

func TestSettingsColumnsFollowSelectedField(t *testing.T) {
	columns := make([]uikit.ProjectCheck, 14)
	for index := range columns {
		columns[index].ID = uikit.FormatInt(index)
	}
	visible, offset := visibleSettingsColumns(80, columns, 14)
	if len(visible) != 7 || offset != 7 || visible[6].ID != "13" {
		t.Fatalf("visible columns = %v, offset = %d", visible, offset)
	}
}

func TestSettingsScreenRendersCheckVersionNextToSymbol(t *testing.T) {
	screen := NewSettingsScreen(uikit.Palette{})
	if got := screen.checkValue(uikit.CheckStatePass, "3"); got != uikit.SymbolCheckPass+" 3" {
		t.Fatalf("check value = %q, want symbol and version", got)
	}
}

func TestSettingsScreenRendersWarningInWarningColor(t *testing.T) {
	palette := uikit.NewPalette()
	screen := NewSettingsScreen(palette)

	if got := screen.checkSymbol(uikit.CheckStateWarning); got != uikit.SymbolCheckWarning {
		t.Fatalf("warning symbol = %q, want %q", got, uikit.SymbolCheckWarning)
	}
	if got := screen.checkColor(uikit.CheckStateWarning); got != palette.Warning {
		t.Fatalf("warning color = %q, want %q", got, palette.Warning)
	}
}

func TestSettingsScreenMarksSelectedAndSortedHeader(t *testing.T) {
	palette := uikit.NewPalette()
	screen := NewSettingsScreen(palette)
	state := uikit.SettingsTableState{
		SelectedColumn: 1,
		SortColumn:     1,
		SortDescending: true,
		SortActive:     true,
	}

	selected := screen.headerCell("master", settingsCheckColumnWidth, 1, state)
	if selected.Foreground != palette.Primary {
		t.Fatalf("selected header color = %q, want %q", selected.Foreground, palette.Primary)
	}
	if selected.Value != uikit.SymbolSortDescending+"master" {
		t.Fatalf("selected header = %q, want descending marker", selected.Value)
	}

	unselected := screen.headerCell("project", settingsProjectColumnWidth, 0, state)
	if unselected.Foreground != palette.Hint {
		t.Fatalf("unselected header color = %q, want %q", unselected.Foreground, palette.Hint)
	}
}

func TestSettingsScreenUsesGreenProjectIconWhenAllApplicableChecksPass(t *testing.T) {
	palette := uikit.NewPalette()
	screen := NewSettingsScreen(palette)
	checks := []uikit.ProjectCheck{
		{ID: "branch"},
		{ID: "next"},
	}
	results := map[string]uikit.CheckState{
		"branch": uikit.CheckStatePass,
		"next":   uikit.CheckStateNotApplicable,
	}
	defaultColor := lipgloss.Color("#FFFFFF")

	if got := screen.projectIconColor(checks, results, defaultColor); got != palette.Primary {
		t.Fatalf("project icon color = %q, want green %q", got, palette.Primary)
	}
}

func TestSettingsScreenKeepsProjectIconColorWhenChecksAreNotAllSuccessful(t *testing.T) {
	palette := uikit.NewPalette()
	screen := NewSettingsScreen(palette)
	checks := []uikit.ProjectCheck{{ID: "branch"}}
	defaultColor := lipgloss.Color("#FFFFFF")

	for name, results := range map[string]map[string]uikit.CheckState{
		"failed":  {"branch": uikit.CheckStateFail},
		"warning": {"branch": uikit.CheckStateWarning},
		"unknown": {"branch": uikit.CheckStateUnknown},
		"missing": {},
	} {
		t.Run(name, func(t *testing.T) {
			if got := screen.projectIconColor(checks, results, defaultColor); got != defaultColor {
				t.Fatalf("project icon color = %q, want default %q", got, defaultColor)
			}
		})
	}
}
