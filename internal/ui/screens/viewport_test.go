package screens

import (
	"strconv"
	"testing"

	"github.com/dnwSilver/tld/internal/ui/uikit"
)

func TestVisibleRowsByIDKeepsSelectionVisible(t *testing.T) {
	rows := []int64{1, 2, 3, 4, 5, 6}
	got := visibleRowsByID(rows, 5, 3, func(id int64) int64 { return id })
	if len(got) != 3 || got[0] != 3 || got[2] != 5 {
		t.Fatalf("visible rows = %v", got)
	}
}

func BenchmarkProjectsRenderViewport(b *testing.B) {
	for _, count := range []int{100, 1000, 5000} {
		b.Run(strconv.Itoa(count), func(b *testing.B) {
			rows := make([]uikit.Project, count)
			for index := range rows {
				rows[index] = uikit.Project{ID: int64(index + 1), Name: "project-" + strconv.Itoa(index), NamespaceName: "namespace", SourceName: "source"}
			}
			screen := NewProjectsScreen(uikit.NewPalette())
			b.ReportAllocs()
			for index := 0; index < b.N; index++ {
				_ = screen.renderProjectsContent(110, 20, rows, int64(count))
			}
		})
	}
}
