package screens

import (
	"strings"
	"testing"

	"github.com/dnwSilver/tld/internal/ui/uikit"
)

func TestVulnerabilityEmptyStateDistinguishesUnscannedFromClean(t *testing.T) {
	screen := NewVulnerabilitiesScreen(uikit.NewPalette())
	rows := []uikit.VulnProjectRow{{ProjectID: 1, ProjectName: "repo"}}
	unscanned := screen.Render(100, 20, rows, 1, nil, 0, uikit.VulnPaneProjects, uikit.VulnModeProd, uikit.SettingsStatus{}, "")
	if !strings.Contains(unscanned, "Not scanned") {
		t.Fatal("unscanned project was shown as clean")
	}
	rows[0].Scanned = true
	rows[0].ScannedAt = "2026-10-02 10:00 UTC"
	rows[0].Revision = "abcdef1234567890"
	clean := screen.Render(100, 20, rows, 1, nil, 0, uikit.VulnPaneProjects, uikit.VulnModeProd, uikit.SettingsStatus{}, "")
	if !strings.Contains(clean, "No CVE found") || strings.Contains(clean, "Not scanned") {
		t.Fatal("clean scan was not distinguished from missing scan")
	}
	if !strings.Contains(clean, "2026-10-02 10:00 UTC @abcdef12") {
		t.Fatal("scan provenance was not rendered")
	}
	rows[0].Stale = true
	rows[0].LastOutcome = "failed"
	rows[0].LastAttemptAt = "2026-10-02 11:00 UTC"
	stale := screen.Render(100, 20, rows, 1, nil, 0, uikit.VulnPaneProjects, uikit.VulnModeProd, uikit.SettingsStatus{}, "")
	if !strings.Contains(stale, "STALE") || !strings.Contains(stale, "last attempt failed") {
		t.Fatal("failed refresh did not mark last-good report stale")
	}
}
