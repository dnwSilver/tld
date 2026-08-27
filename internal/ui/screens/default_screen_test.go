package screens

import (
	"strings"
	"testing"

	"github.com/dnwSilver/tld/internal/ui/uikit"
)

func TestDefaultScreenRendersAttentionTable(t *testing.T) {
	screen := NewDefaultScreen(uikit.Palette{})
	content := screen.Render(120, 14, 4, []uikit.DashboardAttentionRow{
		{
			ProjectID:        1,
			ProjectName:      "payments",
			Critical:         1,
			High:             2,
			Major:            3,
			MinorPatch:       4,
			SettingsErrors:   5,
			SettingsWarnings: 6,
		},
	}, uikit.DashboardPaneAttention, 1)

	for _, expected := range []string{"Team lead dashboard", "Stacks: 4", "Attention [1]", "project", "crit", "high", "major", "min/patch", "errors", "warnings", "payments"} {
		if !strings.Contains(content, expected) {
			t.Fatalf("content does not contain %q:\n%s", expected, content)
		}
	}
}

func TestDefaultScreenRendersEmptyAttentionState(t *testing.T) {
	screen := NewDefaultScreen(uikit.Palette{})
	content := screen.Render(100, 10, 0, nil, uikit.DashboardPaneSummary, 0)

	if !strings.Contains(content, "No projects require attention") {
		t.Fatalf("unexpected empty state:\n%s", content)
	}
}

func TestVisibleDashboardAttentionRowsFollowsSelection(t *testing.T) {
	rows := []uikit.DashboardAttentionRow{
		{ProjectID: 1},
		{ProjectID: 2},
		{ProjectID: 3},
		{ProjectID: 4},
	}

	visible := visibleDashboardAttentionRows(rows, 4, 2)
	if len(visible) != 2 || visible[0].ProjectID != 3 || visible[1].ProjectID != 4 {
		t.Fatalf("visible rows = %#v", visible)
	}
}
