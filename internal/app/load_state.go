package app

import "github.com/dnwSilver/tld/internal/ui"

func initialLoadStates() map[ui.Screen]ui.LoadState {
	states := make(map[ui.Screen]ui.LoadState)
	for _, screen := range []ui.Screen{
		ui.ScreenDefault,
		ui.ScreenStacks,
		ui.ScreenNamespaces,
		ui.ScreenDependencies,
		ui.ScreenProjects,
		ui.ScreenSources,
		ui.ScreenPolicies,
	} {
		states[screen] = ui.LoadState{Phase: ui.LoadPhaseLoading}
	}
	return states
}

func initialLoadPending() map[ui.Screen]int {
	return map[ui.Screen]int{
		ui.ScreenDefault:      2,
		ui.ScreenStacks:       1,
		ui.ScreenNamespaces:   1,
		ui.ScreenDependencies: 1,
		ui.ScreenProjects:     1,
		ui.ScreenSources:      1,
		ui.ScreenPolicies:     1,
	}
}

func (m *model) beginLoad(screen ui.Screen) {
	m.beginLoads(screen, 1)
}

func (m *model) beginLoads(screen ui.Screen, pending int) {
	if m.loadStates == nil {
		m.loadStates = make(map[ui.Screen]ui.LoadState)
	}
	if m.loadPending == nil {
		m.loadPending = make(map[ui.Screen]int)
	}
	if pending < 1 {
		pending = 1
	}
	previous := m.loadStates[screen]
	m.loadStates[screen] = ui.LoadState{Phase: ui.LoadPhaseLoading, HasSnapshot: previous.HasSnapshot}
	m.loadPending[screen] = pending
}

func (m *model) finishLoad(screen ui.Screen, err error) {
	if m.loadStates == nil {
		m.loadStates = make(map[ui.Screen]ui.LoadState)
	}
	if m.loadPending == nil {
		m.loadPending = make(map[ui.Screen]int)
	}
	previous := m.loadStates[screen]
	pending := m.loadPending[screen]
	if pending > 0 {
		pending--
		m.loadPending[screen] = pending
	}
	if err != nil {
		previous.Error = err.Error()
	}
	if pending > 0 {
		previous.Phase = ui.LoadPhaseLoading
		m.loadStates[screen] = previous
		return
	}
	if previous.Error == "" {
		m.loadStates[screen] = ui.LoadState{Phase: ui.LoadPhaseReady, HasSnapshot: true}
		return
	}
	phase := ui.LoadPhaseError
	if previous.HasSnapshot || previous.Phase == ui.LoadPhaseReady || previous.Phase == ui.LoadPhaseStale || m.screenHasData(screen) {
		phase = ui.LoadPhaseStale
	}
	m.loadStates[screen] = ui.LoadState{Phase: phase, Error: previous.Error, HasSnapshot: previous.HasSnapshot}
}

func (m model) loadState(screen ui.Screen) ui.LoadState {
	if m.loadStates == nil {
		return ui.LoadState{Phase: ui.LoadPhaseIdle}
	}
	return m.loadStates[screen]
}

func (m model) screenHasData(screen ui.Screen) bool {
	switch screen {
	case ui.ScreenDefault:
		return len(m.dashboard.attentionRows) > 0 || len(m.dashboard.tokenRights.Projects) > 0
	case ui.ScreenStacks:
		return len(m.stacks) > 0
	case ui.ScreenNamespaces:
		return len(m.namespaces) > 0
	case ui.ScreenDependencies:
		return len(m.dependencies) > 0
	case ui.ScreenProjects:
		return len(m.projects) > 0 || len(m.projectDependencies) > 0
	case ui.ScreenSources:
		return len(m.sources) > 0
	case ui.ScreenPolicies:
		return len(m.policies) > 0 || len(m.policyValues) > 0
	case ui.ScreenView:
		return len(m.dependencyView.Rows) > 0
	case ui.ScreenSettings:
		return len(m.projectCheckRows) > 0
	case ui.ScreenReleases:
		return len(m.releaseRows) > 0
	case ui.ScreenVulnerabilities:
		return len(m.vulnRows) > 0 || len(m.vulnItems) > 0
	default:
		return false
	}
}
