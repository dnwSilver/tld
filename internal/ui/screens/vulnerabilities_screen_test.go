package screens

import (
	"testing"

	"github.com/dnwSilver/tld/internal/ui/uikit"
)

func TestVulnerabilityModeTabsShowSelectedMode(t *testing.T) {
	prod := vulnModeTabs(uikit.VulnModeProd)
	if len(prod) != 2 || prod[0].Label != "prod" || prod[1].Label != "dev" {
		t.Fatalf("prod tabs = %#v", prod)
	}
	if !prod[0].Active || prod[1].Active {
		t.Fatalf("prod active tabs = %#v", prod)
	}

	dev := vulnModeTabs(uikit.VulnModeDev)
	if dev[0].Active || !dev[1].Active {
		t.Fatalf("dev active tabs = %#v", dev)
	}
}
