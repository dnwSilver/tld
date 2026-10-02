package screens

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/dnwSilver/tld/internal/ui/uikit"
)

func TestProjectsNarrowLayoutShowsFocusedPaneWithinWidth(t *testing.T) {
	screen := NewProjectsScreen(uikit.NewPalette())
	view := screen.Render(48, 16, []uikit.Project{{ID: 1, Name: "repo"}}, 1, []uikit.ProjectDependency{{ID: 2, Name: "dep", Version: "1.0.0"}}, 2, uikit.ProjectDependencyRun{}, uikit.ProjectSyncStatus{}, uikit.ProjectPaneDependencies, nil, nil, nil, uikit.ProjectForm{}, uikit.DeleteConfirm{}, "")
	if !strings.Contains(view, "Dependencies") || strings.Contains(view, "Projects [") {
		t.Fatalf("focused narrow pane not rendered: %q", view)
	}
	for _, line := range strings.Split(view, "\n") {
		if lipgloss.Width(line) > 48 {
			t.Fatalf("line width = %d, want <= 48: %q", lipgloss.Width(line), line)
		}
	}
}

func TestVulnerabilitiesNarrowLayoutShowsDetailsPane(t *testing.T) {
	screen := NewVulnerabilitiesScreen(uikit.NewPalette())
	view := screen.Render(48, 16, []uikit.VulnProjectRow{{ProjectID: 1, ProjectName: "repo", Scanned: true}}, 1, []uikit.VulnerabilityItem{{Title: "CVE-1"}}, 0, uikit.VulnPaneDetails, uikit.VulnModeProd, uikit.SettingsStatus{}, "")
	if !strings.Contains(view, "CVE-1") || strings.Contains(view, "Vulnerabilities [") {
		t.Fatalf("focused narrow pane not rendered: %q", view)
	}
}
