package screens

import (
	"testing"

	"github.com/dnwSilver/tld/internal/ui/uikit"
)

func TestSettingsScreenUsesDashForNotApplicableCheck(t *testing.T) {
	screen := NewSettingsScreen(uikit.Palette{})
	if got := screen.checkSymbol(uikit.CheckStateNotApplicable); got != "—" {
		t.Fatalf("not-applicable symbol = %q, want dash", got)
	}
}

func TestSettingsScreenRendersCheckVersionNextToSymbol(t *testing.T) {
	screen := NewSettingsScreen(uikit.Palette{})
	if got := screen.checkValue(uikit.CheckStatePass, "3"); got != uikit.SymbolCheckPass+" 3" {
		t.Fatalf("check value = %q, want symbol and version", got)
	}
}
