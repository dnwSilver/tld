package app

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dnwSilver/tld/internal/ui"
)

// reloadCurrentScreen retries the screen's loaders without starting a batch refresh.
func (m *model) reloadCurrentScreen() tea.Cmd {
	if m.screen == ui.ScreenDefault {
		m.beginLoads(m.screen, 2)
	} else {
		m.beginLoad(m.screen)
	}
	switch m.screen {
	case ui.ScreenDefault:
		return tea.Batch(m.loadDashboardAttention(), m.loadTokenRights())
	case ui.ScreenStacks:
		return m.loadStacks()
	case ui.ScreenNamespaces:
		return m.loadNamespaces()
	case ui.ScreenDependencies:
		return m.loadDependencies()
	case ui.ScreenProjects:
		return m.loadProjects()
	case ui.ScreenSources:
		return m.loadSources()
	case ui.ScreenPolicies:
		return m.loadPolicies()
	case ui.ScreenView:
		return m.loadDependencyView()
	case ui.ScreenSettings:
		return m.loadProjects()
	case ui.ScreenReleases:
		return m.loadReleases()
	case ui.ScreenVulnerabilities:
		return m.loadVulnerabilities()
	default:
		return nil
	}
}
