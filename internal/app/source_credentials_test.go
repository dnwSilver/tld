package app

import (
	"testing"

	"github.com/dnwSilver/tld/internal/storage"
)

func TestSourceCredentialsStayOutsideUIState(t *testing.T) {
	stored := []storage.Source{{ID: 7, Name: "gitlab", PATToken: "secret", URL: "https://gitlab.example", Type: storage.SourceTypeGitLab}}
	visible := toUISources(stored)
	if len(visible) != 1 || visible[0].PATToken != "" {
		t.Fatalf("UI source contains credential: %#v", visible)
	}
	m := newModel(nil)
	m.sources = visible
	m.sourceTokens = sourceTokenMap(stored)
	if got := m.sourceWithCredential(visible[0]).PATToken; got != "secret" {
		t.Fatalf("credential lookup = %q", got)
	}
}
