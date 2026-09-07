package app

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dnwSilver/tld/internal/projectsync"
	"github.com/dnwSilver/tld/internal/registry"
	"github.com/dnwSilver/tld/internal/storage"
	"github.com/dnwSilver/tld/internal/ui"
)

type model struct {
	width                    int
	height                   int
	creator                  ui.Creator
	store                    *storage.Store
	screen                   ui.Screen
	stacks                   []ui.Stack
	dashboard                dashboardState
	selectedStackID          int64
	namespaces               []ui.Namespace
	selectedNamespaceID      int64
	dependencies             []ui.Dependency
	selectedDependencyID     int64
	projects                 []ui.Project
	selectedProjectID        int64
	projectDependencies      []ui.ProjectDependency
	selectedProjectDepID     int64
	projectLatestRun         ui.ProjectDependencyRun
	projectSyncStatus        ui.ProjectSyncStatus
	projectSyncCh            <-chan projectSyncMsg
	projectFocus             ui.ProjectPane
	sources                  []ui.Source
	selectedSourceID         int64
	policies                 []ui.Policy
	selectedPolicyID         int64
	policyValues             []ui.PolicyValue
	selectedPolicyValueID    int64
	policyFocus              ui.PolicyPane
	policyUpdateForm         ui.PolicyUpdateForm
	policyUpdateStatus       ui.SettingsStatus
	dependencyView           ui.DependencyView
	viewStackID              int64
	selectedViewProjectID    int64
	viewColumnOffset         int
	form                     ui.StackForm
	namespaceForm            ui.StackForm
	dependencyForm           ui.DependencyForm
	projectForm              ui.ProjectForm
	sourceForm               ui.SourceForm
	policyForm               ui.PolicyForm
	policyValueForm          ui.PolicyValueForm
	deleteConfirm            ui.DeleteConfirm
	namespaceDeleteConfirm   ui.DeleteConfirm
	dependencyDeleteConfirm  ui.DeleteConfirm
	projectDeleteConfirm     ui.DeleteConfirm
	sourceDeleteConfirm      ui.DeleteConfirm
	policyDeleteConfirm      ui.DeleteConfirm
	checkColumns             []ui.ProjectCheck
	projectCheckRows         []ui.ProjectCheckRow
	selectedCheckProjectID   int64
	settingsTableState       ui.SettingsTableState
	checksStatus             ui.SettingsStatus
	checksSyncCh             <-chan checkSyncMsg
	releaseRows              []ui.ReleaseRow
	selectedReleaseProjectID int64
	releasePeriod            ui.ReleasePeriod
	releasesStatus           ui.SettingsStatus
	releasesSyncCh           <-chan releaseSyncMsg
	vulnRows                 []ui.VulnProjectRow
	vulnItems                []ui.VulnerabilityItem
	selectedVulnProjectID    int64
	selectedVulnItemIndex    int
	vulnFocus                ui.VulnPane
	vulnMode                 ui.VulnMode
	vulnsStatus              ui.SettingsStatus
	vulnsSyncCh              <-chan vulnSyncMsg
	navModalOpen             bool
	navModalIndex            int
	settingsOperations       []ui.SettingsOperation
	operationsModalOpen      bool
	operationsModalIndex     int
	err                      error
}

type stacksLoadedMsg struct {
	stacks []ui.Stack
	err    error
}

type namespacesLoadedMsg struct {
	namespaces []ui.Namespace
	err        error
}

type dependenciesLoadedMsg struct {
	dependencies []ui.Dependency
	err          error
}

type projectsLoadedMsg struct {
	projects []ui.Project
	err      error
}

type projectDependenciesLoadedMsg struct {
	dependencies []ui.ProjectDependency
	latestRun    ui.ProjectDependencyRun
	err          error
}

type projectSyncMsg struct {
	projectID int64
	message   string
	err       error
	done      bool
	step      bool
	current   int
	total     int
}

type sourcesLoadedMsg struct {
	sources []ui.Source
	err     error
}

type policiesLoadedMsg struct {
	policies []ui.Policy
	err      error
}

type policyValuesLoadedMsg struct {
	values []ui.PolicyValue
	err    error
}

type dependencyViewLoadedMsg struct {
	view ui.DependencyView
	err  error
}

type stackSavedMsg struct {
	stackID int64
	err     error
}

type namespaceSavedMsg struct {
	namespaceID int64
	err         error
}

type dependencySavedMsg struct {
	dependencyID int64
	err          error
}

type projectSavedMsg struct {
	projectID int64
	err       error
}

type sourceSavedMsg struct {
	sourceID int64
	err      error
}

type policySavedMsg struct {
	policyID int64
	err      error
}

type policyValueSavedMsg struct {
	valueID int64
	err     error
}

type stackDeletedMsg struct {
	stackID int64
	err     error
}

type namespaceDeletedMsg struct {
	namespaceID int64
	err         error
}

type dependencyDeletedMsg struct {
	dependencyID int64
	err          error
}

type projectDeletedMsg struct {
	projectID int64
	err       error
}

type sourceDeletedMsg struct {
	sourceID int64
	err      error
}

type policyDeletedMsg struct {
	policyID int64
	err      error
}

type policyValueDeletedMsg struct {
	valueID int64
	err     error
}

type policyPinsUpdatedMsg struct {
	checked  int
	updated  int
	failures []string
	err      error
}

type policyValueLatestLoadedMsg struct {
	version string
	err     error
}

func newModel(store *storage.Store) model {
	return model{
		creator:            ui.NewCreator(),
		store:              store,
		screen:             ui.ScreenDefault,
		stacks:             []ui.Stack{},
		namespaces:         []ui.Namespace{},
		dependencies:       []ui.Dependency{},
		projects:           []ui.Project{},
		projectFocus:       ui.ProjectPaneProjects,
		sources:            []ui.Source{},
		policies:           []ui.Policy{},
		policyValues:       []ui.PolicyValue{},
		policyFocus:        ui.PolicyPanePolicies,
		checkColumns:       defaultCheckColumns(),
		vulnFocus:          ui.VulnPaneProjects,
		settingsOperations: settingsOperationItems(),
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.loadStacks(), m.loadDashboardAttention(), m.loadTokenRights(), m.loadNamespaces(), m.loadDependencies(), m.loadProjects(), m.loadSources(), m.loadPolicies())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		if m.policyDeleteConfirm.Open {
			return m.updatePolicyDeleteConfirm(msg)
		}
		if m.sourceDeleteConfirm.Open {
			return m.updateSourceDeleteConfirm(msg)
		}
		if m.dependencyDeleteConfirm.Open {
			return m.updateDependencyDeleteConfirm(msg)
		}
		if m.projectDeleteConfirm.Open {
			return m.updateProjectDeleteConfirm(msg)
		}
		if m.namespaceDeleteConfirm.Open {
			return m.updateNamespaceDeleteConfirm(msg)
		}
		if m.deleteConfirm.Open {
			return m.updateDeleteConfirm(msg)
		}
		if m.dependencyForm.Open {
			return m.updateDependencyForm(msg)
		}
		if m.projectForm.Open {
			return m.updateProjectForm(msg)
		}
		if m.sourceForm.Open {
			return m.updateSourceForm(msg)
		}
		if m.policyForm.Open {
			return m.updatePolicyForm(msg)
		}
		if m.policyValueForm.Open {
			return m.updatePolicyValueForm(msg)
		}
		if m.policyUpdateForm.Open {
			return m.updatePolicyUpdateForm(msg)
		}
		if m.namespaceForm.Open {
			return m.updateNamespaceForm(msg)
		}
		if m.form.Open {
			return m.updateStackForm(msg)
		}
		if m.operationsModalOpen {
			return m.updateOperationsModal(msg)
		}
		if m.navModalOpen {
			return m.updateNavModal(msg)
		}

		key := msg.String()
		switch {
		case ui.KeyToggleHead.Matches(key):
			m = m.openNavModal()
		case ui.KeyHome.Matches(key):
			return m.switchToScreen(ui.ScreenDefault)
		case ui.KeyStacks.Matches(key):
			return m.switchToScreen(ui.ScreenStacks)
		case ui.KeyNamespaces.Matches(key):
			return m.switchToScreen(ui.ScreenNamespaces)
		case ui.KeyDependencies.Matches(key):
			return m.switchToScreen(ui.ScreenDependencies)
		case ui.KeyProjects.Matches(key):
			return m.switchToScreen(ui.ScreenProjects)
		case ui.KeySources.Matches(key):
			return m.switchToScreen(ui.ScreenSources)
		case ui.KeyPolicies.Matches(key):
			return m.switchToScreen(ui.ScreenPolicies)
		case ui.KeyView.Matches(key):
			return m.switchToScreen(ui.ScreenView)
		case ui.KeySettings.Matches(key):
			return m.switchToScreen(ui.ScreenSettings)
		case ui.KeyReleases.Matches(key):
			return m.switchToScreen(ui.ScreenReleases)
		case ui.KeyVulnerabilities.Matches(key):
			return m.switchToScreen(ui.ScreenVulnerabilities)
		case ui.KeyAdd.Matches(key):
			m.openAddForm()
		case ui.KeyEdit.Matches(key):
			m.openEditForm()
		case ui.KeyClone.Matches(key):
			if m.screen == ui.ScreenProjects {
				m.openCloneProjectForm()
			}
		case ui.KeyDelete.Matches(key):
			m.openDelete()
		case ui.KeyRefreshDeps.Matches(key):
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
		case ui.KeyRefreshAll.Matches(key):
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
		case ui.KeyVulnMode.Matches(key):
			if m.screen == ui.ScreenVulnerabilities && !m.vulnsStatus.Running {
				m.vulnMode = m.vulnMode.Next()
				m.selectedVulnItemIndex = 0
				return m, m.loadVulnerabilities()
			}
		case ui.KeyUpdatePins.Matches(key):
			if m.screen == ui.ScreenPolicies && !m.policyUpdateStatus.Running {
				if m.policyFocus == ui.PolicyPaneValues {
					return m.startSelectedPolicyValueUpdate()
				}
			}
			if m.screen == ui.ScreenSettings && !m.checksStatus.Running {
				m = m.openOperationsModal()
			}
		case ui.KeyColumnPick.Matches(key):
			if m.screen == ui.ScreenSettings {
				if key == "shift+left" {
					m.selectPreviousCheckColumn()
				} else {
					m.selectNextCheckColumn()
				}
				return m, nil
			}
		case ui.KeySort.Matches(key):
			if m.screen == ui.ScreenSettings {
				m.sortProjectChecksBySelectedColumn()
				return m, nil
			}
		case ui.KeyPrev.Matches(key):
			if m.screen == ui.ScreenView {
				m.selectPreviousViewProject()
				return m, nil
			}
			m.selectPreviousOnScreen()
			if m.screen == ui.ScreenPolicies && m.policyFocus == ui.PolicyPanePolicies {
				return m, m.loadPolicyValues()
			}
			if m.screen == ui.ScreenProjects {
				return m, m.loadProjectDependencies()
			}
			if m.screen == ui.ScreenVulnerabilities && m.vulnFocus == ui.VulnPaneProjects {
				return m, m.loadVulnerabilities()
			}
		case ui.KeyNext.Matches(key):
			if m.screen == ui.ScreenView {
				m.selectNextViewProject()
				return m, nil
			}
			m.selectNextOnScreen()
			if m.screen == ui.ScreenPolicies && m.policyFocus == ui.PolicyPanePolicies {
				return m, m.loadPolicyValues()
			}
			if m.screen == ui.ScreenProjects {
				return m, m.loadProjectDependencies()
			}
			if m.screen == ui.ScreenVulnerabilities && m.vulnFocus == ui.VulnPaneProjects {
				return m, m.loadVulnerabilities()
			}
			if m.screen == ui.ScreenView {
				m.selectNextViewStack()
				return m, m.loadDependencyView()
			}
		case key == "left":
			if m.screen == ui.ScreenDefault {
				m.dashboard.focus = ui.DashboardPaneSummary
				return m, nil
			}
			if m.screen == ui.ScreenView {
				m.scrollViewColumnsLeft()
				return m, nil
			}
		case key == "right":
			if m.screen == ui.ScreenDefault {
				m.dashboard.focus = ui.DashboardPaneAttention
				m.ensureSelectedAttentionProject()
				return m, nil
			}
			if m.screen == ui.ScreenView {
				m.scrollViewColumnsRight()
				return m, nil
			}
		case isOneOf(msg, "h", "shift+tab"):
			if m.screen == ui.ScreenDefault {
				m.dashboard.focus = ui.DashboardPaneSummary
				return m, nil
			}
			if m.screen == ui.ScreenView {
				m.selectPreviousViewStack()
				return m, m.loadDependencyView()
			}
		case isOneOf(msg, "l"):
			if m.screen == ui.ScreenDefault {
				m.dashboard.focus = ui.DashboardPaneAttention
				m.ensureSelectedAttentionProject()
				return m, nil
			}
			if m.screen == ui.ScreenView {
				m.selectNextViewStack()
				return m, m.loadDependencyView()
			}
		case key == "tab":
			if m.screen == ui.ScreenDefault {
				m.toggleDashboardPane()
				return m, nil
			}
			if m.screen == ui.ScreenView {
				m.selectNextViewStack()
				return m, m.loadDependencyView()
			}
			if m.screen == ui.ScreenProjects {
				m.toggleProjectPane()
				return m, nil
			}
			if m.screen == ui.ScreenReleases {
				m.releasePeriod = m.releasePeriod.Next()
				return m, m.loadReleases()
			}
			if m.screen == ui.ScreenVulnerabilities {
				m.toggleVulnPane()
				return m, nil
			}
			m.togglePolicyPane()
		case ui.KeyQuit.Matches(key):
			return m, tea.Quit
		}
	case stacksLoadedMsg:
		m.err = msg.err
		if msg.err == nil {
			m.stacks = msg.stacks
			m.ensureSelectedStack()
		}
	case namespacesLoadedMsg:
		m.err = msg.err
		if msg.err == nil {
			m.namespaces = msg.namespaces
			m.ensureSelectedNamespace()
		}
	case dependenciesLoadedMsg:
		m.err = msg.err
		if msg.err == nil {
			m.dependencies = msg.dependencies
			m.ensureSelectedDependency()
		}
	case projectsLoadedMsg:
		m.err = msg.err
		if msg.err == nil {
			m.projects = msg.projects
			m.ensureSelectedProject()
			if m.screen == ui.ScreenSettings {
				return m, tea.Batch(m.loadProjectDependencies(), m.loadProjectChecks())
			}
			return m, m.loadProjectDependencies()
		}
	case projectDependenciesLoadedMsg:
		m.err = msg.err
		if msg.err == nil {
			m.projectDependencies = msg.dependencies
			m.ensureSelectedProjectDependency()
			m.projectLatestRun = msg.latestRun
		}
	case sourcesLoadedMsg:
		m.err = msg.err
		if msg.err == nil {
			m.sources = msg.sources
			m.ensureSelectedSource()
		}
	case policiesLoadedMsg:
		m.err = msg.err
		if msg.err == nil {
			m.policies = msg.policies
			m.ensureSelectedPolicy()
			return m, m.loadPolicyValues()
		}
	case policyValuesLoadedMsg:
		m.err = msg.err
		if msg.err == nil {
			m.policyValues = msg.values
			m.ensureSelectedPolicyValue()
		}
	case dependencyViewLoadedMsg:
		m.err = msg.err
		if msg.err == nil {
			m.dependencyView = msg.view
			m.ensureSelectedViewProject()
		}
	case stackSavedMsg:
		m.err = msg.err
		if msg.err == nil {
			m.selectedStackID = msg.stackID
			m.form = ui.StackForm{}
			return m, tea.Batch(m.loadStacks(), m.loadDependencies(), m.loadProjects())
		}
		m.form.Error = msg.err.Error()
	case namespaceSavedMsg:
		m.err = msg.err
		if msg.err == nil {
			m.selectedNamespaceID = msg.namespaceID
			m.namespaceForm = ui.StackForm{}
			return m, tea.Batch(m.loadNamespaces(), m.loadPolicies(), m.loadProjects())
		}
		m.namespaceForm.Error = msg.err.Error()
	case dependencySavedMsg:
		m.err = msg.err
		if msg.err == nil {
			m.selectedDependencyID = msg.dependencyID
			m.dependencyForm = ui.DependencyForm{}
			return m, m.loadDependencies()
		}
		m.dependencyForm.Error = msg.err.Error()
	case projectSavedMsg:
		m.err = msg.err
		if msg.err == nil {
			m.selectedProjectID = msg.projectID
			m.projectForm = ui.ProjectForm{}
			return m, m.loadProjects()
		}
		m.projectForm.Error = msg.err.Error()
	case sourceSavedMsg:
		m.err = msg.err
		if msg.err == nil {
			m.selectedSourceID = msg.sourceID
			m.sourceForm = ui.SourceForm{}
			return m, tea.Batch(m.loadSources(), m.loadProjects())
		}
		m.sourceForm.Error = msg.err.Error()
	case policySavedMsg:
		m.err = msg.err
		if msg.err == nil {
			m.selectedPolicyID = msg.policyID
			m.policyForm = ui.PolicyForm{}
			return m, m.loadPolicies()
		}
		m.policyForm.Error = msg.err.Error()
	case policyValueSavedMsg:
		m.err = msg.err
		if msg.err == nil {
			m.selectedPolicyValueID = msg.valueID
			m.policyValueForm = ui.PolicyValueForm{}
			return m, m.loadPolicyValues()
		}
		m.policyValueForm.Error = msg.err.Error()
	case dashboardAttentionLoadedMsg:
		m.err = msg.err
		if msg.err == nil {
			m.dashboard.attentionRows = msg.rows
			m.ensureSelectedAttentionProject()
		}
	case tokenRightsLoadedMsg:
		m.err = msg.err
		if msg.err == nil {
			m.dashboard.tokenRights = msg.rights
		}
	case policyValueLatestLoadedMsg:
		m.err = msg.err
		if msg.err != nil {
			m.policyValueForm.Error = msg.err.Error()
			return m, nil
		}
		m.policyValueForm.Version = msg.version
		m.policyValueForm.Error = ""
	case stackDeletedMsg:
		m.err = msg.err
		if msg.err == nil {
			if m.selectedStackID == msg.stackID {
				m.selectedStackID = 0
			}
			m.deleteConfirm = ui.DeleteConfirm{}
			return m, m.loadStacks()
		}
		m.deleteConfirm.Error = msg.err.Error()
	case namespaceDeletedMsg:
		m.err = msg.err
		if msg.err == nil {
			if m.selectedNamespaceID == msg.namespaceID {
				m.selectedNamespaceID = 0
			}
			m.namespaceDeleteConfirm = ui.DeleteConfirm{}
			return m, tea.Batch(m.loadNamespaces(), m.loadPolicies())
		}
		m.namespaceDeleteConfirm.Error = msg.err.Error()
	case dependencyDeletedMsg:
		m.err = msg.err
		if msg.err == nil {
			if m.selectedDependencyID == msg.dependencyID {
				m.selectedDependencyID = 0
			}
			m.dependencyDeleteConfirm = ui.DeleteConfirm{}
			return m, m.loadDependencies()
		}
		m.dependencyDeleteConfirm.Error = msg.err.Error()
	case projectDeletedMsg:
		m.err = msg.err
		if msg.err == nil {
			if m.selectedProjectID == msg.projectID {
				m.selectedProjectID = 0
			}
			m.projectDeleteConfirm = ui.DeleteConfirm{}
			return m, m.loadProjects()
		}
		m.projectDeleteConfirm.Error = msg.err.Error()
	case sourceDeletedMsg:
		m.err = msg.err
		if msg.err == nil {
			if m.selectedSourceID == msg.sourceID {
				m.selectedSourceID = 0
			}
			m.sourceDeleteConfirm = ui.DeleteConfirm{}
			return m, m.loadSources()
		}
		m.sourceDeleteConfirm.Error = msg.err.Error()
	case policyDeletedMsg:
		m.err = msg.err
		if msg.err == nil {
			if m.selectedPolicyID == msg.policyID {
				m.selectedPolicyID = 0
			}
			m.policyValues = []ui.PolicyValue{}
			m.selectedPolicyValueID = 0
			m.policyDeleteConfirm = ui.DeleteConfirm{}
			return m, m.loadPolicies()
		}
		m.policyDeleteConfirm.Error = msg.err.Error()
	case policyValueDeletedMsg:
		m.err = msg.err
		if msg.err == nil {
			if m.selectedPolicyValueID == msg.valueID {
				m.selectedPolicyValueID = 0
			}
			m.policyDeleteConfirm = ui.DeleteConfirm{}
			return m, m.loadPolicyValues()
		}
		m.policyDeleteConfirm.Error = msg.err.Error()
	case policyPinsUpdatedMsg:
		m.policyUpdateStatus.Running = false
		m.policyUpdateStatus.Current = msg.checked
		m.policyUpdateStatus.Total = msg.checked
		if msg.err != nil {
			m.err = msg.err
			m.policyUpdateStatus.Message = "Pins update failed"
			m.policyUpdateStatus.Error = msg.err.Error()
			return m, m.loadPolicyValues()
		}
		m.policyUpdateStatus.Message = fmt.Sprintf("Updated %d of %d pins", msg.updated, msg.checked)
		m.policyUpdateStatus.Error = strings.Join(msg.failures, "; ")
		return m, m.loadPolicyValues()
	case projectChecksLoadedMsg:
		m.err = msg.err
		if msg.err == nil {
			m.projectCheckRows = msg.rows
			m.applyProjectCheckSort()
			m.ensureSelectedCheckProject()
		}
	case operationAppliedMsg:
		m.checksStatus.Running = false
		if msg.err != nil {
			m.err = msg.err
			m.checksStatus.Message = msg.title + " failed for " + msg.project
			m.checksStatus.Error = msg.err.Error()
			return m, nil
		}
		m.checksStatus.Message = msg.title + " applied to " + msg.project
		m.checksStatus.Error = ""
		if project, ok := findByID(m.projects, msg.projectID, projectID); ok {
			return m.startProjectChecksRefresh([]ui.Project{project})
		}
	case checkSyncMsg:
		m.checksStatus.Message = msg.message
		m.checksStatus.Running = !msg.done
		m.checksStatus.Error = ""
		if msg.total > 0 {
			m.checksStatus.Current = msg.current
			m.checksStatus.Total = msg.total
		}
		if msg.err != nil {
			m.err = msg.err
			m.checksStatus.Error = msg.err.Error()
			m.checksStatus.Running = false
			return m, m.loadProjectChecks()
		}
		if msg.done {
			m.checksSyncCh = nil
			return m, m.loadProjectChecks()
		}
		if msg.step && m.checksSyncCh != nil {
			return m, tea.Batch(m.loadProjectChecks(), waitCheckSync(m.checksSyncCh))
		}
		if m.checksSyncCh != nil {
			return m, waitCheckSync(m.checksSyncCh)
		}
	case releasesLoadedMsg:
		m.err = msg.err
		if msg.err == nil {
			m.releaseRows = msg.rows
			m.ensureSelectedReleaseProject()
		}
	case releaseSyncMsg:
		m.releasesStatus.Message = msg.message
		m.releasesStatus.Running = !msg.done
		m.releasesStatus.Error = ""
		if msg.total > 0 {
			m.releasesStatus.Current = msg.current
			m.releasesStatus.Total = msg.total
		}
		if msg.err != nil {
			m.err = msg.err
			m.releasesStatus.Error = msg.err.Error()
			m.releasesStatus.Running = false
			return m, m.loadReleases()
		}
		if msg.done {
			m.releasesSyncCh = nil
			return m, m.loadReleases()
		}
		if msg.step && m.releasesSyncCh != nil {
			return m, tea.Batch(m.loadReleases(), waitReleaseSync(m.releasesSyncCh))
		}
		if m.releasesSyncCh != nil {
			return m, waitReleaseSync(m.releasesSyncCh)
		}
	case vulnsLoadedMsg:
		m.err = msg.err
		if msg.err == nil {
			m.vulnRows = msg.rows
			m.vulnItems = msg.items
			m.ensureSelectedVulnProject()
			if m.selectedVulnItemIndex >= len(m.vulnItems) {
				m.selectedVulnItemIndex = 0
			}
		}
	case vulnSyncMsg:
		m.vulnsStatus.Message = msg.message
		m.vulnsStatus.Running = !msg.done
		m.vulnsStatus.Error = ""
		if msg.total > 0 {
			m.vulnsStatus.Current = msg.current
			m.vulnsStatus.Total = msg.total
		}
		if msg.err != nil {
			m.err = msg.err
			m.vulnsStatus.Error = msg.err.Error()
			m.vulnsStatus.Running = false
			return m, m.loadVulnerabilities()
		}
		if msg.done {
			m.vulnsSyncCh = nil
			return m, m.loadVulnerabilities()
		}
		if msg.step && m.vulnsSyncCh != nil {
			return m, tea.Batch(m.loadVulnerabilities(), waitVulnSync(m.vulnsSyncCh))
		}
		if m.vulnsSyncCh != nil {
			return m, waitVulnSync(m.vulnsSyncCh)
		}
	case projectSyncMsg:
		if msg.projectID != 0 && msg.projectID != m.selectedProjectID {
			return m, nil
		}
		m.projectSyncStatus.ProjectID = msg.projectID
		m.projectSyncStatus.Message = msg.message
		m.projectSyncStatus.Running = !msg.done
		m.projectSyncStatus.Error = ""
		if msg.total > 0 {
			m.projectSyncStatus.Current = msg.current
			m.projectSyncStatus.Total = msg.total
		}
		if msg.err != nil {
			m.err = msg.err
			m.projectSyncStatus.Error = msg.err.Error()
			m.projectSyncStatus.Running = false
			return m, m.loadProjectDependencies()
		}
		if msg.done {
			m.projectSyncCh = nil
			return m, m.reloadAfterProjectSync()
		}
		if msg.step && m.projectSyncCh != nil {
			return m, tea.Batch(m.reloadAfterProjectSync(), waitProjectSync(m.projectSyncCh))
		}
		if m.projectSyncCh != nil {
			return m, waitProjectSync(m.projectSyncCh)
		}
	}

	return m, nil
}

func (m model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	return m.creator.Render(
		m.width,
		m.height,
		m.screen,
		m.stacks,
		m.dashboard.tokenRights,
		m.dashboard.attentionRows,
		m.dashboard.focus,
		m.dashboard.selectedProjectID,
		m.selectedStackID,
		m.namespaces,
		m.selectedNamespaceID,
		m.dependencies,
		m.selectedDependencyID,
		m.projects,
		m.selectedProjectID,
		m.projectDependencies,
		m.selectedProjectDepID,
		m.projectLatestRun,
		m.projectSyncStatus,
		m.projectFocus,
		m.sources,
		m.selectedSourceID,
		m.policies,
		m.selectedPolicyID,
		m.policyValues,
		m.selectedPolicyValueID,
		m.policyFocus,
		m.policyUpdateStatus,
		m.dependencyView,
		m.selectedViewProjectID,
		m.viewColumnOffset,
		m.checkColumns,
		m.projectCheckRows,
		m.selectedCheckProjectID,
		m.settingsTableState,
		m.checksStatus,
		m.releaseRows,
		m.selectedReleaseProjectID,
		m.releasePeriod,
		m.releasesStatus,
		m.vulnRows,
		m.selectedVulnProjectID,
		m.vulnItems,
		m.selectedVulnItemIndex,
		m.vulnFocus,
		m.vulnMode,
		m.vulnsStatus,
		m.form,
		m.namespaceForm,
		m.dependencyForm,
		m.projectForm,
		m.sourceForm,
		m.policyForm,
		m.policyValueForm,
		m.policyUpdateForm,
		m.deleteConfirm,
		m.namespaceDeleteConfirm,
		m.dependencyDeleteConfirm,
		m.projectDeleteConfirm,
		m.sourceDeleteConfirm,
		m.policyDeleteConfirm,
		m.navModalOpen,
		m.navModalIndex,
		m.settingsOperations,
		m.operationsModalOpen,
		m.operationsModalIndex,
	)
}

func (m model) updateStackForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var action stackFormAction
	m.form, action = updateStackFormState(msg, m.form)
	return m.finishStackFormAction(action, m.saveStack)
}

func (m model) updateNamespaceForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var action stackFormAction
	m.namespaceForm, action = updateNamespaceFormState(msg, m.namespaceForm, m.policies)
	return m.finishStackFormAction(action, m.saveNamespace)
}

type stackFormAction int

const (
	stackFormActionNone stackFormAction = iota
	stackFormActionQuit
	stackFormActionSave
)

func updateStackFormState(msg tea.KeyMsg, form ui.StackForm) (ui.StackForm, stackFormAction) {
	switch {
	case isQuitKey(msg):
		return form, stackFormActionQuit
	case isCancelKey(msg):
		return ui.StackForm{}, stackFormActionNone
	case isEnterKey(msg):
		if !form.CanSave {
			return form, stackFormActionNone
		}
		return form, stackFormActionSave
	case isOneOf(msg, "tab", "down"):
		form.Focus = nextStackFormField(form.Focus)
	case isOneOf(msg, "shift+tab", "up"):
		form.Focus = previousStackFormField(form.Focus)
	case isBackspaceKey(msg):
		form = deleteStackFormRune(form)
	default:
		if msg.Type == tea.KeyRunes {
			form = appendStackFormRunes(form, msg.Runes)
		}
	}

	return normalizeStackForm(form), stackFormActionNone
}

func updateNamespaceFormState(msg tea.KeyMsg, form ui.StackForm, policies []ui.Policy) (ui.StackForm, stackFormAction) {
	switch {
	case isQuitKey(msg):
		return form, stackFormActionQuit
	case isCancelKey(msg):
		return ui.StackForm{}, stackFormActionNone
	case isEnterKey(msg):
		if !form.CanSave {
			return form, stackFormActionNone
		}
		return form, stackFormActionSave
	case isOneOf(msg, "tab", "down"):
		form.Focus = nextNamespaceFormField(form.Focus)
	case isOneOf(msg, "shift+tab", "up"):
		form.Focus = previousNamespaceFormField(form.Focus)
	case form.Focus == ui.StackFormFieldPolicy && isOneOf(msg, "left", "h"):
		form.PolicyID = previousPolicyID(policies, form.PolicyID)
	case form.Focus == ui.StackFormFieldPolicy && isOneOf(msg, "right", "l"):
		form.PolicyID = nextPolicyID(policies, form.PolicyID)
	case isBackspaceKey(msg):
		form = deleteNamespaceFormRune(form)
	default:
		if msg.Type == tea.KeyRunes {
			form = appendNamespaceFormRunes(form, msg.Runes)
		}
	}

	return normalizeNamespaceForm(form, policies), stackFormActionNone
}

func (m model) finishStackFormAction(action stackFormAction, save func() tea.Cmd) (tea.Model, tea.Cmd) {
	switch action {
	case stackFormActionQuit:
		return m, tea.Quit
	case stackFormActionSave:
		return m, save()
	default:
		return m, nil
	}
}

func (m model) updateDependencyForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var action stackFormAction
	m.dependencyForm, action = updateDependencyFormState(msg, m.dependencyForm, m.stacks, registryUISources(m.sources))
	return m.finishStackFormAction(action, m.saveDependency)
}

func (m model) updateProjectForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var action stackFormAction
	m.projectForm, action = updateProjectFormState(msg, m.projectForm, m.namespaces, vcsSources(m.sources), m.stacks)
	return m.finishStackFormAction(action, m.saveProject)
}

func (m model) updateSourceForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var action stackFormAction
	m.sourceForm, action = updateSourceFormState(msg, m.sourceForm)
	return m.finishStackFormAction(action, m.saveSource)
}

func (m model) updatePolicyForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var action stackFormAction
	m.policyForm, action = updatePolicyFormState(msg, m.policyForm, m.namespaces)
	return m.finishStackFormAction(action, m.savePolicy)
}

func (m model) updatePolicyValueForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if ui.KeyUpdatePins.Matches(msg.String()) {
		if !m.policyValueForm.CanUpdate {
			return m, nil
		}
		return m, m.loadPolicyValueLatest()
	}
	var action stackFormAction
	m.policyValueForm, action = updatePolicyValueFormState(msg, m.policyValueForm, m.dependencies, registryUISources(m.sources))
	return m.finishStackFormAction(action, m.savePolicyValue)
}

func (m model) updatePolicyUpdateForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var action stackFormAction
	m.policyUpdateForm, action = updatePolicyUpdateFormState(msg, m.policyUpdateForm, policyStacks(m.policyValues, m.stacks), registryUISources(m.sources))
	switch action {
	case stackFormActionQuit:
		return m, tea.Quit
	case stackFormActionSave:
		form := m.policyUpdateForm
		registrySource, ok := findByID(registryUISources(m.sources), form.RegistryID, sourceID)
		if !ok {
			m.policyUpdateForm.Error = "Add registry source first"
			return m, nil
		}
		m.policyUpdateForm = ui.PolicyUpdateForm{}
		m.policyUpdateStatus = ui.SettingsStatus{
			Message: "Updating pinned deps",
			Running: true,
		}
		return m, m.policyPinsUpdateCmd(form, registrySource)
	default:
		return m, nil
	}
}

func updateDependencyFormState(msg tea.KeyMsg, form ui.DependencyForm, stacks []ui.Stack, sources []ui.Source) (ui.DependencyForm, stackFormAction) {
	switch {
	case isQuitKey(msg):
		return form, stackFormActionQuit
	case isCancelKey(msg):
		return ui.DependencyForm{}, stackFormActionNone
	case isEnterKey(msg):
		if !form.CanSave {
			return form, stackFormActionNone
		}
		return form, stackFormActionSave
	case isOneOf(msg, "tab", "down"):
		form.Focus = nextDependencyFormField(form.Focus)
	case isOneOf(msg, "shift+tab", "up"):
		form.Focus = previousDependencyFormField(form.Focus)
	case form.Focus == ui.DependencyFormFieldStack && isOneOf(msg, "left", "h"):
		form.StackID = previousStackID(stacks, form.StackID)
	case form.Focus == ui.DependencyFormFieldStack && isOneOf(msg, "right", "l"):
		form.StackID = nextStackID(stacks, form.StackID)
	case form.Focus == ui.DependencyFormFieldRegistry && isOneOf(msg, "left", "h"):
		form.RegistryID = previousSourceID(sources, form.RegistryID)
	case form.Focus == ui.DependencyFormFieldRegistry && isOneOf(msg, "right", "l"):
		form.RegistryID = nextSourceID(sources, form.RegistryID)
	case isBackspaceKey(msg):
		form = deleteDependencyFormRune(form)
	default:
		if msg.Type == tea.KeyRunes {
			form = appendDependencyFormRunes(form, msg.Runes)
		}
	}

	return normalizeDependencyForm(form, stacks, sources), stackFormActionNone
}

func updateProjectFormState(msg tea.KeyMsg, form ui.ProjectForm, namespaces []ui.Namespace, sources []ui.Source, stacks []ui.Stack) (ui.ProjectForm, stackFormAction) {
	switch {
	case isQuitKey(msg):
		return form, stackFormActionQuit
	case isCancelKey(msg):
		return ui.ProjectForm{}, stackFormActionNone
	case isEnterKey(msg):
		if !form.CanSave {
			return form, stackFormActionNone
		}
		return form, stackFormActionSave
	case isOneOf(msg, "tab", "down"):
		form.Focus = nextProjectFormField(form.Focus)
	case isOneOf(msg, "shift+tab", "up"):
		form.Focus = previousProjectFormField(form.Focus)
	case isProjectPickerFocused(form.Focus) && isOneOf(msg, "left", "h"):
		switch form.Focus {
		case ui.ProjectFormFieldNamespace:
			form.NamespaceID = previousNamespaceID(namespaces, form.NamespaceID)
		case ui.ProjectFormFieldSource:
			form.SourceID = previousSourceID(sources, form.SourceID)
		case ui.ProjectFormFieldStack:
			form.StackID = previousStackID(stacks, form.StackID)
		}
	case isProjectPickerFocused(form.Focus) && isOneOf(msg, "right", "l"):
		switch form.Focus {
		case ui.ProjectFormFieldNamespace:
			form.NamespaceID = nextNamespaceID(namespaces, form.NamespaceID)
		case ui.ProjectFormFieldSource:
			form.SourceID = nextSourceID(sources, form.SourceID)
		case ui.ProjectFormFieldStack:
			form.StackID = nextStackID(stacks, form.StackID)
		}
	case isOneOf(msg, " ", "space"):
		switch form.Focus {
		case ui.ProjectFormFieldFreezing:
			form.Freezing = !form.Freezing
		case ui.ProjectFormFieldEndOfLife:
			form.EndOfLife = !form.EndOfLife
		}
	case isBackspaceKey(msg):
		form = deleteProjectFormRune(form)
	default:
		if msg.Type == tea.KeyRunes {
			form = appendProjectFormRunes(form, msg.Runes)
		}
	}

	return normalizeProjectForm(form, namespaces, sources, stacks), stackFormActionNone
}

func updateSourceFormState(msg tea.KeyMsg, form ui.SourceForm) (ui.SourceForm, stackFormAction) {
	switch {
	case isQuitKey(msg):
		return form, stackFormActionQuit
	case isCancelKey(msg):
		return ui.SourceForm{}, stackFormActionNone
	case isEnterKey(msg):
		if !form.CanSave {
			return form, stackFormActionNone
		}
		return form, stackFormActionSave
	case isOneOf(msg, "tab", "down"):
		form.Focus = nextSourceFormFieldForForm(form)
	case isOneOf(msg, "shift+tab", "up"):
		form.Focus = previousSourceFormFieldForForm(form)
	case form.Focus == ui.SourceFormFieldType && isOneOf(msg, "left", "h"):
		form.Type = previousSourceType(form.Type)
	case form.Focus == ui.SourceFormFieldType && isOneOf(msg, "right", "l"):
		form.Type = nextSourceType(form.Type)
	case form.Focus == ui.SourceFormFieldRegistryKind && isOneOf(msg, "left", "h"):
		form.RegistryKind = previousRegistryKind(form.RegistryKind)
	case form.Focus == ui.SourceFormFieldRegistryKind && isOneOf(msg, "right", "l"):
		form.RegistryKind = nextRegistryKind(form.RegistryKind)
	case isBackspaceKey(msg):
		form = deleteSourceFormRune(form)
	default:
		if msg.Type == tea.KeyRunes {
			form = appendSourceFormRunes(form, msg.Runes)
		}
	}

	return normalizeSourceForm(form), stackFormActionNone
}

func updatePolicyFormState(msg tea.KeyMsg, form ui.PolicyForm, namespaces []ui.Namespace) (ui.PolicyForm, stackFormAction) {
	switch {
	case isQuitKey(msg):
		return form, stackFormActionQuit
	case isCancelKey(msg):
		return ui.PolicyForm{}, stackFormActionNone
	case isEnterKey(msg):
		if !form.CanSave {
			return form, stackFormActionNone
		}
		return form, stackFormActionSave
	case isOneOf(msg, "tab", "down"):
		form.Focus = nextPolicyFormField(form.Focus)
	case isOneOf(msg, "shift+tab", "up"):
		form.Focus = previousPolicyFormField(form.Focus)
	case form.Focus == ui.PolicyFormFieldNamespace && isOneOf(msg, "left", "h"):
		form.NamespaceID = previousNamespaceID(namespaces, form.NamespaceID)
	case form.Focus == ui.PolicyFormFieldNamespace && isOneOf(msg, "right", "l"):
		form.NamespaceID = nextNamespaceID(namespaces, form.NamespaceID)
	case isBackspaceKey(msg):
		form = deletePolicyFormRune(form)
	default:
		if msg.Type == tea.KeyRunes {
			form = appendPolicyFormRunes(form, msg.Runes)
		}
	}

	return normalizePolicyForm(form, namespaces), stackFormActionNone
}

func updatePolicyValueFormState(msg tea.KeyMsg, form ui.PolicyValueForm, dependencies []ui.Dependency, sources []ui.Source) (ui.PolicyValueForm, stackFormAction) {
	switch {
	case isQuitKey(msg):
		return form, stackFormActionQuit
	case isCancelKey(msg):
		return ui.PolicyValueForm{}, stackFormActionNone
	case isEnterKey(msg):
		if !form.CanSave {
			return form, stackFormActionNone
		}
		return form, stackFormActionSave
	case isOneOf(msg, "tab", "down"):
		form.Focus = nextPolicyValueFormField(form.Focus)
	case isOneOf(msg, "shift+tab", "up"):
		form.Focus = previousPolicyValueFormField(form.Focus)
	case form.Focus == ui.PolicyValueFormFieldDependency && isOneOf(msg, "left", "h"):
		form.DependencyID = previousDependencyID(dependencies, form.DependencyID)
	case form.Focus == ui.PolicyValueFormFieldDependency && isOneOf(msg, "right", "l"):
		form.DependencyID = nextDependencyID(dependencies, form.DependencyID)
	case form.Focus == ui.PolicyValueFormFieldRegistry && isOneOf(msg, "left", "h"):
		form.RegistryID = previousSourceID(sources, form.RegistryID)
	case form.Focus == ui.PolicyValueFormFieldRegistry && isOneOf(msg, "right", "l"):
		form.RegistryID = nextSourceID(sources, form.RegistryID)
	case isBackspaceKey(msg):
		form = deletePolicyValueFormRune(form)
	default:
		if msg.Type == tea.KeyRunes {
			form = appendPolicyValueFormRunes(form, msg.Runes)
		}
	}

	return normalizePolicyValueForm(form, dependencies, sources), stackFormActionNone
}

func updatePolicyUpdateFormState(msg tea.KeyMsg, form ui.PolicyUpdateForm, stacks []ui.Stack, sources []ui.Source) (ui.PolicyUpdateForm, stackFormAction) {
	switch {
	case isQuitKey(msg):
		return form, stackFormActionQuit
	case isCancelKey(msg):
		return ui.PolicyUpdateForm{}, stackFormActionNone
	case isEnterKey(msg):
		if !form.CanStart {
			return form, stackFormActionNone
		}
		return form, stackFormActionSave
	case isOneOf(msg, "tab", "down"):
		form.Focus = nextPolicyUpdateFormField(form.Focus)
	case isOneOf(msg, "shift+tab", "up"):
		form.Focus = previousPolicyUpdateFormField(form.Focus)
	case form.Focus == ui.PolicyUpdateFormFieldStack && isOneOf(msg, "left", "h"):
		form.StackID = previousStackID(stacks, form.StackID)
	case form.Focus == ui.PolicyUpdateFormFieldStack && isOneOf(msg, "right", "l"):
		form.StackID = nextStackID(stacks, form.StackID)
	case form.Focus == ui.PolicyUpdateFormFieldRegistry && isOneOf(msg, "left", "h"):
		form.RegistryID = previousSourceID(sources, form.RegistryID)
	case form.Focus == ui.PolicyUpdateFormFieldRegistry && isOneOf(msg, "right", "l"):
		form.RegistryID = nextSourceID(sources, form.RegistryID)
	}

	return normalizePolicyUpdateForm(form, stacks, sources), stackFormActionNone
}

func (m model) updateDeleteConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var action deleteConfirmAction
	m.deleteConfirm, action = updateDeleteConfirmState(msg, m.deleteConfirm)
	return m.finishDeleteConfirmAction(action, m.deleteStack)
}

func (m model) updateNamespaceDeleteConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var action deleteConfirmAction
	m.namespaceDeleteConfirm, action = updateDeleteConfirmState(msg, m.namespaceDeleteConfirm)
	return m.finishDeleteConfirmAction(action, m.deleteNamespace)
}

func (m model) updateDependencyDeleteConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var action deleteConfirmAction
	m.dependencyDeleteConfirm, action = updateDeleteConfirmState(msg, m.dependencyDeleteConfirm)
	return m.finishDeleteConfirmAction(action, m.deleteDependency)
}

func (m model) updateProjectDeleteConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var action deleteConfirmAction
	m.projectDeleteConfirm, action = updateDeleteConfirmState(msg, m.projectDeleteConfirm)
	return m.finishDeleteConfirmAction(action, m.deleteProject)
}

func (m model) updateSourceDeleteConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var action deleteConfirmAction
	m.sourceDeleteConfirm, action = updateDeleteConfirmState(msg, m.sourceDeleteConfirm)
	return m.finishDeleteConfirmAction(action, m.deleteSource)
}

func (m model) updatePolicyDeleteConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var action deleteConfirmAction
	m.policyDeleteConfirm, action = updateDeleteConfirmState(msg, m.policyDeleteConfirm)
	if m.policyFocus == ui.PolicyPaneValues {
		return m.finishDeleteConfirmAction(action, m.deletePolicyValue)
	}
	return m.finishDeleteConfirmAction(action, m.deletePolicy)
}

type deleteConfirmAction int

const (
	deleteConfirmActionNone deleteConfirmAction = iota
	deleteConfirmActionQuit
	deleteConfirmActionDelete
)

func updateDeleteConfirmState(msg tea.KeyMsg, confirm ui.DeleteConfirm) (ui.DeleteConfirm, deleteConfirmAction) {
	switch {
	case isQuitKey(msg):
		return confirm, deleteConfirmActionQuit
	case isCancelKey(msg):
		return ui.DeleteConfirm{}, deleteConfirmActionNone
	case isEnterKey(msg):
		return confirm, deleteConfirmActionDelete
	}

	return confirm, deleteConfirmActionNone
}

func (m model) finishDeleteConfirmAction(action deleteConfirmAction, performDelete func() tea.Cmd) (tea.Model, tea.Cmd) {
	switch action {
	case deleteConfirmActionQuit:
		return m, tea.Quit
	case deleteConfirmActionDelete:
		return m, performDelete()
	default:
		return m, nil
	}
}

func (m *model) openAddForm() {
	switch m.screen {
	case ui.ScreenStacks:
		m.form = newStackForm()
	case ui.ScreenNamespaces:
		m.namespaceForm = newNamespaceForm(m.policies)
	case ui.ScreenDependencies:
		m.dependencyForm = newDependencyForm(m.stacks, m.sources)
	case ui.ScreenProjects:
		m.projectForm = newProjectForm(m.namespaces, m.sources, m.stacks)
	case ui.ScreenSources:
		m.sourceForm = newSourceForm()
	case ui.ScreenPolicies:
		if m.policyFocus == ui.PolicyPaneValues {
			if m.selectedPolicyID == 0 {
				return
			}
			m.policyValueForm = newPolicyValueForm(m.selectedPolicyID, m.dependencies, m.sources)
			return
		}
		m.policyForm = newPolicyForm(m.namespaces, m.policies)
	}
}

func (m *model) openEditForm() {
	switch m.screen {
	case ui.ScreenStacks:
		m.openEditStackForm()
	case ui.ScreenNamespaces:
		m.openEditNamespaceForm()
	case ui.ScreenDependencies:
		m.openEditDependencyForm()
	case ui.ScreenProjects:
		m.openEditProjectForm()
	case ui.ScreenSources:
		m.openEditSourceForm()
	case ui.ScreenPolicies:
		if m.policyFocus == ui.PolicyPaneValues {
			m.openEditPolicyValueForm()
			return
		}
		m.openEditPolicyForm()
	}
}

func (m *model) openDelete() {
	switch m.screen {
	case ui.ScreenStacks:
		m.openDeleteConfirm()
	case ui.ScreenNamespaces:
		m.openNamespaceDeleteConfirm()
	case ui.ScreenDependencies:
		m.openDependencyDeleteConfirm()
	case ui.ScreenProjects:
		m.openProjectDeleteConfirm()
	case ui.ScreenSources:
		m.openSourceDeleteConfirm()
	case ui.ScreenPolicies:
		if m.policyFocus == ui.PolicyPaneValues {
			m.openPolicyValueDeleteConfirm()
			return
		}
		m.openPolicyDeleteConfirm()
	}
}

func (m *model) selectPreviousOnScreen() {
	switch m.screen {
	case ui.ScreenDefault:
		if m.dashboard.focus == ui.DashboardPaneAttention {
			m.selectPreviousAttentionProject()
		}
	case ui.ScreenStacks:
		m.selectPreviousStack()
	case ui.ScreenNamespaces:
		m.selectPreviousNamespace()
	case ui.ScreenDependencies:
		m.selectPreviousDependency()
	case ui.ScreenProjects:
		if m.projectFocus == ui.ProjectPaneDependencies {
			m.selectPreviousProjectDependency()
			return
		}
		m.selectPreviousProject()
	case ui.ScreenSources:
		m.selectPreviousSource()
	case ui.ScreenPolicies:
		if m.policyFocus == ui.PolicyPaneValues {
			m.selectPreviousPolicyValue()
			return
		}
		m.selectPreviousPolicy()
	case ui.ScreenSettings:
		m.selectPreviousCheckProject()
	case ui.ScreenReleases:
		m.selectPreviousReleaseProject()
	case ui.ScreenVulnerabilities:
		if m.vulnFocus == ui.VulnPaneDetails {
			m.selectPreviousVulnItem()
			return
		}
		m.selectPreviousVulnProject()
	}
}

func (m *model) selectNextOnScreen() {
	switch m.screen {
	case ui.ScreenDefault:
		if m.dashboard.focus == ui.DashboardPaneAttention {
			m.selectNextAttentionProject()
		}
	case ui.ScreenStacks:
		m.selectNextStack()
	case ui.ScreenNamespaces:
		m.selectNextNamespace()
	case ui.ScreenDependencies:
		m.selectNextDependency()
	case ui.ScreenProjects:
		if m.projectFocus == ui.ProjectPaneDependencies {
			m.selectNextProjectDependency()
			return
		}
		m.selectNextProject()
	case ui.ScreenSources:
		m.selectNextSource()
	case ui.ScreenPolicies:
		if m.policyFocus == ui.PolicyPaneValues {
			m.selectNextPolicyValue()
			return
		}
		m.selectNextPolicy()
	case ui.ScreenSettings:
		m.selectNextCheckProject()
	case ui.ScreenReleases:
		m.selectNextReleaseProject()
	case ui.ScreenVulnerabilities:
		if m.vulnFocus == ui.VulnPaneDetails {
			m.selectNextVulnItem()
			return
		}
		m.selectNextVulnProject()
	}
}

func (m *model) togglePolicyPane() {
	if m.screen != ui.ScreenPolicies {
		return
	}
	if m.selectedPolicyID == 0 {
		m.policyFocus = ui.PolicyPanePolicies
		return
	}
	if m.policyFocus == ui.PolicyPanePolicies {
		m.policyFocus = ui.PolicyPaneValues
		return
	}
	m.policyFocus = ui.PolicyPanePolicies
}

func (m *model) toggleDashboardPane() {
	if m.screen != ui.ScreenDefault {
		return
	}
	if m.dashboard.focus == ui.DashboardPaneSummary {
		m.dashboard.focus = ui.DashboardPaneAttention
		m.ensureSelectedAttentionProject()
		return
	}
	m.dashboard.focus = ui.DashboardPaneSummary
}

func (m *model) toggleProjectPane() {
	if m.screen != ui.ScreenProjects {
		return
	}
	if m.projectFocus == ui.ProjectPaneProjects {
		m.projectFocus = ui.ProjectPaneDependencies
		m.ensureSelectedProjectDependency()
		return
	}
	m.projectFocus = ui.ProjectPaneProjects
}

func (m model) startProjectDependencySync() (tea.Model, tea.Cmd) {
	if m.projectSyncStatus.Running {
		return m, nil
	}

	project, ok := m.selectedProject()
	if !ok {
		m.projectSyncStatus = ui.ProjectSyncStatus{Error: "project is not selected"}
		return m, nil
	}
	source, ok := findByID(m.sources, project.SourceID, sourceID)
	if !ok {
		m.projectSyncStatus = ui.ProjectSyncStatus{ProjectID: project.ID, Error: "project source is not found"}
		return m, nil
	}

	ch := make(chan projectSyncMsg, 16)
	m.projectSyncCh = ch
	m.projectSyncStatus = ui.ProjectSyncStatus{
		ProjectID: project.ID,
		Message:   "Resolving commit...",
		Running:   true,
	}

	return m, tea.Batch(runProjectDependencySync(m.store, project, source, ch), waitProjectSync(ch))
}

func runProjectDependencySync(store *storage.Store, project ui.Project, source ui.Source, ch chan<- projectSyncMsg) tea.Cmd {
	return func() tea.Msg {
		defer close(ch)
		if store == nil {
			ch <- projectSyncMsg{projectID: project.ID, message: "store is not ready", err: errors.New("store is not ready"), done: true}
			return nil
		}

		client, err := projectsync.NewSourceClient(source.Type, nil)
		if err != nil {
			ch <- projectSyncMsg{projectID: project.ID, message: err.Error(), err: err, done: true}
			return nil
		}

		service := projectsync.Service{
			Cache:        store.Cache(),
			Runs:         store.ProjectDependencies(),
			SourceClient: client,
			Force:        true,
		}
		result, err := service.Sync(context.Background(), projectsync.Source{
			ID:       source.ID,
			Type:     source.Type,
			URL:      source.URL,
			PATToken: source.PATToken,
		}, projectsync.Project{
			ID:         project.ID,
			ProviderID: project.ProjectID,
			Name:       project.Name,
			StackName:  project.StackName,
		}, func(message string) {
			ch <- projectSyncMsg{projectID: project.ID, message: message}
		})
		if err != nil {
			ch <- projectSyncMsg{projectID: project.ID, message: err.Error(), err: err, done: true}
			return nil
		}

		message := "Saved " + strconv.Itoa(result.Count) + " dependencies"
		if result.UpToDate {
			message = "already up to date"
		}
		ch <- projectSyncMsg{projectID: project.ID, message: message, done: true}
		return nil
	}
}

func waitProjectSync(ch <-chan projectSyncMsg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return nil
		}

		return msg
	}
}

func (m model) reloadAfterProjectSync() tea.Cmd {
	if m.screen == ui.ScreenView {
		return tea.Batch(m.loadProjects(), m.loadDependencyView())
	}

	return m.loadProjects()
}

func (m model) startAllProjectsDependencySync() (tea.Model, tea.Cmd) {
	return m.startProjectsDependencySync(activeProjects(m.projects), "Syncing all projects...")
}

func (m model) startViewDependencySync(all bool) (tea.Model, tea.Cmd) {
	message := "Syncing project..."
	if all {
		message = "Syncing all projects..."
	}

	return m.startProjectsDependencySync(m.viewProjects(all), message)
}

func (m model) startProjectsDependencySync(projects []ui.Project, message string) (tea.Model, tea.Cmd) {
	if m.projectSyncStatus.Running {
		return m, nil
	}
	if len(projects) == 0 {
		m.projectSyncStatus = ui.ProjectSyncStatus{Error: "project is not selected"}
		return m, nil
	}

	ch := make(chan projectSyncMsg, 16)
	m.projectSyncCh = ch
	m.projectSyncStatus = ui.ProjectSyncStatus{
		Message: message,
		Running: true,
	}

	return m, tea.Batch(runProjectsDependencySync(m.store, projects, m.sources, ch), waitProjectSync(ch))
}

func (m model) viewProjects(all bool) []ui.Project {
	if all {
		projects := make([]ui.Project, 0, len(m.dependencyView.Rows))
		for _, row := range m.dependencyView.Rows {
			if project, ok := findByID(m.projects, row.ProjectID, projectID); ok {
				projects = append(projects, project)
			}
		}

		return projects
	}
	if project, ok := findByID(m.projects, m.selectedViewProjectID, projectID); ok && !project.EndOfLife {
		return []ui.Project{project}
	}

	return nil
}

func activeProjects(projects []ui.Project) []ui.Project {
	result := make([]ui.Project, 0, len(projects))
	for _, project := range projects {
		if !project.EndOfLife {
			result = append(result, project)
		}
	}

	return result
}

func runProjectsDependencySync(store *storage.Store, projects []ui.Project, sources []ui.Source, ch chan<- projectSyncMsg) tea.Cmd {
	return func() tea.Msg {
		defer close(ch)
		if store == nil {
			ch <- projectSyncMsg{message: "store is not ready", err: errors.New("store is not ready"), done: true}
			return nil
		}

		total := len(projects)
		for index, project := range projects {
			source, ok := findByID(sources, project.SourceID, sourceID)
			if !ok {
				ch <- projectSyncMsg{message: "Skipping " + project.Name + ": source not found", current: index, total: total}
				continue
			}

			client, err := projectsync.NewSourceClient(source.Type, nil)
			if err != nil {
				ch <- projectSyncMsg{message: err.Error(), err: err, done: true}
				return nil
			}

			service := projectsync.Service{
				Cache:        store.Cache(),
				Runs:         store.ProjectDependencies(),
				SourceClient: client,
				Force:        true,
			}
			_, err = service.Sync(context.Background(), projectsync.Source{
				ID:       source.ID,
				Type:     source.Type,
				URL:      source.URL,
				PATToken: source.PATToken,
			}, projectsync.Project{
				ID:         project.ID,
				ProviderID: project.ProjectID,
				Name:       project.Name,
				StackName:  project.StackName,
			}, func(message string) {
				ch <- projectSyncMsg{message: project.Name + ": " + message, current: index, total: total}
			})
			if err != nil {
				ch <- projectSyncMsg{message: err.Error(), err: err, done: true}
				return nil
			}

			ch <- projectSyncMsg{message: project.Name, step: true, current: index + 1, total: total}
		}

		ch <- projectSyncMsg{message: "Sync complete", done: true, current: total, total: total}
		return nil
	}
}

func (m model) loadStacks() tea.Cmd {
	return func() tea.Msg {
		if m.store == nil {
			return stacksLoadedMsg{stacks: []ui.Stack{}}
		}

		stacks, err := m.store.Stacks().List(context.Background())
		if err != nil {
			return stacksLoadedMsg{err: err}
		}

		return stacksLoadedMsg{stacks: toUIStacks(stacks)}
	}
}

func (m model) loadNamespaces() tea.Cmd {
	return func() tea.Msg {
		if m.store == nil {
			return namespacesLoadedMsg{namespaces: []ui.Namespace{}}
		}

		namespaces, err := m.store.Namespaces().List(context.Background())
		if err != nil {
			return namespacesLoadedMsg{err: err}
		}

		return namespacesLoadedMsg{namespaces: toUINamespaces(namespaces)}
	}
}

func (m model) loadDependencies() tea.Cmd {
	return func() tea.Msg {
		if m.store == nil {
			return dependenciesLoadedMsg{dependencies: []ui.Dependency{}}
		}

		dependencies, err := m.store.Dependencies().List(context.Background())
		if err != nil {
			return dependenciesLoadedMsg{err: err}
		}

		return dependenciesLoadedMsg{dependencies: toUIDependencies(dependencies)}
	}
}

func (m model) loadProjects() tea.Cmd {
	return func() tea.Msg {
		if m.store == nil {
			return projectsLoadedMsg{projects: []ui.Project{}}
		}

		projects, err := m.store.Projects().List(context.Background())
		if err != nil {
			return projectsLoadedMsg{err: err}
		}

		return projectsLoadedMsg{projects: toUIProjects(projects)}
	}
}

func (m model) loadProjectDependencies() tea.Cmd {
	projectID := m.selectedProjectID
	return func() tea.Msg {
		if m.store == nil || projectID == 0 {
			return projectDependenciesLoadedMsg{dependencies: []ui.ProjectDependency{}}
		}

		dependencies, err := m.store.ProjectDependencies().ListByProject(context.Background(), projectID)
		if err != nil {
			return projectDependenciesLoadedMsg{err: err}
		}
		latestRun, err := m.store.ProjectDependencies().LatestRun(context.Background(), projectID)
		if err != nil {
			return projectDependenciesLoadedMsg{err: err}
		}

		return projectDependenciesLoadedMsg{
			dependencies: toUIProjectDependencies(dependencies),
			latestRun:    toUIProjectDependencyRun(latestRun),
		}
	}
}

func (m model) loadSources() tea.Cmd {
	return func() tea.Msg {
		if m.store == nil {
			return sourcesLoadedMsg{sources: []ui.Source{}}
		}

		sources, err := m.store.Sources().List(context.Background())
		if err != nil {
			return sourcesLoadedMsg{err: err}
		}

		return sourcesLoadedMsg{sources: toUISources(sources)}
	}
}

func (m model) loadPolicies() tea.Cmd {
	return func() tea.Msg {
		if m.store == nil {
			return policiesLoadedMsg{policies: []ui.Policy{}}
		}

		policies, err := m.store.Policies().List(context.Background())
		if err != nil {
			return policiesLoadedMsg{err: err}
		}

		return policiesLoadedMsg{policies: toUIPolicies(policies)}
	}
}

func (m model) loadPolicyValues() tea.Cmd {
	policyID := m.selectedPolicyID
	return func() tea.Msg {
		if m.store == nil {
			return policyValuesLoadedMsg{values: []ui.PolicyValue{}}
		}

		values, err := m.store.Policies().ListValues(context.Background(), policyID)
		if err != nil {
			return policyValuesLoadedMsg{err: err}
		}

		return policyValuesLoadedMsg{values: toUIPolicyValues(values)}
	}
}

func (m model) loadDependencyView() tea.Cmd {
	stackID := m.viewStackID
	return func() tea.Msg {
		if m.store == nil || stackID == 0 {
			return dependencyViewLoadedMsg{view: ui.DependencyView{StackID: stackID}}
		}

		view, err := m.store.ProjectDependencies().ViewByStack(context.Background(), stackID)
		if err != nil {
			return dependencyViewLoadedMsg{err: err}
		}

		return dependencyViewLoadedMsg{view: toUIDependencyView(view)}
	}
}

func (m model) saveStack() tea.Cmd {
	form := m.form
	return func() tea.Msg {
		if m.store == nil {
			return stackSavedMsg{stackID: form.StackID}
		}

		icon := strings.TrimSpace(form.Icon)
		name := strings.TrimSpace(form.Name)
		color := strings.TrimSpace(form.Color)
		if form.Mode == ui.StackFormModeEdit {
			err := m.store.Stacks().Update(context.Background(), form.StackID, icon, name, color)
			return stackSavedMsg{stackID: form.StackID, err: err}
		}

		stack, err := m.store.Stacks().Create(context.Background(), icon, name, color)
		return stackSavedMsg{stackID: stack.ID, err: err}
	}
}

func (m model) saveNamespace() tea.Cmd {
	form := m.namespaceForm
	return func() tea.Msg {
		if m.store == nil {
			return namespaceSavedMsg{namespaceID: form.StackID}
		}

		icon := strings.TrimSpace(form.Icon)
		name := strings.TrimSpace(form.Name)
		color := strings.TrimSpace(form.Color)
		if form.Mode == ui.StackFormModeEdit {
			if err := m.store.Namespaces().Update(context.Background(), form.StackID, icon, name, color); err != nil {
				return namespaceSavedMsg{namespaceID: form.StackID, err: err}
			}
			err := m.store.Namespaces().SetPolicy(context.Background(), form.StackID, form.PolicyID)
			return namespaceSavedMsg{namespaceID: form.StackID, err: err}
		}

		namespace, err := m.store.Namespaces().Create(context.Background(), icon, name, color)
		if err != nil {
			return namespaceSavedMsg{err: err}
		}
		err = m.store.Namespaces().SetPolicy(context.Background(), namespace.ID, form.PolicyID)
		return namespaceSavedMsg{namespaceID: namespace.ID, err: err}
	}
}

func (m model) saveDependency() tea.Cmd {
	form := m.dependencyForm
	return func() tea.Msg {
		if m.store == nil {
			return dependencySavedMsg{dependencyID: form.DependencyID}
		}

		icon := strings.TrimSpace(form.Icon)
		name := strings.TrimSpace(form.Name)
		color := strings.TrimSpace(form.Color)
		registryName := strings.TrimSpace(form.RegistryName)
		registrySourceID := form.RegistryID
		if form.Mode == ui.StackFormModeEdit {
			err := m.store.Dependencies().UpdateWithRegistry(context.Background(), form.DependencyID, form.StackID, icon, name, color, registryName, registrySourceID)
			return dependencySavedMsg{dependencyID: form.DependencyID, err: err}
		}

		dependency, err := m.store.Dependencies().CreateWithRegistry(context.Background(), form.StackID, icon, name, color, registryName, registrySourceID)
		return dependencySavedMsg{dependencyID: dependency.ID, err: err}
	}
}

func (m model) saveProject() tea.Cmd {
	form := m.projectForm
	return func() tea.Msg {
		if m.store == nil {
			return projectSavedMsg{projectID: form.ID}
		}

		icon := strings.TrimSpace(form.Icon)
		name := strings.TrimSpace(form.Name)
		color := strings.TrimSpace(form.Color)
		projectID := strings.TrimSpace(form.ProjectID)
		if projectID == "" {
			return projectSavedMsg{projectID: form.ID, err: errors.New("project PROJECT_ID is empty")}
		}
		if form.Mode == ui.StackFormModeEdit {
			err := m.store.Projects().Update(context.Background(), form.ID, projectID, form.NamespaceID, form.SourceID, form.StackID, icon, name, color, form.Freezing, form.EndOfLife)
			return projectSavedMsg{projectID: form.ID, err: err}
		}

		project, err := m.store.Projects().Create(context.Background(), projectID, form.NamespaceID, form.SourceID, form.StackID, icon, name, color, form.Freezing, form.EndOfLife)
		return projectSavedMsg{projectID: project.ID, err: err}
	}
}

func (m model) saveSource() tea.Cmd {
	form := m.sourceForm
	return func() tea.Msg {
		if m.store == nil {
			return sourceSavedMsg{sourceID: form.SourceID}
		}

		name := strings.TrimSpace(form.Name)
		patToken := strings.TrimSpace(form.PATToken)
		sourceURL := strings.TrimSpace(form.URL)
		sourceType := strings.TrimSpace(form.Type)
		registryKind := strings.TrimSpace(form.RegistryKind)
		if form.Mode == ui.StackFormModeEdit {
			err := m.store.Sources().UpdateWithRegistryKind(context.Background(), form.SourceID, name, patToken, sourceURL, sourceType, registryKind)
			return sourceSavedMsg{sourceID: form.SourceID, err: err}
		}

		source, err := m.store.Sources().CreateWithRegistryKind(context.Background(), name, patToken, sourceURL, sourceType, registryKind)
		return sourceSavedMsg{sourceID: source.ID, err: err}
	}
}

func (m model) savePolicy() tea.Cmd {
	form := m.policyForm
	return func() tea.Msg {
		if m.store == nil {
			return policySavedMsg{policyID: form.PolicyID}
		}

		name := strings.TrimSpace(form.Name)
		if form.Mode == ui.StackFormModeEdit {
			err := m.store.Policies().Update(context.Background(), form.PolicyID, name)
			return policySavedMsg{policyID: form.PolicyID, err: err}
		}

		policy, err := m.store.Policies().Create(context.Background(), name)
		return policySavedMsg{policyID: policy.ID, err: err}
	}
}

func (m model) savePolicyValue() tea.Cmd {
	form := m.policyValueForm
	return func() tea.Msg {
		if m.store == nil {
			return policyValueSavedMsg{valueID: form.PolicyValueID}
		}

		version := strings.TrimSpace(form.Version)
		if form.Mode == ui.StackFormModeEdit {
			err := m.store.Policies().UpdateValue(context.Background(), form.PolicyValueID, form.DependencyID, version)
			return policyValueSavedMsg{valueID: form.PolicyValueID, err: err}
		}

		value, err := m.store.Policies().CreateValue(context.Background(), form.PolicyID, form.DependencyID, version)
		return policyValueSavedMsg{valueID: value.ID, err: err}
	}
}

func (m model) policyPinsUpdateCmd(form ui.PolicyUpdateForm, registrySource ui.Source) tea.Cmd {
	return func() tea.Msg {
		if m.store == nil {
			return policyPinsUpdatedMsg{}
		}
		values, err := m.store.Policies().ListValuesByStack(context.Background(), form.PolicyID, form.StackID)
		if err != nil {
			return policyPinsUpdatedMsg{err: err}
		}
		client := registry.Client{}
		source := registry.Source{
			URL:   registrySource.URL,
			Token: registrySource.PATToken,
			Kind:  registrySource.RegistryKind,
		}
		updated := 0
		failures := make([]string, 0)
		for _, value := range values {
			latest, err := client.LatestStable(context.Background(), source, value.RegistryName)
			if err != nil {
				failures = append(failures, value.DependencyName+": "+err.Error())
				continue
			}
			if latest == value.Version {
				continue
			}
			if err := m.store.Policies().UpdateValueVersion(context.Background(), value.ID, latest); err != nil {
				failures = append(failures, value.DependencyName+": "+err.Error())
				continue
			}
			updated++
		}
		return policyPinsUpdatedMsg{checked: len(values), updated: updated, failures: failures}
	}
}

func (m model) startSelectedPolicyValueUpdate() (tea.Model, tea.Cmd) {
	value, ok := m.selectedPolicyValue()
	if !ok {
		return m, nil
	}
	if value.RegistryID == 0 {
		m.policyUpdateStatus = ui.SettingsStatus{
			Message: "Pin update failed",
			Error:   value.DependencyName + ": registry is empty",
		}
		return m, nil
	}
	registrySource, ok := findByID(registryUISources(m.sources), value.RegistryID, sourceID)
	if !ok {
		m.policyUpdateStatus = ui.SettingsStatus{
			Message: "Pin update failed",
			Error:   value.DependencyName + ": registry not found",
		}
		return m, nil
	}
	m.policyUpdateStatus = ui.SettingsStatus{
		Message: "Updating " + value.DependencyName,
		Running: true,
		Current: 0,
		Total:   1,
	}
	return m, m.selectedPolicyValueUpdateCmd(value, registrySource)
}

func (m model) selectedPolicyValueUpdateCmd(value ui.PolicyValue, registrySource ui.Source) tea.Cmd {
	return func() tea.Msg {
		if m.store == nil {
			return policyPinsUpdatedMsg{checked: 1}
		}
		client := registry.Client{}
		latest, err := client.LatestStable(context.Background(), registry.Source{
			URL:   registrySource.URL,
			Token: registrySource.PATToken,
			Kind:  registrySource.RegistryKind,
		}, value.RegistryName)
		if err != nil {
			return policyPinsUpdatedMsg{checked: 1, err: fmt.Errorf("%s: %w", value.DependencyName, err)}
		}
		if latest == value.Version {
			return policyPinsUpdatedMsg{checked: 1, updated: 0}
		}
		if err := m.store.Policies().UpdateValueVersion(context.Background(), value.ID, latest); err != nil {
			return policyPinsUpdatedMsg{checked: 1, err: fmt.Errorf("%s: %w", value.DependencyName, err)}
		}
		return policyPinsUpdatedMsg{checked: 1, updated: 1}
	}
}

func (m model) loadPolicyValueLatest() tea.Cmd {
	form := m.policyValueForm
	dependency, ok := findByID(m.dependencies, form.DependencyID, dependencyID)
	if !ok {
		return func() tea.Msg {
			return policyValueLatestLoadedMsg{err: errors.New("dependency is empty")}
		}
	}
	registrySource, ok := findByID(registryUISources(m.sources), form.RegistryID, sourceID)
	if !ok {
		return func() tea.Msg {
			return policyValueLatestLoadedMsg{err: errors.New("registry is empty")}
		}
	}
	return func() tea.Msg {
		client := registry.Client{}
		version, err := client.LatestStable(context.Background(), registry.Source{
			URL:   registrySource.URL,
			Token: registrySource.PATToken,
			Kind:  registrySource.RegistryKind,
		}, dependency.RegistryName)
		return policyValueLatestLoadedMsg{version: version, err: err}
	}
}

func (m model) deleteStack() tea.Cmd {
	confirm := m.deleteConfirm
	return func() tea.Msg {
		if m.store == nil {
			return stackDeletedMsg{stackID: confirm.StackID}
		}

		err := m.store.Stacks().Delete(context.Background(), confirm.StackID)
		return stackDeletedMsg{stackID: confirm.StackID, err: err}
	}
}

func (m model) deleteNamespace() tea.Cmd {
	confirm := m.namespaceDeleteConfirm
	return func() tea.Msg {
		if m.store == nil {
			return namespaceDeletedMsg{namespaceID: confirm.StackID}
		}

		err := m.store.Namespaces().Delete(context.Background(), confirm.StackID)
		return namespaceDeletedMsg{namespaceID: confirm.StackID, err: err}
	}
}

func (m model) deleteDependency() tea.Cmd {
	confirm := m.dependencyDeleteConfirm
	return func() tea.Msg {
		if m.store == nil {
			return dependencyDeletedMsg{dependencyID: confirm.StackID}
		}

		err := m.store.Dependencies().Delete(context.Background(), confirm.StackID)
		return dependencyDeletedMsg{dependencyID: confirm.StackID, err: err}
	}
}

func (m model) deleteProject() tea.Cmd {
	confirm := m.projectDeleteConfirm
	return func() tea.Msg {
		if m.store == nil {
			return projectDeletedMsg{projectID: confirm.StackID}
		}

		err := m.store.Projects().Delete(context.Background(), confirm.StackID)
		return projectDeletedMsg{projectID: confirm.StackID, err: err}
	}
}

func (m model) deleteSource() tea.Cmd {
	confirm := m.sourceDeleteConfirm
	return func() tea.Msg {
		if m.store == nil {
			return sourceDeletedMsg{sourceID: confirm.StackID}
		}

		err := m.store.Sources().Delete(context.Background(), confirm.StackID)
		return sourceDeletedMsg{sourceID: confirm.StackID, err: err}
	}
}

func (m model) deletePolicy() tea.Cmd {
	confirm := m.policyDeleteConfirm
	return func() tea.Msg {
		if m.store == nil {
			return policyDeletedMsg{policyID: confirm.StackID}
		}

		err := m.store.Policies().Delete(context.Background(), confirm.StackID)
		return policyDeletedMsg{policyID: confirm.StackID, err: err}
	}
}

func (m model) deletePolicyValue() tea.Cmd {
	confirm := m.policyDeleteConfirm
	return func() tea.Msg {
		if m.store == nil {
			return policyValueDeletedMsg{valueID: confirm.StackID}
		}

		err := m.store.Policies().DeleteValue(context.Background(), confirm.StackID)
		return policyValueDeletedMsg{valueID: confirm.StackID, err: err}
	}
}

func toUIStacks(stacks []storage.Stack) []ui.Stack {
	result := make([]ui.Stack, 0, len(stacks))
	for _, stack := range stacks {
		result = append(result, ui.Stack{
			ID:    stack.ID,
			Icon:  stack.Icon,
			Name:  stack.Name,
			Color: stack.Color,
		})
	}

	return result
}

func toUINamespaces(namespaces []storage.Namespace) []ui.Namespace {
	result := make([]ui.Namespace, 0, len(namespaces))
	for _, namespace := range namespaces {
		result = append(result, ui.Namespace{
			ID:       namespace.ID,
			Icon:     namespace.Icon,
			Name:     namespace.Name,
			Color:    namespace.Color,
			PolicyID: namespace.PolicyID,
		})
	}

	return result
}

func toUIDependencies(dependencies []storage.Dependency) []ui.Dependency {
	result := make([]ui.Dependency, 0, len(dependencies))
	for _, dependency := range dependencies {
		result = append(result, ui.Dependency{
			ID:                 dependency.ID,
			StackID:            dependency.StackID,
			StackName:          dependency.StackName,
			Icon:               dependency.Icon,
			Name:               dependency.Name,
			Color:              dependency.Color,
			RegistryName:       dependency.RegistryName,
			RegistryID:         dependency.RegistrySourceID,
			RegistrySourceName: dependency.RegistrySourceName,
		})
	}

	return result
}

func toUIProjects(projects []storage.Project) []ui.Project {
	result := make([]ui.Project, 0, len(projects))
	for _, project := range projects {
		result = append(result, ui.Project{
			ID:              project.ID,
			ProjectID:       project.ProjectID,
			NamespaceID:     project.NamespaceID,
			NamespaceName:   project.NamespaceName,
			SourceID:        project.SourceID,
			SourceName:      project.SourceName,
			StackID:         project.StackID,
			StackName:       project.StackName,
			StackIcon:       project.StackIcon,
			StackColor:      project.StackColor,
			Icon:            project.Icon,
			Name:            project.Name,
			Color:           project.Color,
			Freezing:        project.Freezing,
			EndOfLife:       project.EndOfLife,
			DependencyCount: project.DependencyCount,
		})
	}

	return result
}

func toUIProjectDependencies(dependencies []storage.ProjectDependency) []ui.ProjectDependency {
	result := make([]ui.ProjectDependency, 0, len(dependencies))
	for _, dependency := range dependencies {
		result = append(result, ui.ProjectDependency{
			ID:             dependency.ID,
			Name:           dependency.Name,
			Version:        dependency.Version,
			DependencyType: dependency.DependencyType,
			SourceFile:     dependency.SourceFile,
		})
	}

	return result
}

func toUIProjectDependencyRun(run storage.ProjectDependencyRun) ui.ProjectDependencyRun {
	if run.ID == 0 {
		return ui.ProjectDependencyRun{}
	}
	updatedAt := run.UpdatedAt
	if run.FinishedAt != nil {
		updatedAt = *run.FinishedAt
	}

	return ui.ProjectDependencyRun{
		CommitShortSHA: run.CommitShortSHA,
		Status:         run.Status,
		UpdatedAt:      updatedAt.Format("2006-01-02 15:04"),
		Error:          run.Error,
		HasValue:       true,
	}
}

func toUISources(sources []storage.Source) []ui.Source {
	result := make([]ui.Source, 0, len(sources))
	for _, source := range sources {
		result = append(result, ui.Source{
			ID:           source.ID,
			Icon:         source.Icon,
			Name:         source.Name,
			Color:        source.Color,
			PATToken:     source.PATToken,
			URL:          source.URL,
			Type:         source.Type,
			RegistryKind: source.RegistryKind,
		})
	}

	return result
}

func toUIPolicies(policies []storage.Policy) []ui.Policy {
	result := make([]ui.Policy, 0, len(policies))
	for _, policy := range policies {
		result = append(result, ui.Policy{
			ID:              policy.ID,
			Name:            policy.Name,
			DependencyCount: policy.DependencyCount,
		})
	}

	return result
}

func toUIPolicyValues(values []storage.PolicyValue) []ui.PolicyValue {
	result := make([]ui.PolicyValue, 0, len(values))
	for _, value := range values {
		result = append(result, ui.PolicyValue{
			ID:                 value.ID,
			PolicyID:           value.PolicyID,
			DependencyID:       value.DependencyID,
			DependencyIcon:     value.DependencyIcon,
			DependencyName:     value.DependencyName,
			DependencyColor:    value.DependencyColor,
			StackID:            value.StackID,
			StackName:          value.StackName,
			RegistryName:       value.RegistryName,
			RegistryID:         value.RegistrySourceID,
			RegistrySourceName: value.RegistrySourceName,
			Version:            value.Version,
		})
	}

	return result
}

func toUIDependencyView(view storage.DependencyView) ui.DependencyView {
	result := ui.DependencyView{
		StackID:   view.StackID,
		StackName: view.StackName,
		Columns:   make([]ui.DependencyViewColumn, 0, len(view.Columns)),
		Rows:      make([]ui.DependencyViewRow, 0, len(view.Rows)),
	}
	for _, column := range view.Columns {
		result.Columns = append(result.Columns, ui.DependencyViewColumn{
			DependencyID:  column.DependencyID,
			Icon:          column.Icon,
			Name:          column.Name,
			Color:         column.Color,
			PolicyVersion: column.PolicyVersion,
		})
	}
	for _, row := range view.Rows {
		result.Rows = append(result.Rows, ui.DependencyViewRow{
			ProjectID:        row.ProjectID,
			ProjectIcon:      row.ProjectIcon,
			ProjectName:      row.ProjectName,
			ProjectColor:     row.ProjectColor,
			ProjectFreezing:  row.ProjectFreezing,
			ProjectEndOfLife: row.ProjectEndOfLife,
			Versions:         row.Versions,
		})
	}

	return result
}

func newStackForm() ui.StackForm {
	return ui.StackForm{
		Open:  true,
		Mode:  ui.StackFormModeCreate,
		Focus: ui.StackFormFieldIcon,
	}
}

func newNamespaceForm(policies []ui.Policy) ui.StackForm {
	form := newStackForm()
	if len(policies) > 0 {
		form.PolicyID = policies[0].ID
	}
	return normalizeNamespaceForm(form, policies)
}

func newDependencyForm(stacks []ui.Stack, sources []ui.Source) ui.DependencyForm {
	form := ui.DependencyForm{
		Open:  true,
		Mode:  ui.StackFormModeCreate,
		Focus: ui.DependencyFormFieldIcon,
	}
	if len(stacks) > 0 {
		form.StackID = stacks[0].ID
	}
	registrySources := registryUISources(sources)
	if len(registrySources) > 0 {
		form.RegistryID = registrySources[0].ID
	}

	return normalizeDependencyForm(form, stacks, registrySources)
}

func newProjectForm(namespaces []ui.Namespace, sources []ui.Source, stacks []ui.Stack) ui.ProjectForm {
	sources = vcsSources(sources)
	form := ui.ProjectForm{
		Open:  true,
		Mode:  ui.StackFormModeCreate,
		Focus: ui.ProjectFormFieldIcon,
	}
	if len(namespaces) > 0 {
		form.NamespaceID = namespaces[0].ID
	}
	if len(sources) > 0 {
		form.SourceID = sources[0].ID
	}
	if len(stacks) > 0 {
		form.StackID = stacks[0].ID
	}

	return normalizeProjectForm(form, namespaces, sources, stacks)
}

func newSourceForm() ui.SourceForm {
	return normalizeSourceForm(ui.SourceForm{
		Open:  true,
		Mode:  ui.StackFormModeCreate,
		Focus: ui.SourceFormFieldName,
		Type:  storage.SourceTypeGitLab,
	})
}

func newPolicyUpdateForm(policyID int64, values []ui.PolicyValue, stacks []ui.Stack, sources []ui.Source) ui.PolicyUpdateForm {
	form := ui.PolicyUpdateForm{
		Open:     true,
		PolicyID: policyID,
		Focus:    ui.PolicyUpdateFormFieldStack,
	}
	policyStacks := policyStacks(values, stacks)
	registrySources := registryUISources(sources)
	if len(policyStacks) > 0 {
		form.StackID = policyStacks[0].ID
	}
	if len(registrySources) > 0 {
		form.RegistryID = registrySources[0].ID
	}
	return normalizePolicyUpdateForm(form, policyStacks, registrySources)
}

func newPolicyForm(namespaces []ui.Namespace, policies []ui.Policy) ui.PolicyForm {
	return normalizePolicyForm(ui.PolicyForm{
		Open:  true,
		Mode:  ui.StackFormModeCreate,
		Focus: ui.PolicyFormFieldName,
	}, namespaces)
}

func newPolicyValueForm(policyID int64, dependencies []ui.Dependency, sources []ui.Source) ui.PolicyValueForm {
	form := ui.PolicyValueForm{
		Open:     true,
		Mode:     ui.StackFormModeCreate,
		PolicyID: policyID,
		Focus:    ui.PolicyValueFormFieldDependency,
	}
	if len(dependencies) > 0 {
		form.DependencyID = dependencies[0].ID
	}
	return normalizePolicyValueForm(form, dependencies, registryUISources(sources))
}

func (m *model) openEditStackForm() {
	stack, ok := m.selectedStack()
	if !ok {
		return
	}

	m.form = editStackForm(stack.ID, stack.Icon, stack.Color, stack.Name)
}

func (m *model) openDeleteConfirm() {
	stack, ok := m.selectedStack()
	if !ok {
		return
	}

	m.deleteConfirm = deleteConfirm(stack.ID, stack.Name)
}

func (m *model) openEditNamespaceForm() {
	namespace, ok := m.selectedNamespace()
	if !ok {
		return
	}

	m.namespaceForm = editStackForm(namespace.ID, namespace.Icon, namespace.Color, namespace.Name)
	m.namespaceForm.PolicyID = namespace.PolicyID
	m.namespaceForm = normalizeNamespaceForm(m.namespaceForm, m.policies)
}

func (m *model) openNamespaceDeleteConfirm() {
	namespace, ok := m.selectedNamespace()
	if !ok {
		return
	}

	m.namespaceDeleteConfirm = deleteConfirm(namespace.ID, namespace.Name)
}

func (m *model) openEditDependencyForm() {
	dependency, ok := m.selectedDependency()
	if !ok {
		return
	}

	m.dependencyForm = normalizeDependencyForm(ui.DependencyForm{
		Open:         true,
		Mode:         ui.StackFormModeEdit,
		DependencyID: dependency.ID,
		StackID:      dependency.StackID,
		Focus:        ui.DependencyFormFieldIcon,
		Icon:         dependency.Icon,
		Color:        dependency.Color,
		Name:         dependency.Name,
		RegistryName: dependency.RegistryName,
		RegistryID:   dependency.RegistryID,
	}, m.stacks, registryUISources(m.sources))
}

func (m *model) openEditProjectForm() {
	project, ok := m.selectedProject()
	if !ok {
		return
	}

	m.projectForm = normalizeProjectForm(ui.ProjectForm{
		Open:        true,
		Mode:        ui.StackFormModeEdit,
		ID:          project.ID,
		ProjectID:   project.ProjectID,
		NamespaceID: project.NamespaceID,
		SourceID:    project.SourceID,
		StackID:     project.StackID,
		Focus:       ui.ProjectFormFieldIcon,
		Icon:        project.Icon,
		Color:       project.Color,
		Name:        project.Name,
		Freezing:    project.Freezing,
		EndOfLife:   project.EndOfLife,
	}, m.namespaces, vcsSources(m.sources), m.stacks)
}

func (m *model) openCloneProjectForm() {
	project, ok := m.selectedProject()
	if !ok {
		return
	}

	m.projectForm = normalizeProjectForm(ui.ProjectForm{
		Open:        true,
		Mode:        ui.StackFormModeCreate,
		ProjectID:   project.ProjectID,
		NamespaceID: project.NamespaceID,
		SourceID:    project.SourceID,
		StackID:     project.StackID,
		Focus:       ui.ProjectFormFieldIcon,
		Icon:        project.Icon,
		Color:       project.Color,
		Name:        project.Name,
	}, m.namespaces, vcsSources(m.sources), m.stacks)
}

func (m *model) openDependencyDeleteConfirm() {
	dependency, ok := m.selectedDependency()
	if !ok {
		return
	}

	m.dependencyDeleteConfirm = deleteConfirm(dependency.ID, dependency.Name)
}

func (m *model) openProjectDeleteConfirm() {
	project, ok := m.selectedProject()
	if !ok {
		return
	}

	m.projectDeleteConfirm = deleteConfirm(project.ID, project.Name)
}

func (m *model) openEditSourceForm() {
	source, ok := m.selectedSource()
	if !ok {
		return
	}

	m.sourceForm = normalizeSourceForm(ui.SourceForm{
		Open:         true,
		Mode:         ui.StackFormModeEdit,
		SourceID:     source.ID,
		Focus:        ui.SourceFormFieldName,
		Name:         source.Name,
		URL:          source.URL,
		PATToken:     source.PATToken,
		Type:         source.Type,
		RegistryKind: source.RegistryKind,
	})
}

func (m *model) openPolicyUpdateForm() {
	if m.screen != ui.ScreenPolicies || m.selectedPolicyID == 0 {
		return
	}
	m.policyUpdateForm = newPolicyUpdateForm(m.selectedPolicyID, m.policyValues, m.stacks, m.sources)
}

func (m *model) openSourceDeleteConfirm() {
	source, ok := m.selectedSource()
	if !ok {
		return
	}

	m.sourceDeleteConfirm = deleteConfirm(source.ID, source.Name)
}

func (m *model) openEditPolicyForm() {
	policy, ok := m.selectedPolicy()
	if !ok {
		return
	}

	m.policyForm = normalizePolicyForm(ui.PolicyForm{
		Open:     true,
		Mode:     ui.StackFormModeEdit,
		PolicyID: policy.ID,
		Focus:    ui.PolicyFormFieldName,
		Name:     policy.Name,
	}, m.namespaces)
}

func (m *model) openPolicyDeleteConfirm() {
	policy, ok := m.selectedPolicy()
	if !ok {
		return
	}

	m.policyDeleteConfirm = deleteConfirm(policy.ID, policy.Name)
}

func (m *model) openEditPolicyValueForm() {
	value, ok := m.selectedPolicyValue()
	if !ok {
		return
	}

	m.policyValueForm = normalizePolicyValueForm(ui.PolicyValueForm{
		Open:          true,
		Mode:          ui.StackFormModeEdit,
		PolicyValueID: value.ID,
		PolicyID:      value.PolicyID,
		DependencyID:  value.DependencyID,
		Focus:         ui.PolicyValueFormFieldDependency,
		Version:       value.Version,
	}, m.dependencies, registryUISources(m.sources))
}

func (m *model) openPolicyValueDeleteConfirm() {
	value, ok := m.selectedPolicyValue()
	if !ok {
		return
	}

	m.policyDeleteConfirm = deleteConfirm(value.ID, value.DependencyName)
}

func editStackForm(id int64, icon string, color string, name string) ui.StackForm {
	return normalizeStackForm(ui.StackForm{
		Open:    true,
		Mode:    ui.StackFormModeEdit,
		StackID: id,
		Focus:   ui.StackFormFieldIcon,
		Icon:    icon,
		Color:   color,
		Name:    name,
	})
}

func deleteConfirm(id int64, name string) ui.DeleteConfirm {
	return ui.DeleteConfirm{
		Open:    true,
		StackID: id,
		Name:    name,
	}
}

func stackID(s ui.Stack) int64           { return s.ID }
func namespaceID(n ui.Namespace) int64   { return n.ID }
func dependencyID(d ui.Dependency) int64 { return d.ID }
func projectID(p ui.Project) int64       { return p.ID }
func projectDependencyID(d ui.ProjectDependency) int64 {
	return d.ID
}
func sourceID(s ui.Source) int64 { return s.ID }
func policyID(p ui.Policy) int64 { return p.ID }
func policyValueID(v ui.PolicyValue) int64 {
	return v.ID
}
func dependencyViewRowID(r ui.DependencyViewRow) int64 {
	return r.ProjectID
}
func dashboardAttentionRowID(r ui.DashboardAttentionRow) int64 {
	return r.ProjectID
}

func (m *model) ensureSelectedStack() {
	m.selectedStackID = ensureSelected(m.stacks, m.selectedStackID, stackID)
	m.ensureViewStack()
}

func (m *model) ensureSelectedAttentionProject() {
	m.dashboard.selectedProjectID = ensureSelected(m.dashboard.attentionRows, m.dashboard.selectedProjectID, dashboardAttentionRowID)
}

func (m *model) ensureSelectedNamespace() {
	m.selectedNamespaceID = ensureSelected(m.namespaces, m.selectedNamespaceID, namespaceID)
}

func (m *model) ensureSelectedDependency() {
	m.selectedDependencyID = ensureSelected(m.dependencies, m.selectedDependencyID, dependencyID)
}

func (m *model) ensureSelectedProject() {
	m.selectedProjectID = ensureSelected(m.projects, m.selectedProjectID, projectID)
}

func (m *model) ensureSelectedProjectDependency() {
	m.selectedProjectDepID = ensureSelected(m.projectDependencies, m.selectedProjectDepID, projectDependencyID)
}

func (m *model) ensureViewStack() {
	m.viewStackID = ensureSelected(m.stacks, m.viewStackID, stackID)
}

func (m *model) ensureSelectedViewProject() {
	m.selectedViewProjectID = ensureSelected(m.dependencyView.Rows, m.selectedViewProjectID, dependencyViewRowID)
}

func (m *model) ensureSelectedSource() {
	m.selectedSourceID = ensureSelected(m.sources, m.selectedSourceID, sourceID)
}

func (m *model) ensureSelectedPolicy() {
	oldID := m.selectedPolicyID
	m.selectedPolicyID = ensureSelected(m.policies, m.selectedPolicyID, policyID)
	if oldID != m.selectedPolicyID {
		m.policyValues = []ui.PolicyValue{}
		m.selectedPolicyValueID = 0
	}
}

func (m *model) ensureSelectedPolicyValue() {
	m.selectedPolicyValueID = ensureSelected(m.policyValues, m.selectedPolicyValueID, policyValueID)
}

func (m *model) selectPreviousAttentionProject() {
	m.dashboard.selectedProjectID = selectPrevious(m.dashboard.attentionRows, m.dashboard.selectedProjectID, dashboardAttentionRowID)
}

func (m *model) selectNextAttentionProject() {
	m.dashboard.selectedProjectID = selectNext(m.dashboard.attentionRows, m.dashboard.selectedProjectID, dashboardAttentionRowID)
}

func (m *model) selectPreviousStack() {
	m.selectedStackID = selectPrevious(m.stacks, m.selectedStackID, stackID)
}

func (m *model) selectNextStack() {
	m.selectedStackID = selectNext(m.stacks, m.selectedStackID, stackID)
}

func (m *model) selectPreviousNamespace() {
	m.selectedNamespaceID = selectPrevious(m.namespaces, m.selectedNamespaceID, namespaceID)
}

func (m *model) selectNextNamespace() {
	m.selectedNamespaceID = selectNext(m.namespaces, m.selectedNamespaceID, namespaceID)
}

func (m *model) selectPreviousDependency() {
	m.selectedDependencyID = selectPrevious(m.dependencies, m.selectedDependencyID, dependencyID)
}

func (m *model) selectPreviousProject() {
	m.selectedProjectID = selectPrevious(m.projects, m.selectedProjectID, projectID)
}

func (m *model) selectPreviousProjectDependency() {
	m.selectedProjectDepID = selectPreviousBounded(m.projectDependencies, m.selectedProjectDepID, projectDependencyID)
}

func (m *model) selectNextDependency() {
	m.selectedDependencyID = selectNext(m.dependencies, m.selectedDependencyID, dependencyID)
}

func (m *model) selectNextProject() {
	m.selectedProjectID = selectNext(m.projects, m.selectedProjectID, projectID)
}

func (m *model) selectNextProjectDependency() {
	m.selectedProjectDepID = selectNextBounded(m.projectDependencies, m.selectedProjectDepID, projectDependencyID)
}

func (m *model) selectNextViewStack() {
	m.viewStackID = selectNext(m.stacks, m.viewStackID, stackID)
	m.viewColumnOffset = 0
}

func (m *model) selectPreviousViewStack() {
	m.viewStackID = selectPrevious(m.stacks, m.viewStackID, stackID)
	m.viewColumnOffset = 0
}

func (m *model) selectPreviousViewProject() {
	m.selectedViewProjectID = selectPreviousBounded(m.dependencyView.Rows, m.selectedViewProjectID, dependencyViewRowID)
}

func (m *model) selectNextViewProject() {
	m.selectedViewProjectID = selectNextBounded(m.dependencyView.Rows, m.selectedViewProjectID, dependencyViewRowID)
}

func (m *model) scrollViewColumnsLeft() {
	if m.viewColumnOffset > 0 {
		m.viewColumnOffset--
	}
}

func (m *model) scrollViewColumnsRight() {
	if len(m.dependencyView.Columns) == 0 {
		m.viewColumnOffset = 0
		return
	}
	if m.viewColumnOffset < len(m.dependencyView.Columns)-1 {
		m.viewColumnOffset++
	}
}

func (m *model) selectPreviousSource() {
	m.selectedSourceID = selectPrevious(m.sources, m.selectedSourceID, sourceID)
}

func (m *model) selectNextSource() {
	m.selectedSourceID = selectNext(m.sources, m.selectedSourceID, sourceID)
}

func (m *model) selectPreviousPolicy() {
	oldID := m.selectedPolicyID
	m.selectedPolicyID = selectPrevious(m.policies, m.selectedPolicyID, policyID)
	if oldID != m.selectedPolicyID {
		m.policyValues = []ui.PolicyValue{}
		m.selectedPolicyValueID = 0
	}
}

func (m *model) selectNextPolicy() {
	oldID := m.selectedPolicyID
	m.selectedPolicyID = selectNext(m.policies, m.selectedPolicyID, policyID)
	if oldID != m.selectedPolicyID {
		m.policyValues = []ui.PolicyValue{}
		m.selectedPolicyValueID = 0
	}
}

func (m *model) selectPreviousPolicyValue() {
	m.selectedPolicyValueID = selectPrevious(m.policyValues, m.selectedPolicyValueID, policyValueID)
}

func (m *model) selectNextPolicyValue() {
	m.selectedPolicyValueID = selectNext(m.policyValues, m.selectedPolicyValueID, policyValueID)
}

func (m model) selectedStack() (ui.Stack, bool) {
	return findByID(m.stacks, m.selectedStackID, stackID)
}

func (m model) selectedNamespace() (ui.Namespace, bool) {
	return findByID(m.namespaces, m.selectedNamespaceID, namespaceID)
}

func (m model) selectedDependency() (ui.Dependency, bool) {
	return findByID(m.dependencies, m.selectedDependencyID, dependencyID)
}

func (m model) selectedProject() (ui.Project, bool) {
	return findByID(m.projects, m.selectedProjectID, projectID)
}

func (m model) selectedSource() (ui.Source, bool) {
	return findByID(m.sources, m.selectedSourceID, sourceID)
}

func (m model) selectedPolicy() (ui.Policy, bool) {
	return findByID(m.policies, m.selectedPolicyID, policyID)
}

func (m model) selectedPolicyValue() (ui.PolicyValue, bool) {
	return findByID(m.policyValues, m.selectedPolicyValueID, policyValueID)
}

func (m model) policyNameForNamespace(id int64) string {
	namespace, ok := findByID(m.namespaces, id, namespaceID)
	if !ok {
		return ""
	}
	for _, policy := range m.policies {
		if policy.ID == namespace.PolicyID {
			return policy.Name
		}
	}

	return ""
}

func nextStackFormField(field ui.StackFormField) ui.StackFormField {
	if field == ui.StackFormFieldName {
		return ui.StackFormFieldIcon
	}

	return field + 1
}

func previousStackFormField(field ui.StackFormField) ui.StackFormField {
	if field == ui.StackFormFieldIcon {
		return ui.StackFormFieldName
	}

	return field - 1
}

func nextNamespaceFormField(field ui.StackFormField) ui.StackFormField {
	if field == ui.StackFormFieldPolicy {
		return ui.StackFormFieldIcon
	}

	return field + 1
}

func previousNamespaceFormField(field ui.StackFormField) ui.StackFormField {
	if field == ui.StackFormFieldIcon {
		return ui.StackFormFieldPolicy
	}

	return field - 1
}

func appendStackFormRunes(form ui.StackForm, runes []rune) ui.StackForm {
	value := string(runes)
	switch form.Focus {
	case ui.StackFormFieldIcon:
		form.Icon += value
	case ui.StackFormFieldColor:
		form.Color += value
	case ui.StackFormFieldName:
		form.Name += value
	}

	return normalizeStackForm(form)
}

func deleteStackFormRune(form ui.StackForm) ui.StackForm {
	switch form.Focus {
	case ui.StackFormFieldIcon:
		form.Icon = trimLastRune(form.Icon)
	case ui.StackFormFieldColor:
		form.Color = trimLastRune(form.Color)
	case ui.StackFormFieldName:
		form.Name = trimLastRune(form.Name)
	}

	return normalizeStackForm(form)
}

func appendNamespaceFormRunes(form ui.StackForm, runes []rune) ui.StackForm {
	value := string(runes)
	switch form.Focus {
	case ui.StackFormFieldIcon:
		form.Icon += value
	case ui.StackFormFieldColor:
		form.Color += value
	case ui.StackFormFieldName:
		form.Name += value
	case ui.StackFormFieldPolicy:
	}

	return form
}

func deleteNamespaceFormRune(form ui.StackForm) ui.StackForm {
	switch form.Focus {
	case ui.StackFormFieldIcon:
		form.Icon = trimLastRune(form.Icon)
	case ui.StackFormFieldColor:
		form.Color = trimLastRune(form.Color)
	case ui.StackFormFieldName:
		form.Name = trimLastRune(form.Name)
	case ui.StackFormFieldPolicy:
	}

	return form
}

func normalizeStackForm(form ui.StackForm) ui.StackForm {
	form.CanSave = strings.TrimSpace(form.Icon) != "" &&
		strings.TrimSpace(form.Color) != "" &&
		strings.TrimSpace(form.Name) != ""
	if form.CanSave {
		form.Error = ""
	}

	return form
}

func normalizeNamespaceForm(form ui.StackForm, policies []ui.Policy) ui.StackForm {
	if !hasPolicyID(policies, form.PolicyID) && len(policies) > 0 {
		form.PolicyID = policies[0].ID
	}
	form.Policy = policyNameByID(policies, form.PolicyID)
	form.CanSave = strings.TrimSpace(form.Icon) != "" &&
		strings.TrimSpace(form.Color) != "" &&
		strings.TrimSpace(form.Name) != "" &&
		hasPolicyID(policies, form.PolicyID)
	if form.CanSave {
		form.Error = ""
	}

	return form
}

func nextDependencyFormField(field ui.DependencyFormField) ui.DependencyFormField {
	if field == ui.DependencyFormFieldStack {
		return ui.DependencyFormFieldIcon
	}

	return field + 1
}

func previousDependencyFormField(field ui.DependencyFormField) ui.DependencyFormField {
	if field == ui.DependencyFormFieldIcon {
		return ui.DependencyFormFieldStack
	}

	return field - 1
}

func nextPolicyUpdateFormField(field ui.PolicyUpdateFormField) ui.PolicyUpdateFormField {
	if field == ui.PolicyUpdateFormFieldRegistry {
		return ui.PolicyUpdateFormFieldStack
	}
	return field + 1
}

func previousPolicyUpdateFormField(field ui.PolicyUpdateFormField) ui.PolicyUpdateFormField {
	if field == ui.PolicyUpdateFormFieldStack {
		return ui.PolicyUpdateFormFieldRegistry
	}
	return field - 1
}

func nextProjectFormField(field ui.ProjectFormField) ui.ProjectFormField {
	switch field {
	case ui.ProjectFormFieldIcon:
		return ui.ProjectFormFieldColor
	case ui.ProjectFormFieldColor:
		return ui.ProjectFormFieldName
	case ui.ProjectFormFieldName:
		return ui.ProjectFormFieldNamespace
	case ui.ProjectFormFieldNamespace:
		return ui.ProjectFormFieldSource
	case ui.ProjectFormFieldSource:
		return ui.ProjectFormFieldStack
	case ui.ProjectFormFieldStack:
		return ui.ProjectFormFieldProjectID
	case ui.ProjectFormFieldProjectID:
		return ui.ProjectFormFieldFreezing
	case ui.ProjectFormFieldFreezing:
		return ui.ProjectFormFieldEndOfLife
	default:
		return ui.ProjectFormFieldIcon
	}
}

func previousProjectFormField(field ui.ProjectFormField) ui.ProjectFormField {
	switch field {
	case ui.ProjectFormFieldIcon:
		return ui.ProjectFormFieldEndOfLife
	case ui.ProjectFormFieldColor:
		return ui.ProjectFormFieldIcon
	case ui.ProjectFormFieldName:
		return ui.ProjectFormFieldColor
	case ui.ProjectFormFieldNamespace:
		return ui.ProjectFormFieldName
	case ui.ProjectFormFieldSource:
		return ui.ProjectFormFieldNamespace
	case ui.ProjectFormFieldStack:
		return ui.ProjectFormFieldSource
	case ui.ProjectFormFieldProjectID:
		return ui.ProjectFormFieldStack
	case ui.ProjectFormFieldFreezing:
		return ui.ProjectFormFieldProjectID
	default:
		return ui.ProjectFormFieldFreezing
	}
}

func isProjectPickerFocused(field ui.ProjectFormField) bool {
	return field == ui.ProjectFormFieldNamespace ||
		field == ui.ProjectFormFieldSource ||
		field == ui.ProjectFormFieldStack
}

func nextSourceFormField(field ui.SourceFormField) ui.SourceFormField {
	if field == ui.SourceFormFieldType {
		return ui.SourceFormFieldName
	}

	return field + 1
}

func nextSourceFormFieldForForm(form ui.SourceForm) ui.SourceFormField {
	if form.Type == storage.SourceTypeRegistry && form.Focus == ui.SourceFormFieldType {
		return ui.SourceFormFieldRegistryKind
	}
	if form.Focus == ui.SourceFormFieldRegistryKind {
		return ui.SourceFormFieldName
	}
	return nextSourceFormField(form.Focus)
}

func previousSourceFormField(field ui.SourceFormField) ui.SourceFormField {
	if field == ui.SourceFormFieldName {
		return ui.SourceFormFieldType
	}

	return field - 1
}

func previousSourceFormFieldForForm(form ui.SourceForm) ui.SourceFormField {
	if form.Type == storage.SourceTypeRegistry && form.Focus == ui.SourceFormFieldName {
		return ui.SourceFormFieldRegistryKind
	}
	if form.Focus == ui.SourceFormFieldRegistryKind {
		return ui.SourceFormFieldType
	}
	return previousSourceFormField(form.Focus)
}

func nextPolicyFormField(field ui.PolicyFormField) ui.PolicyFormField {
	return ui.PolicyFormFieldName
}

func previousPolicyFormField(field ui.PolicyFormField) ui.PolicyFormField {
	return ui.PolicyFormFieldName
}

func nextPolicyValueFormField(field ui.PolicyValueFormField) ui.PolicyValueFormField {
	if field == ui.PolicyValueFormFieldVersion {
		return ui.PolicyValueFormFieldDependency
	}

	return field + 1
}

func previousPolicyValueFormField(field ui.PolicyValueFormField) ui.PolicyValueFormField {
	if field == ui.PolicyValueFormFieldDependency {
		return ui.PolicyValueFormFieldVersion
	}

	return field - 1
}

func appendDependencyFormRunes(form ui.DependencyForm, runes []rune) ui.DependencyForm {
	value := string(runes)
	switch form.Focus {
	case ui.DependencyFormFieldIcon:
		form.Icon += value
	case ui.DependencyFormFieldColor:
		form.Color += value
	case ui.DependencyFormFieldName:
		form.Name += value
	case ui.DependencyFormFieldPackage:
		form.RegistryName += value
	case ui.DependencyFormFieldRegistry, ui.DependencyFormFieldStack:
	}

	return form
}

func appendProjectFormRunes(form ui.ProjectForm, runes []rune) ui.ProjectForm {
	value := string(runes)
	switch form.Focus {
	case ui.ProjectFormFieldIcon:
		form.Icon += value
	case ui.ProjectFormFieldColor:
		form.Color += value
	case ui.ProjectFormFieldName:
		form.Name += value
	case ui.ProjectFormFieldProjectID:
		form.ProjectID += value
	case ui.ProjectFormFieldNamespace, ui.ProjectFormFieldSource, ui.ProjectFormFieldStack, ui.ProjectFormFieldFreezing, ui.ProjectFormFieldEndOfLife:
	}

	return form
}

func deleteDependencyFormRune(form ui.DependencyForm) ui.DependencyForm {
	switch form.Focus {
	case ui.DependencyFormFieldIcon:
		form.Icon = trimLastRune(form.Icon)
	case ui.DependencyFormFieldColor:
		form.Color = trimLastRune(form.Color)
	case ui.DependencyFormFieldName:
		form.Name = trimLastRune(form.Name)
	case ui.DependencyFormFieldPackage:
		form.RegistryName = trimLastRune(form.RegistryName)
	case ui.DependencyFormFieldRegistry, ui.DependencyFormFieldStack:
	}

	return form
}

func deleteProjectFormRune(form ui.ProjectForm) ui.ProjectForm {
	switch form.Focus {
	case ui.ProjectFormFieldIcon:
		form.Icon = trimLastRune(form.Icon)
	case ui.ProjectFormFieldColor:
		form.Color = trimLastRune(form.Color)
	case ui.ProjectFormFieldName:
		form.Name = trimLastRune(form.Name)
	case ui.ProjectFormFieldProjectID:
		form.ProjectID = trimLastRune(form.ProjectID)
	case ui.ProjectFormFieldNamespace, ui.ProjectFormFieldSource, ui.ProjectFormFieldStack, ui.ProjectFormFieldFreezing, ui.ProjectFormFieldEndOfLife:
	}

	return form
}

func appendSourceFormRunes(form ui.SourceForm, runes []rune) ui.SourceForm {
	value := string(runes)
	switch form.Focus {
	case ui.SourceFormFieldName:
		form.Name += value
	case ui.SourceFormFieldURL:
		form.URL += value
	case ui.SourceFormFieldPATToken:
		form.PATToken += value
	case ui.SourceFormFieldType:
	case ui.SourceFormFieldRegistryKind:
	}

	return form
}

func deleteSourceFormRune(form ui.SourceForm) ui.SourceForm {
	switch form.Focus {
	case ui.SourceFormFieldName:
		form.Name = trimLastRune(form.Name)
	case ui.SourceFormFieldURL:
		form.URL = trimLastRune(form.URL)
	case ui.SourceFormFieldPATToken:
		form.PATToken = trimLastRune(form.PATToken)
	case ui.SourceFormFieldType:
	case ui.SourceFormFieldRegistryKind:
	}

	return form
}

func appendPolicyFormRunes(form ui.PolicyForm, runes []rune) ui.PolicyForm {
	if form.Focus == ui.PolicyFormFieldName {
		form.Name += string(runes)
	}

	return form
}

func deletePolicyFormRune(form ui.PolicyForm) ui.PolicyForm {
	if form.Focus == ui.PolicyFormFieldName {
		form.Name = trimLastRune(form.Name)
	}

	return form
}

func appendPolicyValueFormRunes(form ui.PolicyValueForm, runes []rune) ui.PolicyValueForm {
	if form.Focus == ui.PolicyValueFormFieldVersion {
		form.Version += string(runes)
	}

	return form
}

func deletePolicyValueFormRune(form ui.PolicyValueForm) ui.PolicyValueForm {
	if form.Focus == ui.PolicyValueFormFieldVersion {
		form.Version = trimLastRune(form.Version)
	}

	return form
}

func normalizeDependencyForm(form ui.DependencyForm, stacks []ui.Stack, sources []ui.Source) ui.DependencyForm {
	if !hasStackID(stacks, form.StackID) && len(stacks) > 0 {
		form.StackID = stacks[0].ID
	}
	if !hasSourceID(sources, form.RegistryID) && len(sources) > 0 {
		form.RegistryID = sources[0].ID
	}
	if strings.TrimSpace(form.RegistryName) == "" && form.Focus != ui.DependencyFormFieldPackage {
		form.RegistryName = form.Name
	}
	form.CanSave = strings.TrimSpace(form.Icon) != "" &&
		strings.TrimSpace(form.Color) != "" &&
		strings.TrimSpace(form.Name) != "" &&
		strings.TrimSpace(form.RegistryName) != "" &&
		hasSourceID(sources, form.RegistryID) &&
		hasStackID(stacks, form.StackID)
	if form.CanSave {
		form.Error = ""
	}

	return form
}

func normalizeProjectForm(form ui.ProjectForm, namespaces []ui.Namespace, sources []ui.Source, stacks []ui.Stack) ui.ProjectForm {
	if !hasNamespaceID(namespaces, form.NamespaceID) && len(namespaces) > 0 {
		form.NamespaceID = namespaces[0].ID
	}
	if !hasSourceID(sources, form.SourceID) && len(sources) > 0 {
		form.SourceID = sources[0].ID
	}
	if !hasStackID(stacks, form.StackID) && len(stacks) > 0 {
		form.StackID = stacks[0].ID
	}
	form.CanSave = strings.TrimSpace(form.Icon) != "" &&
		strings.TrimSpace(form.Color) != "" &&
		strings.TrimSpace(form.Name) != "" &&
		strings.TrimSpace(form.ProjectID) != "" &&
		hasNamespaceID(namespaces, form.NamespaceID) &&
		hasSourceID(sources, form.SourceID) &&
		hasStackID(stacks, form.StackID)
	if form.CanSave {
		form.Error = ""
	}

	return form
}

func normalizeSourceForm(form ui.SourceForm) ui.SourceForm {
	if !hasSourceType(form.Type) {
		form.Type = storage.SourceTypeGitLab
	}
	if form.Type != storage.SourceTypeRegistry {
		form.RegistryKind = ""
	}
	if form.Type == storage.SourceTypeRegistry && !hasRegistryKind(form.RegistryKind) {
		form.RegistryKind = registry.KindNPM
	}
	form.CanSave = strings.TrimSpace(form.Name) != "" &&
		strings.TrimSpace(form.URL) != "" &&
		hasSourceType(form.Type) &&
		(form.Type == storage.SourceTypeRegistry || strings.TrimSpace(form.PATToken) != "") &&
		(form.Type != storage.SourceTypeRegistry || hasRegistryKind(form.RegistryKind))
	if form.CanSave {
		form.Error = ""
	}

	return form
}

func normalizePolicyForm(form ui.PolicyForm, namespaces []ui.Namespace) ui.PolicyForm {
	form.CanSave = strings.TrimSpace(form.Name) != ""
	if form.CanSave {
		form.Error = ""
	}

	return form
}

func normalizePolicyValueForm(form ui.PolicyValueForm, dependencies []ui.Dependency, sources []ui.Source) ui.PolicyValueForm {
	if !hasDependencyID(dependencies, form.DependencyID) && len(dependencies) > 0 {
		form.DependencyID = dependencies[0].ID
	}
	if dependency, ok := findByID(dependencies, form.DependencyID, dependencyID); ok && dependency.RegistryID != 0 {
		form.RegistryID = dependency.RegistryID
	}
	if !hasSourceID(sources, form.RegistryID) && len(sources) > 0 {
		form.RegistryID = sources[0].ID
	}
	form.CanSave = form.PolicyID != 0 &&
		strings.TrimSpace(form.Version) != "" &&
		hasDependencyID(dependencies, form.DependencyID)
	form.CanUpdate = hasDependencyID(dependencies, form.DependencyID) &&
		hasSourceID(sources, form.RegistryID)
	if form.CanSave {
		form.Error = ""
	}

	return form
}

func normalizePolicyUpdateForm(form ui.PolicyUpdateForm, stacks []ui.Stack, sources []ui.Source) ui.PolicyUpdateForm {
	if !hasStackID(stacks, form.StackID) && len(stacks) > 0 {
		form.StackID = stacks[0].ID
	}
	if !hasSourceID(sources, form.RegistryID) && len(sources) > 0 {
		form.RegistryID = sources[0].ID
	}
	form.CanStart = form.PolicyID != 0 &&
		hasStackID(stacks, form.StackID) &&
		hasSourceID(sources, form.RegistryID)
	if form.CanStart {
		form.Error = ""
	}
	if len(stacks) == 0 {
		form.Error = "No pinned deps for policy stack"
	}
	if len(sources) == 0 {
		form.Error = "Add registry source first"
	}
	return form
}

func hasStackID(stacks []ui.Stack, id int64) bool {
	_, ok := findByID(stacks, id, stackID)
	return ok
}

func hasNamespaceID(namespaces []ui.Namespace, id int64) bool {
	_, ok := findByID(namespaces, id, namespaceID)
	return ok
}

func hasPolicyID(policies []ui.Policy, id int64) bool {
	_, ok := findByID(policies, id, policyID)
	return ok
}

func hasSourceID(sources []ui.Source, id int64) bool {
	_, ok := findByID(sources, id, sourceID)
	return ok
}

func hasDependencyID(dependencies []ui.Dependency, id int64) bool {
	_, ok := findByID(dependencies, id, dependencyID)
	return ok
}

func previousStackID(stacks []ui.Stack, id int64) int64 {
	return selectPrevious(stacks, id, stackID)
}

func nextStackID(stacks []ui.Stack, id int64) int64 {
	return selectNext(stacks, id, stackID)
}

func previousNamespaceID(namespaces []ui.Namespace, id int64) int64 {
	return selectPrevious(namespaces, id, namespaceID)
}

func nextNamespaceID(namespaces []ui.Namespace, id int64) int64 {
	return selectNext(namespaces, id, namespaceID)
}

func previousSourceID(sources []ui.Source, id int64) int64 {
	return selectPrevious(sources, id, sourceID)
}

func nextSourceID(sources []ui.Source, id int64) int64 {
	return selectNext(sources, id, sourceID)
}

func previousPolicyID(policies []ui.Policy, id int64) int64 {
	return selectPrevious(policies, id, policyID)
}

func nextPolicyID(policies []ui.Policy, id int64) int64 {
	return selectNext(policies, id, policyID)
}

func policyNameByID(policies []ui.Policy, id int64) string {
	for _, policy := range policies {
		if policy.ID == id {
			return policy.Name
		}
	}
	if len(policies) == 0 {
		return "No policies"
	}
	return ""
}

func previousDependencyID(dependencies []ui.Dependency, id int64) int64 {
	return selectPrevious(dependencies, id, dependencyID)
}

func nextDependencyID(dependencies []ui.Dependency, id int64) int64 {
	return selectNext(dependencies, id, dependencyID)
}

var sourceTypes = []string{
	storage.SourceTypeGitLab,
	storage.SourceTypeGitHub,
	storage.SourceTypeGitea,
	storage.SourceTypeBitbucket,
	storage.SourceTypeRegistry,
}

var registryKinds = []string{
	registry.KindNPM,
	registry.KindGo,
	registry.KindMaven,
	registry.KindRubyGems,
	registry.KindCocoaPods,
}

func hasSourceType(sourceType string) bool {
	for _, item := range sourceTypes {
		if item == sourceType {
			return true
		}
	}

	return false
}

func previousSourceType(sourceType string) string {
	if len(sourceTypes) == 0 {
		return ""
	}

	index := int(sourceTypeIndex(sourceType))
	if index <= 0 {
		index = len(sourceTypes)
	}

	return sourceTypes[index-1]
}

func nextSourceType(sourceType string) string {
	if len(sourceTypes) == 0 {
		return ""
	}

	index := int(sourceTypeIndex(sourceType))
	return sourceTypes[(index+1)%len(sourceTypes)]
}

func sourceTypeIndex(sourceType string) int {
	for index, item := range sourceTypes {
		if item == sourceType {
			return index
		}
	}

	return 0
}

func hasRegistryKind(kind string) bool {
	for _, item := range registryKinds {
		if item == kind {
			return true
		}
	}
	return false
}

func previousRegistryKind(kind string) string {
	if len(registryKinds) == 0 {
		return ""
	}
	index := int(registryKindIndex(kind))
	if index <= 0 {
		index = len(registryKinds)
	}
	return registryKinds[index-1]
}

func nextRegistryKind(kind string) string {
	if len(registryKinds) == 0 {
		return ""
	}
	index := int(registryKindIndex(kind))
	return registryKinds[(index+1)%len(registryKinds)]
}

func registryKindIndex(kind string) int {
	for index, item := range registryKinds {
		if item == kind {
			return index
		}
	}
	return 0
}

func registryUISources(sources []ui.Source) []ui.Source {
	result := make([]ui.Source, 0, len(sources))
	for _, source := range sources {
		if source.Type == storage.SourceTypeRegistry {
			result = append(result, source)
		}
	}
	return result
}

func vcsSources(sources []ui.Source) []ui.Source {
	result := make([]ui.Source, 0, len(sources))
	for _, source := range sources {
		if source.Type != storage.SourceTypeRegistry {
			result = append(result, source)
		}
	}
	return result
}

func policyStacks(values []ui.PolicyValue, stacks []ui.Stack) []ui.Stack {
	used := make(map[int64]bool, len(values))
	for _, value := range values {
		used[value.StackID] = true
	}
	result := make([]ui.Stack, 0, len(stacks))
	for _, stack := range stacks {
		if used[stack.ID] {
			result = append(result, stack)
		}
	}
	return result
}

func trimLastRune(value string) string {
	runes := []rune(value)
	if len(runes) == 0 {
		return value
	}

	return string(runes[:len(runes)-1])
}
