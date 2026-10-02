package screens

import (
	"strings"
	"testing"

	"github.com/dnwSilver/tld/internal/ui/uikit"
)

func TestDependenciesScreenKeepsGapsBetweenDynamicColumns(t *testing.T) {
	dependency := uikit.Dependency{
		Icon:               "*",
		Name:               "@spectrum/stylelint-config-recommended",
		RegistrySourceName: "Spectrum",
		StackName:          "JavaScript",
	}
	screen := NewDependenciesScreen(uikit.Palette{})

	got := screen.renderContent(100, 2, []uikit.Dependency{dependency}, 0, nil, nil, uikit.DependencyForm{}, uikit.DeleteConfirm{})

	if !strings.Contains(got, dependency.Name+"  "+dependency.RegistrySourceName+"  "+dependency.StackName) {
		t.Fatalf("dependency columns have no gaps: %q", got)
	}
}
