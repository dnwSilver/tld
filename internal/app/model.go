package app

import (
	"context"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dnwSilver/tld/internal/storage"
	"github.com/dnwSilver/tld/internal/ui"
)

type model struct {
	width                   int
	height                  int
	creator                 ui.Creator
	store                   *storage.Store
	screen                  ui.Screen
	stacks                  []ui.Stack
	selectedStackID         int64
	namespaces              []ui.Namespace
	selectedNamespaceID     int64
	dependencies            []ui.Dependency
	selectedDependencyID    int64
	sources                 []ui.Source
	selectedSourceID        int64
	policies                []ui.Policy
	selectedPolicyID        int64
	policyValues            []ui.PolicyValue
	selectedPolicyValueID   int64
	policyFocus             ui.PolicyPane
	form                    ui.StackForm
	namespaceForm           ui.StackForm
	dependencyForm          ui.DependencyForm
	sourceForm              ui.SourceForm
	policyForm              ui.PolicyForm
	policyValueForm         ui.PolicyValueForm
	deleteConfirm           ui.DeleteConfirm
	namespaceDeleteConfirm  ui.DeleteConfirm
	dependencyDeleteConfirm ui.DeleteConfirm
	sourceDeleteConfirm     ui.DeleteConfirm
	policyDeleteConfirm     ui.DeleteConfirm
	err                     error
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

func newModel(store *storage.Store) model {
	return model{
		creator:      ui.NewCreator(),
		store:        store,
		screen:       ui.ScreenDefault,
		stacks:       []ui.Stack{},
		namespaces:   []ui.Namespace{},
		dependencies: []ui.Dependency{},
		sources:      []ui.Source{},
		policies:     []ui.Policy{},
		policyValues: []ui.PolicyValue{},
		policyFocus:  ui.PolicyPanePolicies,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.loadStacks(), m.loadNamespaces(), m.loadDependencies(), m.loadSources(), m.loadPolicies())
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
		if m.namespaceDeleteConfirm.Open {
			return m.updateNamespaceDeleteConfirm(msg)
		}
		if m.deleteConfirm.Open {
			return m.updateDeleteConfirm(msg)
		}
		if m.dependencyForm.Open {
			return m.updateDependencyForm(msg)
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
		if m.namespaceForm.Open {
			return m.updateNamespaceForm(msg)
		}
		if m.form.Open {
			return m.updateStackForm(msg)
		}

		key := msg.String()
		switch {
		case ui.KeyHome.Matches(key):
			m.screen = ui.ScreenDefault
		case ui.KeyStacks.Matches(key):
			m.screen = ui.ScreenStacks
		case ui.KeyNamespaces.Matches(key):
			m.screen = ui.ScreenNamespaces
		case ui.KeyDependencies.Matches(key):
			m.screen = ui.ScreenDependencies
		case ui.KeySources.Matches(key):
			m.screen = ui.ScreenSources
		case ui.KeyPolicies.Matches(key):
			m.screen = ui.ScreenPolicies
		case ui.KeyAdd.Matches(key):
			m.openAddForm()
		case ui.KeyEdit.Matches(key):
			m.openEditForm()
		case ui.KeyDelete.Matches(key):
			m.openDelete()
		case ui.KeyPrev.Matches(key):
			m.selectPreviousOnScreen()
			if m.screen == ui.ScreenPolicies && m.policyFocus == ui.PolicyPanePolicies {
				return m, m.loadPolicyValues()
			}
		case ui.KeyNext.Matches(key):
			m.selectNextOnScreen()
			if m.screen == ui.ScreenPolicies && m.policyFocus == ui.PolicyPanePolicies {
				return m, m.loadPolicyValues()
			}
		case key == "tab":
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
	case stackSavedMsg:
		m.err = msg.err
		if msg.err == nil {
			m.selectedStackID = msg.stackID
			m.form = ui.StackForm{}
			return m, tea.Batch(m.loadStacks(), m.loadDependencies())
		}
		m.form.Error = msg.err.Error()
	case namespaceSavedMsg:
		m.err = msg.err
		if msg.err == nil {
			m.selectedNamespaceID = msg.namespaceID
			m.namespaceForm = ui.StackForm{}
			return m, tea.Batch(m.loadNamespaces(), m.loadPolicies())
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
	case sourceSavedMsg:
		m.err = msg.err
		if msg.err == nil {
			m.selectedSourceID = msg.sourceID
			m.sourceForm = ui.SourceForm{}
			return m, m.loadSources()
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
		m.selectedStackID,
		m.namespaces,
		m.selectedNamespaceID,
		m.dependencies,
		m.selectedDependencyID,
		m.sources,
		m.selectedSourceID,
		m.policies,
		m.selectedPolicyID,
		m.policyValues,
		m.selectedPolicyValueID,
		m.policyFocus,
		m.form,
		m.namespaceForm,
		m.dependencyForm,
		m.sourceForm,
		m.policyForm,
		m.policyValueForm,
		m.deleteConfirm,
		m.namespaceDeleteConfirm,
		m.dependencyDeleteConfirm,
		m.sourceDeleteConfirm,
		m.policyDeleteConfirm,
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
	case isOneOf(msg, "left", "h"):
		if form.Focus == ui.StackFormFieldPolicy {
			form.PolicyID = previousPolicyID(policies, form.PolicyID)
		}
	case isOneOf(msg, "right", "l"):
		if form.Focus == ui.StackFormFieldPolicy {
			form.PolicyID = nextPolicyID(policies, form.PolicyID)
		}
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
	m.dependencyForm, action = updateDependencyFormState(msg, m.dependencyForm, m.stacks)
	return m.finishStackFormAction(action, m.saveDependency)
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
	var action stackFormAction
	m.policyValueForm, action = updatePolicyValueFormState(msg, m.policyValueForm, m.dependencies)
	return m.finishStackFormAction(action, m.savePolicyValue)
}

func updateDependencyFormState(msg tea.KeyMsg, form ui.DependencyForm, stacks []ui.Stack) (ui.DependencyForm, stackFormAction) {
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
	case isOneOf(msg, "left", "h"):
		if form.Focus == ui.DependencyFormFieldStack {
			form.StackID = previousStackID(stacks, form.StackID)
		}
	case isOneOf(msg, "right", "l"):
		if form.Focus == ui.DependencyFormFieldStack {
			form.StackID = nextStackID(stacks, form.StackID)
		}
	case isBackspaceKey(msg):
		form = deleteDependencyFormRune(form)
	default:
		if msg.Type == tea.KeyRunes {
			form = appendDependencyFormRunes(form, msg.Runes)
		}
	}

	return normalizeDependencyForm(form, stacks), stackFormActionNone
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
		form.Focus = nextSourceFormField(form.Focus)
	case isOneOf(msg, "shift+tab", "up"):
		form.Focus = previousSourceFormField(form.Focus)
	case isOneOf(msg, "left", "h"):
		if form.Focus == ui.SourceFormFieldType {
			form.Type = previousSourceType(form.Type)
		}
	case isOneOf(msg, "right", "l"):
		if form.Focus == ui.SourceFormFieldType {
			form.Type = nextSourceType(form.Type)
		}
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
	case isOneOf(msg, "left", "h"):
		if form.Focus == ui.PolicyFormFieldNamespace {
			form.NamespaceID = previousNamespaceID(namespaces, form.NamespaceID)
		}
	case isOneOf(msg, "right", "l"):
		if form.Focus == ui.PolicyFormFieldNamespace {
			form.NamespaceID = nextNamespaceID(namespaces, form.NamespaceID)
		}
	case isBackspaceKey(msg):
		form = deletePolicyFormRune(form)
	default:
		if msg.Type == tea.KeyRunes {
			form = appendPolicyFormRunes(form, msg.Runes)
		}
	}

	return normalizePolicyForm(form, namespaces), stackFormActionNone
}

func updatePolicyValueFormState(msg tea.KeyMsg, form ui.PolicyValueForm, dependencies []ui.Dependency) (ui.PolicyValueForm, stackFormAction) {
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
	case isOneOf(msg, "left", "h"):
		if form.Focus == ui.PolicyValueFormFieldDependency {
			form.DependencyID = previousDependencyID(dependencies, form.DependencyID)
		}
	case isOneOf(msg, "right", "l"):
		if form.Focus == ui.PolicyValueFormFieldDependency {
			form.DependencyID = nextDependencyID(dependencies, form.DependencyID)
		}
	case isBackspaceKey(msg):
		form = deletePolicyValueFormRune(form)
	default:
		if msg.Type == tea.KeyRunes {
			form = appendPolicyValueFormRunes(form, msg.Runes)
		}
	}

	return normalizePolicyValueForm(form, dependencies), stackFormActionNone
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
		m.dependencyForm = newDependencyForm(m.stacks)
	case ui.ScreenSources:
		m.sourceForm = newSourceForm()
	case ui.ScreenPolicies:
		if m.policyFocus == ui.PolicyPaneValues {
			if m.selectedPolicyID == 0 {
				return
			}
			m.policyValueForm = newPolicyValueForm(m.selectedPolicyID, m.dependencies)
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
	case ui.ScreenStacks:
		m.selectPreviousStack()
	case ui.ScreenNamespaces:
		m.selectPreviousNamespace()
	case ui.ScreenDependencies:
		m.selectPreviousDependency()
	case ui.ScreenSources:
		m.selectPreviousSource()
	case ui.ScreenPolicies:
		if m.policyFocus == ui.PolicyPaneValues {
			m.selectPreviousPolicyValue()
			return
		}
		m.selectPreviousPolicy()
	}
}

func (m *model) selectNextOnScreen() {
	switch m.screen {
	case ui.ScreenStacks:
		m.selectNextStack()
	case ui.ScreenNamespaces:
		m.selectNextNamespace()
	case ui.ScreenDependencies:
		m.selectNextDependency()
	case ui.ScreenSources:
		m.selectNextSource()
	case ui.ScreenPolicies:
		if m.policyFocus == ui.PolicyPaneValues {
			m.selectNextPolicyValue()
			return
		}
		m.selectNextPolicy()
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
		if form.Mode == ui.StackFormModeEdit {
			err := m.store.Dependencies().Update(context.Background(), form.DependencyID, form.StackID, icon, name, color)
			return dependencySavedMsg{dependencyID: form.DependencyID, err: err}
		}

		dependency, err := m.store.Dependencies().Create(context.Background(), form.StackID, icon, name, color)
		return dependencySavedMsg{dependencyID: dependency.ID, err: err}
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
		if form.Mode == ui.StackFormModeEdit {
			err := m.store.Sources().Update(context.Background(), form.SourceID, name, patToken, sourceURL, sourceType)
			return sourceSavedMsg{sourceID: form.SourceID, err: err}
		}

		source, err := m.store.Sources().Create(context.Background(), name, patToken, sourceURL, sourceType)
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
			ID:        dependency.ID,
			StackID:   dependency.StackID,
			StackName: dependency.StackName,
			Icon:      dependency.Icon,
			Name:      dependency.Name,
			Color:     dependency.Color,
		})
	}

	return result
}

func toUISources(sources []storage.Source) []ui.Source {
	result := make([]ui.Source, 0, len(sources))
	for _, source := range sources {
		result = append(result, ui.Source{
			ID:       source.ID,
			Icon:     source.Icon,
			Name:     source.Name,
			Color:    source.Color,
			PATToken: source.PATToken,
			URL:      source.URL,
			Type:     source.Type,
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
			ID:              value.ID,
			PolicyID:        value.PolicyID,
			DependencyID:    value.DependencyID,
			DependencyIcon:  value.DependencyIcon,
			DependencyName:  value.DependencyName,
			DependencyColor: value.DependencyColor,
			Version:         value.Version,
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

func newDependencyForm(stacks []ui.Stack) ui.DependencyForm {
	form := ui.DependencyForm{
		Open:  true,
		Mode:  ui.StackFormModeCreate,
		Focus: ui.DependencyFormFieldIcon,
	}
	if len(stacks) > 0 {
		form.StackID = stacks[0].ID
	}

	return normalizeDependencyForm(form, stacks)
}

func newSourceForm() ui.SourceForm {
	return normalizeSourceForm(ui.SourceForm{
		Open:  true,
		Mode:  ui.StackFormModeCreate,
		Focus: ui.SourceFormFieldName,
		Type:  storage.SourceTypeGitLab,
	})
}

func newPolicyForm(namespaces []ui.Namespace, policies []ui.Policy) ui.PolicyForm {
	return normalizePolicyForm(ui.PolicyForm{
		Open:  true,
		Mode:  ui.StackFormModeCreate,
		Focus: ui.PolicyFormFieldName,
	}, namespaces)
}

func newPolicyValueForm(policyID int64, dependencies []ui.Dependency) ui.PolicyValueForm {
	form := ui.PolicyValueForm{
		Open:     true,
		Mode:     ui.StackFormModeCreate,
		PolicyID: policyID,
		Focus:    ui.PolicyValueFormFieldDependency,
	}
	if len(dependencies) > 0 {
		form.DependencyID = dependencies[0].ID
	}
	return normalizePolicyValueForm(form, dependencies)
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
	}, m.stacks)
}

func (m *model) openDependencyDeleteConfirm() {
	dependency, ok := m.selectedDependency()
	if !ok {
		return
	}

	m.dependencyDeleteConfirm = deleteConfirm(dependency.ID, dependency.Name)
}

func (m *model) openEditSourceForm() {
	source, ok := m.selectedSource()
	if !ok {
		return
	}

	m.sourceForm = normalizeSourceForm(ui.SourceForm{
		Open:     true,
		Mode:     ui.StackFormModeEdit,
		SourceID: source.ID,
		Focus:    ui.SourceFormFieldName,
		Name:     source.Name,
		URL:      source.URL,
		PATToken: source.PATToken,
		Type:     source.Type,
	})
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
	}, m.dependencies)
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
func sourceID(s ui.Source) int64         { return s.ID }
func policyID(p ui.Policy) int64         { return p.ID }
func policyValueID(v ui.PolicyValue) int64 {
	return v.ID
}

func (m *model) ensureSelectedStack() {
	m.selectedStackID = ensureSelected(m.stacks, m.selectedStackID, stackID)
}

func (m *model) ensureSelectedNamespace() {
	m.selectedNamespaceID = ensureSelected(m.namespaces, m.selectedNamespaceID, namespaceID)
}

func (m *model) ensureSelectedDependency() {
	m.selectedDependencyID = ensureSelected(m.dependencies, m.selectedDependencyID, dependencyID)
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

func (m *model) selectNextDependency() {
	m.selectedDependencyID = selectNext(m.dependencies, m.selectedDependencyID, dependencyID)
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

func nextSourceFormField(field ui.SourceFormField) ui.SourceFormField {
	if field == ui.SourceFormFieldType {
		return ui.SourceFormFieldName
	}

	return field + 1
}

func previousSourceFormField(field ui.SourceFormField) ui.SourceFormField {
	if field == ui.SourceFormFieldName {
		return ui.SourceFormFieldType
	}

	return field - 1
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
	case ui.DependencyFormFieldStack:
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
	case ui.DependencyFormFieldStack:
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

func normalizeDependencyForm(form ui.DependencyForm, stacks []ui.Stack) ui.DependencyForm {
	if !hasStackID(stacks, form.StackID) && len(stacks) > 0 {
		form.StackID = stacks[0].ID
	}
	form.CanSave = strings.TrimSpace(form.Icon) != "" &&
		strings.TrimSpace(form.Color) != "" &&
		strings.TrimSpace(form.Name) != "" &&
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
	form.CanSave = strings.TrimSpace(form.Name) != "" &&
		strings.TrimSpace(form.URL) != "" &&
		strings.TrimSpace(form.PATToken) != "" &&
		hasSourceType(form.Type)
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

func normalizePolicyValueForm(form ui.PolicyValueForm, dependencies []ui.Dependency) ui.PolicyValueForm {
	if !hasDependencyID(dependencies, form.DependencyID) && len(dependencies) > 0 {
		form.DependencyID = dependencies[0].ID
	}
	form.CanSave = form.PolicyID != 0 &&
		strings.TrimSpace(form.Version) != "" &&
		hasDependencyID(dependencies, form.DependencyID)
	if form.CanSave {
		form.Error = ""
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

func trimLastRune(value string) string {
	runes := []rune(value)
	if len(runes) == 0 {
		return value
	}

	return string(runes[:len(runes)-1])
}
