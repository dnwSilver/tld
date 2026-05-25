package screens

import (
	"testing"

	"github.com/dnwSilver/tld/internal/ui/uikit"
)

func TestDependencyViewVersionColor(t *testing.T) {
	palette := uikit.NewPalette()
	screen := NewDependencyViewScreen(palette)

	tests := []struct {
		name   string
		actual string
		policy string
		want   string
	}{
		{name: "major ahead", actual: "20.0.0", policy: "19.2.0", want: string(palette.Info)},
		{name: "behind same major", actual: "19.1.0", policy: "19.2.0", want: string(palette.Primary)},
		{name: "behind by more than one major", actual: "17.9.0", policy: "19.2.0", want: string(palette.Warning)},
		{name: "equal", actual: "19.2.0", policy: "19.2.0", want: string(palette.Hint)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := screen.versionColor(tt.actual, tt.policy)
			if string(got) != tt.want {
				t.Fatalf("color = %q, want %q", got, tt.want)
			}
		})
	}
}
