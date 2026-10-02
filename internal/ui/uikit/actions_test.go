package uikit

import "testing"

func TestScreenActionRegistryMatchesBindings(t *testing.T) {
	tests := []struct {
		screen Screen
		key    string
		want   ActionID
	}{
		{ScreenProjects, "r", ActionRefreshRow},
		{ScreenProjects, "c", ActionClone},
		{ScreenSettings, "u", ActionOperations},
		{ScreenPolicies, "u", ActionUpdatePins},
		{ScreenVulnerabilities, "m", ActionVulnMode},
	}
	for _, test := range tests {
		got, ok := ScreenActionForKey(test.screen, test.key)
		if !ok || got != test.want {
			t.Errorf("screen %d key %q = %q, %v; want %q", test.screen, test.key, got, ok, test.want)
		}
	}
	if _, ok := ScreenActionForKey(ScreenStacks, "r"); ok {
		t.Fatal("unregistered stack refresh was enabled")
	}
}

func TestGlobalAndModalActionRegistries(t *testing.T) {
	if got, ok := GlobalActionForKey("8"); !ok || got != ActionGoSettings {
		t.Fatalf("global settings action = %q, %v", got, ok)
	}
	if got, ok := GlobalActionForKey("ctrl+r"); !ok || got != ActionReload {
		t.Fatalf("global reload action = %q, %v", got, ok)
	}
	if got, ok := ModalActionForKey(ModalOperations, "enter"); !ok || got != ModalActionConfirm {
		t.Fatalf("modal confirm action = %q, %v", got, ok)
	}
	if got, ok := ModalActionForKey(ModalNavigation, "?"); !ok || got != ModalActionCancel {
		t.Fatalf("modal cancel action = %q, %v", got, ok)
	}
}
