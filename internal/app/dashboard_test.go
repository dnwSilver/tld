package app

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dnwSilver/tld/internal/projectsync"
	"github.com/dnwSilver/tld/internal/storage"
	"github.com/dnwSilver/tld/internal/ui"
)

func TestDependencyVersionLag(t *testing.T) {
	tests := []struct {
		name   string
		actual string
		policy string
		want   versionLag
	}{
		{name: "major", actual: "^1.9.0", policy: "2.0.0", want: versionLagMajor},
		{name: "minor", actual: "1.8.0", policy: "1.9.0", want: versionLagMinorPatch},
		{name: "patch", actual: "1.9.0", policy: "1.9.1", want: versionLagMinorPatch},
		{name: "equal", actual: "1.9.1", policy: "1.9.1", want: versionLagNone},
		{name: "ahead", actual: "3.0.0", policy: "2.9.0", want: versionLagNone},
		{name: "invalid", actual: "latest", policy: "2.9.0", want: versionLagNone},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := dependencyVersionLag(test.actual, test.policy); got != test.want {
				t.Fatalf("dependencyVersionLag(%q, %q) = %v, want %v", test.actual, test.policy, got, test.want)
			}
		})
	}
}

func TestBuildDashboardAttentionRowsCountsFiltersAndSorts(t *testing.T) {
	projects := []storage.Project{
		{ID: 1, Name: "critical-project"},
		{ID: 2, Name: "high-project"},
		{ID: 3, Name: "healthy-project"},
		{ID: 4, Name: "retired-project", EndOfLife: true},
	}
	vulnerabilities := map[int64]projectsync.VulnCounts{
		1: {Critical: 1},
		2: {High: 2},
		4: {Critical: 10},
	}
	checks := map[int64]projectsync.ProjectCheckResults{
		1: {"nightly": projectsync.CheckStateWarning},
		2: {
			"master": projectsync.CheckStateFail,
			"dev":    projectsync.CheckStateFail,
		},
	}
	comparisons := []storage.PolicyVersionComparison{
		{ProjectID: 1, Actual: "1.0.0", Policy: "2.0.0"},
		{ProjectID: 1, Actual: "1.0.0", Policy: "1.1.0"},
		{ProjectID: 3, Actual: "2.0.0", Policy: "1.0.0"},
		{ProjectID: 4, Actual: "1.0.0", Policy: "3.0.0"},
	}

	rows := buildDashboardAttentionRows(projects, vulnerabilities, checks, comparisons)
	if len(rows) != 2 {
		t.Fatalf("len(rows) = %d, want 2: %#v", len(rows), rows)
	}
	if rows[0].ProjectName != "critical-project" {
		t.Fatalf("first project = %q, want critical-project", rows[0].ProjectName)
	}
	if rows[0].Critical != 1 || rows[0].Major != 1 || rows[0].MinorPatch != 1 || rows[0].SettingsWarnings != 1 {
		t.Fatalf("critical project row = %#v", rows[0])
	}
	if rows[1].ProjectName != "high-project" || rows[1].High != 2 || rows[1].SettingsErrors != 2 {
		t.Fatalf("high project row = %#v", rows[1])
	}
}

func TestDashboardPaneAndProjectNavigation(t *testing.T) {
	m := newModel(nil)
	m.dashboard.attentionRows = []ui.DashboardAttentionRow{
		{ProjectID: 1, ProjectName: "first"},
		{ProjectID: 2, ProjectName: "second"},
	}

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRight})
	updated := next.(model)
	if updated.dashboard.focus != ui.DashboardPaneAttention || updated.dashboard.selectedProjectID != 1 {
		t.Fatalf("right key state = focus %v, project %d", updated.dashboard.focus, updated.dashboard.selectedProjectID)
	}

	next, _ = updated.Update(tea.KeyMsg{Type: tea.KeyDown})
	updated = next.(model)
	if updated.dashboard.selectedProjectID != 2 {
		t.Fatalf("selected project = %d, want 2", updated.dashboard.selectedProjectID)
	}

	next, _ = updated.Update(tea.KeyMsg{Type: tea.KeyLeft})
	updated = next.(model)
	if updated.dashboard.focus != ui.DashboardPaneSummary {
		t.Fatalf("left key focus = %v, want summary", updated.dashboard.focus)
	}

	next, _ = updated.Update(tea.KeyMsg{Type: tea.KeyDown})
	if next.(model).dashboard.selectedProjectID != 2 {
		t.Fatal("dashboard summary must not change the selected attention project")
	}

	next, _ = updated.Update(tea.KeyMsg{Type: tea.KeyTab})
	if next.(model).dashboard.focus != ui.DashboardPaneAttention {
		t.Fatalf("tab focus = %v, want attention", next.(model).dashboard.focus)
	}
}
