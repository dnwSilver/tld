package components

import (
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/dnwSilver/tld/internal/ui/uikit"
)

func TestProgressKeepsLargeOutputOnOneTerminalRow(t *testing.T) {
	dump := "trivy failed: panic: scanner crashed\n" + strings.Repeat("goroutine 1 [running]:\n\truntime.goexit()\r\n", 2000)
	for _, width := range []int{0, 1, 8, 24, 48, 80, 160} {
		for _, failed := range []bool{false, true} {
			t.Run(fmt.Sprintf("width=%d/failed=%t", width, failed), func(t *testing.T) {
				message, errText := dump, ""
				if failed {
					message, errText = "Scans finished: 2/3 failed", dump
				}
				line := RenderProgress(uikit.NewPalette(), width, uikit.SymbolVulnerabilities, message, errText, !failed, 3, 3)
				if strings.ContainsAny(ansi.Strip(line), "\r\n") || lipgloss.Width(line) != width {
					t.Fatalf("status size = %dx%d, want %dx1", lipgloss.Width(line), lipgloss.Height(line), width)
				}
				if width >= 48 && !strings.HasSuffix(ansi.Strip(line), "…") {
					t.Fatal("clipped status has no ellipsis")
				}
			})
		}
	}
}

func TestProgressClipsUnicodeByTerminalCells(t *testing.T) {
	line := RenderProgress(uikit.NewPalette(), 13, "", "Ошибка 界界界界界界\x1b[2J\n", "", false, 0, 0)
	if !utf8.ValidString(line) || lipgloss.Width(line) != 13 || strings.ContainsAny(line, "\r\n") || strings.Contains(line, "\x1b[2J") {
		t.Fatalf("invalid clipped status: %q", line)
	}
	if !strings.HasPrefix(ansi.Strip(line), "Ошибка ") || !strings.Contains(line, "…") {
		t.Fatalf("missing message or clipping indicator: %q", line)
	}
}

func TestProgressKeepsShortErrorVisible(t *testing.T) {
	line := RenderProgress(uikit.NewPalette(), 80, "", "Scanning...", "repo: unsupported stack", false, 1, 1)
	if !strings.Contains(line, "repo: unsupported stack") || strings.Contains(line, "…") || lipgloss.Width(line) != 80 {
		t.Fatalf("short error changed: %q", line)
	}
}
