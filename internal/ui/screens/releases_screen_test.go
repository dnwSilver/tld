package screens

import (
	"github.com/dnwSilver/tld/internal/ui/uikit"
	"strings"
	"testing"
)

func TestReleaseMonthUsesAmbulanceForHotfix(t *testing.T) {
	screen := NewReleasesScreen(uikit.Palette{})
	month := uikit.ReleaseMonth{SlotCount: 4, Marks: []bool{true, false, true, false}, HotfixMarks: []bool{false, true, true, false}}
	got := screen.monthCell(month, 4)
	want := uikit.SymbolRocket + uikit.SymbolHotfix + uikit.SymbolHotfix + " "
	if got != want {
		t.Fatalf("marks = %q, want %q", got, want)
	}
	if got := screen.monthCell(month, 1); !strings.Contains(got, uikit.SymbolHotfix) {
		t.Fatalf("compressed marks = %q", got)
	}
}
