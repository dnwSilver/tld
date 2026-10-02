package components

import (
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/dnwSilver/tld/internal/ui/uikit"
)

func TestModalOverlayClipsToNarrowViewport(t *testing.T) {
	modal := NewModal(uikit.NewPalette(), uikit.NewPalette().Primary)
	lines := []string{"      ", "      "}
	result := modal.Overlay(lines, 6, 2, "0123456789")
	for _, line := range result {
		if ansi.StringWidth(line) > 6 {
			t.Fatalf("overlay exceeded viewport: %q", line)
		}
	}
}

func TestModalWindowKeepsSelectionVisible(t *testing.T) {
	rows := []string{"0", "1", "2", "3", "4", "5"}
	visible := ModalWindow(rows, 5, 3)
	if len(visible) != 3 || visible[0] != "3" || visible[2] != "5" {
		t.Fatalf("visible = %#v", visible)
	}
}
