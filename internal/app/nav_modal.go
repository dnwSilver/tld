package app

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dnwSilver/tld/internal/ui"
)

func (m model) updateNavModal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	action, ok := ui.ModalActionForKey(ui.ModalNavigation, key)
	if !ok {
		for index, section := range ui.NavSections {
			if section.Binding.Matches(key) {
				m.navModalIndex = index
				return m.navigateFromNavModal()
			}
		}
		return m, nil
	}
	switch action {
	case ui.ModalActionCancel:
		m.navModalOpen = false
	case ui.ModalActionPrev:
		if m.navModalIndex > 0 {
			m.navModalIndex--
		}
	case ui.ModalActionNext:
		if m.navModalIndex < len(ui.NavSections)-1 {
			m.navModalIndex++
		}
	case ui.ModalActionConfirm:
		return m.navigateFromNavModal()
	}

	return m, nil
}

func (m model) navigateFromNavModal() (tea.Model, tea.Cmd) {
	m.navModalOpen = false
	return m.switchToScreen(ui.NavSections[m.navModalIndex].Screen)
}

func (m model) switchToScreen(screen ui.Screen) (tea.Model, tea.Cmd) {
	if m.screen != screen {
		m.searchOpen = false
		m.searchQuery = ""
	}
	m.screen = screen

	switch screen {
	case ui.ScreenDefault:
		m.beginLoads(screen, 2)
		return m, tea.Batch(m.loadDashboardAttention(), m.loadTokenRights())
	case ui.ScreenView:
		m.ensureViewStack()
		m.beginLoad(screen)
		return m, m.loadDependencyView()
	case ui.ScreenSettings:
		m.beginLoad(screen)
		return m, m.loadProjectChecks()
	case ui.ScreenReleases:
		m.beginLoad(screen)
		return m, m.loadReleases()
	case ui.ScreenVulnerabilities:
		m.beginLoad(screen)
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
