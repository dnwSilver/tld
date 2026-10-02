package safetext

import (
	"strings"
	"testing"
)

func TestPlainRemovesTerminalControlsButKeepsUnicode(t *testing.T) {
	got := Plain("Пакет 🚀\x1b]52;c;secret\a\u009b[31m\n")
	if strings.ContainsAny(got, "\x1b\a\n") || strings.ContainsRune(got, '\u009b') {
		t.Fatalf("terminal control remained in %q", got)
	}
	if !strings.Contains(got, "Пакет 🚀") {
		t.Fatalf("Unicode lost: %q", got)
	}
}
