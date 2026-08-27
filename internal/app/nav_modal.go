package app

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dnwSilver/tld/internal/ui"
)

func (m model) updateNavModal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch {
	case isCancelKey(msg), ui.KeyToggleHead.Matches(key):
		m.navModalOpen = false
	case ui.KeyPrev.Matches(key):
		if m.navModalIndex > 0 {
			m.navModalIndex--
		}
	case ui.KeyNext.Matches(key):
		if m.navModalIndex < len(ui.NavSections)-1 {
			m.navModalIndex++
		}
	case isEnterKey(msg):
		return m.navigateFromNavModal()
	default:
		for index, section := range ui.NavSections {
			if section.Binding.Matches(key) {
				m.navModalIndex = index
				return m.navigateFromNavModal()
			}
		}
	}

	return m, nil
}

func (m model) navigateFromNavModal() (tea.Model, tea.Cmd) {
	m.navModalOpen = false
	return m.switchToScreen(ui.NavSections[m.navModalIndex].Screen)
}

func (m model) switchToScreen(screen ui.Screen) (tea.Model, tea.Cmd) {
	m.screen = screen

	switch screen {
	case ui.ScreenDefault:
		return m, m.loadDashboardAttention()
	case ui.ScreenView:
		m.ensureViewStack()
		return m, m.loadDependencyView()
	case ui.ScreenSettings:
		return m, m.loadProjectChecks()
	case ui.ScreenReleases:
		return m, m.loadReleases()
	case ui.ScreenVulnerabilities:
		return m, m.loadVulnerabilities()
	default:
		return m, nil
	}
}

func (m model) openNavModal() model {
	m.navModalOpen = true
	m.navModalIndex = ui.NavSectionIndex(m.screen)
	return m
}
