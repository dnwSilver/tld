package app

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dnwSilver/tld/internal/ui"
)

func (m model) screenJobRunning() bool {
	switch m.screen {
	case ui.ScreenProjects, ui.ScreenView:
		return m.projectSyncStatus.Running
	case ui.ScreenSettings:
		return m.checksStatus.Running
	case ui.ScreenReleases:
		return m.releasesStatus.Running
	case ui.ScreenVulnerabilities:
		return m.vulnsStatus.Running
	case ui.ScreenPolicies:
		return m.policyUpdateStatus.Running
	default:
		return false
	}
}

// cancelScreenJob invalidates queued messages before a new run can start.
// The worker observes its context and exits without depending on a UI reader.
func (m *model) cancelScreenJob() (tea.Cmd, bool) {
	switch m.screen {
	case ui.ScreenProjects, ui.ScreenView:
		if !m.projectSyncStatus.Running {
			return nil, false
		}
		if m.projectSyncCancel != nil {
			m.projectSyncCancel()
		}
		m.projectSyncCancel = nil
		m.projectSyncRunID++
		m.projectSyncCh = nil
		m.projectSyncStatus.Running = false
		m.projectSyncStatus.Message = "Sync canceled"
		m.projectSyncStatus.Error = ""
		return m.reloadAfterProjectSync(), true
	case ui.ScreenSettings:
		if !m.checksStatus.Running {
			return nil, false
		}
		if m.checksCancel != nil {
			m.checksCancel()
		}
		m.checksCancel = nil
		m.checksRunID++
		m.checksSyncCh = nil
		m.checksStatus.Running = false
		m.checksStatus.Message = "Checks canceled"
		m.checksStatus.Error = ""
		return m.loadProjectChecks(), true
	case ui.ScreenReleases:
		if !m.releasesStatus.Running {
			return nil, false
		}
		if m.releasesCancel != nil {
			m.releasesCancel()
		}
		m.releasesCancel = nil
		m.releasesRunID++
		m.releasesSyncCh = nil
		m.releasesStatus.Running = false
		m.releasesStatus.Message = "Releases canceled"
		m.releasesStatus.Error = ""
		return m.loadReleases(), true
	case ui.ScreenVulnerabilities:
		if !m.vulnsStatus.Running {
			return nil, false
		}
		if m.vulnsCancel != nil {
			m.vulnsCancel()
		}
		m.vulnsCancel = nil
		m.vulnsRunID++
		m.vulnsSyncCh = nil
		m.vulnsStatus.Running = false
		m.vulnsStatus.Message = "Scan canceled"
		m.vulnsStatus.Error = ""
		return m.loadVulnerabilities(), true
	case ui.ScreenPolicies:
		if !m.policyUpdateStatus.Running {
			return nil, false
		}
		if m.policyPinsCancel != nil {
			m.policyPinsCancel()
		}
		m.policyPinsCancel = nil
		m.policyPinsRunID++
		m.policyPinsSyncCh = nil
		m.policyUpdateStatus.Running = false
		m.policyUpdateStatus.Message = "Pins canceled"
		m.policyUpdateStatus.Error = ""
		return m.loadPolicyValues(), true
	default:
		return nil, false
	}
}
