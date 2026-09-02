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
}
