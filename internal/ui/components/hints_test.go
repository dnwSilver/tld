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
