package app

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dnwSilver/tld/internal/ui"
)

func (m model) dispatchScreenAction(action ui.ActionID) (tea.Model, tea.Cmd) {
	switch action {
	case ui.ActionAdd:
		m.openAddForm()
	case ui.ActionEdit:
		m.openEditForm()
	case ui.ActionClone:
		m.openCloneProjectForm()
	case ui.ActionDelete:
		m.openDelete()
	case ui.ActionRefreshRow:
		switch m.screen {
		case ui.ScreenProjects:
			return m.startProjectDependencySync()
		case ui.ScreenView:
			return m.startViewDependencySync(false)
		case ui.ScreenSettings:
			return m.startProjectChecksRefresh(m.selectedCheckProjects())
		case ui.ScreenReleases:
			return m.startReleasesRefresh(m.selectedReleaseProjects())
		case ui.ScreenVulnerabilities:
			return m.startVulnsRefresh(m.selectedVulnProjects())
		}
	case ui.ActionRefreshAll:
		switch m.screen {
		case ui.ScreenProjects:
			return m.startAllProjectsDependencySync()
		case ui.ScreenView:
			return m.startViewDependencySync(true)
		case ui.ScreenSettings:
			return m.startProjectChecksRefresh(activeProjects(m.projects))
		case ui.ScreenReleases:
			return m.startReleasesRefresh(activeProjects(m.projects))
		case ui.ScreenVulnerabilities:
			return m.startVulnsRefresh(activeProjects(m.projects))
		}
	case ui.ActionVulnMode:
		if !m.vulnsStatus.Running {
			m.vulnMode = m.vulnMode.Next()
			m.selectedVulnItemIndex = 0
			return m, m.loadVulnerabilities()
		}
	case ui.ActionUpdatePins:
		if !m.policyUpdateStatus.Running && m.policyFocus == ui.PolicyPaneValues {
			return m.startSelectedPolicyValueUpdate()
		}
	case ui.ActionOperations:
		if !m.checksStatus.Running {
			m = m.openOperationsModal()
		}
	case ui.ActionSort:
		m.sortProjectChecksBySelectedColumn()
	case ui.ActionToggleFocus:
		switch m.screen {
		case ui.ScreenDefault:
			m.toggleDashboardPane()
		case ui.ScreenView:
			m.selectNextViewStack()
			m.beginLoad(ui.ScreenView)
			return m, m.loadDependencyView()
		case ui.ScreenProjects:
			m.toggleProjectPane()
		case ui.ScreenReleases:
			m.releasePeriod = m.releasePeriod.Next()
			m.beginLoad(ui.ScreenReleases)
			return m, m.loadReleases()
		case ui.ScreenVulnerabilities:
			m.toggleVulnPane()
		case ui.ScreenPolicies:
			m.togglePolicyPane()
		}
	}
	return m, nil
}

func (m model) dispatchGlobalAction(action ui.ActionID) (tea.Model, tea.Cmd) {
	switch action {
	case ui.ActionOpenNavigation:
		m = m.openNavModal()
		return m, nil
	case ui.ActionSearch:
		m.searchOpen = true
		switch m.screen {
		case ui.ScreenDefault:
			m.dashboard.focus = ui.DashboardPaneAttention
		case ui.ScreenProjects:
			m.projectFocus = ui.ProjectPaneProjects
		case ui.ScreenPolicies:
			m.policyFocus = ui.PolicyPanePolicies
		case ui.ScreenVulnerabilities:
			m.vulnFocus = ui.VulnPaneProjects
		}
		return m, nil
	case ui.ActionReload:
		m.lastError = ""
		return m, m.reloadCurrentScreen()
	case ui.ActionQuit:
		return m, tea.Quit
	case ui.ActionGoHome:
		return m.switchToScreen(ui.ScreenDefault)
	case ui.ActionGoStacks:
		return m.switchToScreen(ui.ScreenStacks)
	case ui.ActionGoNamespaces:
		return m.switchToScreen(ui.ScreenNamespaces)
	case ui.ActionGoDependencies:
		return m.switchToScreen(ui.ScreenDependencies)
	case ui.ActionGoSources:
		return m.switchToScreen(ui.ScreenSources)
	case ui.ActionGoProjects:
		return m.switchToScreen(ui.ScreenProjects)
	case ui.ActionGoPolicies:
		return m.switchToScreen(ui.ScreenPolicies)
	case ui.ActionGoView:
		return m.switchToScreen(ui.ScreenView)
	case ui.ActionGoSettings:
		return m.switchToScreen(ui.ScreenSettings)
	case ui.ActionGoReleases:
		return m.switchToScreen(ui.ScreenReleases)
	case ui.ActionGoVulnerabilities:
		return m.switchToScreen(ui.ScreenVulnerabilities)
	default:
		return m, nil
	}
}
