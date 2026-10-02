package app

import (
	"errors"
	"testing"

	"github.com/dnwSilver/tld/internal/ui"
)

func TestLoadStateKeepsLastGoodDataAsStale(t *testing.T) {
	m := newModel(nil)
	m.projects = []ui.Project{{ID: 1, Name: "existing"}}
	m.loadStates[ui.ScreenProjects] = ui.LoadState{Phase: ui.LoadPhaseReady}
	m.beginLoad(ui.ScreenProjects)
	m.finishLoad(ui.ScreenProjects, errors.New("provider timeout"))

	state := m.loadState(ui.ScreenProjects)
	if state.Phase != ui.LoadPhaseStale || state.Error != "provider timeout" {
		t.Fatalf("unexpected load state: %#v", state)
	}
	if len(m.projects) != 1 {
		t.Fatal("last-good projects were discarded")
	}
}

func TestLoadStateWithoutSnapshotIsError(t *testing.T) {
	m := newModel(nil)
	m.beginLoad(ui.ScreenReleases)
	m.finishLoad(ui.ScreenReleases, errors.New("offline"))
	if got := m.loadState(ui.ScreenReleases).Phase; got != ui.LoadPhaseError {
		t.Fatalf("phase = %q, want error", got)
	}
}

func TestMultiLoaderKeepsFailureAfterAnotherLoaderSucceeds(t *testing.T) {
	m := newModel(nil)
	m.beginLoads(ui.ScreenDefault, 2)
	m.finishLoad(ui.ScreenDefault, errors.New("rights denied"))
	if got := m.loadState(ui.ScreenDefault).Phase; got != ui.LoadPhaseLoading {
		t.Fatalf("phase after first reply = %q", got)
	}
	m.finishLoad(ui.ScreenDefault, nil)
	state := m.loadState(ui.ScreenDefault)
	if state.Phase != ui.LoadPhaseError || state.Error != "rights denied" {
		t.Fatalf("final state = %#v", state)
	}
}

func TestEmptyLastGoodSnapshotBecomesStale(t *testing.T) {
	m := newModel(nil)
	m.loadStates[ui.ScreenReleases] = ui.LoadState{Phase: ui.LoadPhaseReady, HasSnapshot: true}
	m.beginLoad(ui.ScreenReleases)
	m.finishLoad(ui.ScreenReleases, errors.New("offline"))
	if got := m.loadState(ui.ScreenReleases).Phase; got != ui.LoadPhaseStale {
		t.Fatalf("phase = %q, want stale", got)
	}
}
