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

func TestReleaseCountColumnsBeforeMonths(t *testing.T) {
	screen := NewReleasesScreen(uikit.Palette{})
	row := uikit.ReleaseRow{ProjectName: "repo", ReleaseCount: 12, HotfixCount: 3, Months: []uikit.ReleaseMonth{{Label: "Oct 26"}}}
	widths := screen.monthWidths(60, 1)
	if widths[0] != 24 {
		t.Fatalf("month width = %d, want 24", widths[0])
	}
	header := screen.tableHeader(60, []uikit.ReleaseRow{row}, widths)
	headerCells := strings.Split(header, releasesSeparator)
	if len(headerCells) != 5 || headerCells[2] != " "+uikit.SymbolRocket+" " || headerCells[3] != " "+uikit.SymbolHotfix+" " {
		t.Fatalf("header cells: %#v", headerCells)
	}
	rendered := screen.renderRow(60, row, widths, false)
	cells := strings.Split(rendered, releasesSeparator)
	if len(cells) != 5 || cells[2] != "12 " || cells[3] != " 3 " {
		t.Fatalf("row cells: %#v", cells)
	}
}

func TestReleaseCountTextFitsThreeCharacters(t *testing.T) {
	for _, test := range []struct {
		count int
		want  string
	}{{0, " 0 "}, {9, " 9 "}, {12, "12 "}, {999, "999"}, {1000, "99+"}} {
		if got := releaseCountText(test.count); got != test.want {
			t.Errorf("%d: %q, want %q", test.count, got, test.want)
		}
	}
}
