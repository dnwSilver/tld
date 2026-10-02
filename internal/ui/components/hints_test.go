package components

import (
	"strings"
	"testing"

	"github.com/dnwSilver/tld/internal/ui/uikit"
)

func TestVulnerabilityHintsIncludeModeSwitch(t *testing.T) {
	hints := NewHints(uikit.NewPalette()).Render(uikit.ScreenVulnerabilities)
	if !strings.Contains(hints, "[m] prod/dev") {
		t.Fatalf("vulnerability hints do not include mode switch: %q", hints)
	}
}

func TestSettingsHintsIncludeOperations(t *testing.T) {
	hints := NewHints(uikit.NewPalette()).Render(uikit.ScreenSettings)
	if !strings.Contains(hints, "[u] ops") {
		t.Fatalf("settings hints do not include operations: %q", hints)
	}
	if !strings.Contains(hints, "[s] sort") {
		t.Fatalf("settings hints do not include sorting: %q", hints)
	}
	if !strings.Contains(hints, "[Shift+←/→] field") {
		t.Fatalf("settings hints do not include column navigation: %q", hints)
	}
}

func TestHintsOnlyShowActionsAvailableOnScreen(t *testing.T) {
	h := NewHints(uikit.NewPalette())
	for _, screen := range []uikit.Screen{uikit.ScreenView, uikit.ScreenReleases} {
		text := h.Render(screen)
		for _, unsupported := range []string{"[a] add", "[e] edit", "[d] delete", "[c] clone"} {
			if strings.Contains(text, unsupported) {
				t.Errorf("screen %v shows %q", screen, unsupported)
			}
		}
	}
	if strings.Contains(h.Render(uikit.ScreenPolicies), "[c] clone") {
		t.Fatal("policies show unsupported clone")
	}
}
