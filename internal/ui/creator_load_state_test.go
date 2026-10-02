package ui

import (
	"strings"
	"testing"

	"github.com/dnwSilver/tld/internal/ui/uikit"
)

func TestCreatorRendersStaleLoadState(t *testing.T) {
	view := NewCreator().Render(RenderState{
		Width: 80, Height: 20, Screen: uikit.ScreenProjects,
		Overlay: OverlayRenderState{LoadState: uikit.LoadState{Phase: uikit.LoadPhaseStale, Error: "timeout\x1b[31m"}},
	})
	if !strings.Contains(view, "STALE: timeout") || !strings.Contains(view, "Ctrl+R") {
		t.Fatalf("stale state is not visible: %q", view)
	}
	if strings.Contains(view, "\x1b[31m") {
		t.Fatal("load error control sequence was not sanitized")
	}
}
