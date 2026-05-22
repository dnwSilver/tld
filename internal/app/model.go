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
	form                    ui.StackForm
	namespaceForm           ui.StackForm
	dependencyForm          ui.DependencyForm
	deleteConfirm           ui.DeleteConfirm
	namespaceDeleteConfirm  ui.DeleteConfirm
	dependencyDeleteConfirm ui.DeleteConfirm
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

func newModel(store *storage.Store) model {
	return model{
		creator:      ui.NewCreator(),
		store:        store,
		screen:       ui.ScreenDefault,
		stacks:       []ui.Stack{},
		namespaces:   []ui.Namespace{},
		dependencies: []ui.Dependency{},
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.loadStacks(), m.loadNamespaces(), m.loadDependencies())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
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
		if m.namespaceForm.Open {
			return m.updateNamespaceForm(msg)
		}
		if m.form.Open {
			return m.updateStackForm(msg)
		}

		switch msg.String() {
		case "0", "cmd+0", "alt+0":
			m.screen = ui.ScreenDefault
		case "1", "cmd+1", "alt+1":
			m.screen = ui.ScreenStacks
		case "2", "cmd+2", "alt+2":
			m.screen = ui.ScreenNamespaces
		case "3", "cmd+3", "alt+3":
			m.screen = ui.ScreenDependencies
		case "a":
			if m.screen == ui.ScreenStacks {
				m.form = newStackForm()
			}
			if m.screen == ui.ScreenNamespaces {
				m.namespaceForm = newStackForm()
			}
			if m.screen == ui.ScreenDependencies {
				m.dependencyForm = newDependencyForm(m.stacks)
			}
		case "e":
			if m.screen == ui.ScreenStacks {
				m.openEditStackForm()
			}
			if m.screen == ui.ScreenNamespaces {
				m.openEditNamespaceForm()
			}
			if m.screen == ui.ScreenDependencies {
				m.openEditDependencyForm()
			}
		case "d":
			if m.screen == ui.ScreenStacks {
				m.openDeleteConfirm()
			}
			if m.screen == ui.ScreenNamespaces {
				m.openNamespaceDeleteConfirm()
			}
			if m.screen == ui.ScreenDependencies {
				m.openDependencyDeleteConfirm()
			}
		case "up", "h", "k":
			if m.screen == ui.ScreenStacks {
				m.selectPreviousStack()
			}
			if m.screen == ui.ScreenNamespaces {
				m.selectPreviousNamespace()
			}
			if m.screen == ui.ScreenDependencies {
				m.selectPreviousDependency()
			}
		case "down", "j":
			if m.screen == ui.ScreenStacks {
				m.selectNextStack()
			}
			if m.screen == ui.ScreenNamespaces {
				m.selectNextNamespace()
			}
			if m.screen == ui.ScreenDependencies {
				m.selectNextDependency()
			}
		case "ctrl+c", "esc", "q":
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
			return m, m.loadNamespaces()
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
			return m, m.loadNamespaces()
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
		m.form,
		m.namespaceForm,
		m.dependencyForm,
		m.deleteConfirm,
		m.namespaceDeleteConfirm,
		m.dependencyDeleteConfirm,
	)
}

func (m model) updateStackForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.form = ui.StackForm{}
		return m, nil
	case "enter":
		if !m.form.CanSave {
			return m, nil
		}
		return m, m.saveStack()
	case "tab", "down":
		m.form.Focus = nextStackFormField(m.form.Focus)
	case "shift+tab", "up":
		m.form.Focus = previousStackFormField(m.form.Focus)
	case "backspace":
		m.form = deleteStackFormRune(m.form)
	default:
		if msg.Type == tea.KeyRunes {
			m.form = appendStackFormRunes(m.form, msg.Runes)
		}
	}

	m.form = normalizeStackForm(m.form)
	return m, nil
}

func (m model) updateNamespaceForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.namespaceForm = ui.StackForm{}
		return m, nil
	case "enter":
		if !m.namespaceForm.CanSave {
			return m, nil
		}
		return m, m.saveNamespace()
	case "tab", "down":
		m.namespaceForm.Focus = nextStackFormField(m.namespaceForm.Focus)
	case "shift+tab", "up":
		m.namespaceForm.Focus = previousStackFormField(m.namespaceForm.Focus)
	case "backspace":
		m.namespaceForm = deleteStackFormRune(m.namespaceForm)
	default:
		if msg.Type == tea.KeyRunes {
			m.namespaceForm = appendStackFormRunes(m.namespaceForm, msg.Runes)
		}
	}

	m.namespaceForm = normalizeStackForm(m.namespaceForm)
	return m, nil
}

func (m model) updateDependencyForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.dependencyForm = ui.DependencyForm{}
		return m, nil
	case "enter":
		if !m.dependencyForm.CanSave {
			return m, nil
		}
		return m, m.saveDependency()
	case "tab", "down":
		m.dependencyForm.Focus = nextDependencyFormField(m.dependencyForm.Focus)
	case "shift+tab", "up":
		m.dependencyForm.Focus = previousDependencyFormField(m.dependencyForm.Focus)
	case "left", "h":
		if m.dependencyForm.Focus == ui.DependencyFormFieldStack {
			m.dependencyForm.StackID = previousStackID(m.stacks, m.dependencyForm.StackID)
		}
	case "right", "j":
		if m.dependencyForm.Focus == ui.DependencyFormFieldStack {
			m.dependencyForm.StackID = nextStackID(m.stacks, m.dependencyForm.StackID)
		}
	case "backspace":
		m.dependencyForm = deleteDependencyFormRune(m.dependencyForm)
	default:
		if msg.Type == tea.KeyRunes {
			m.dependencyForm = appendDependencyFormRunes(m.dependencyForm, msg.Runes)
		}
	}

	m.dependencyForm = normalizeDependencyForm(m.dependencyForm, m.stacks)
	return m, nil
}

func (m model) updateDeleteConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.deleteConfirm = ui.DeleteConfirm{}
	case "enter":
		return m, m.deleteStack()
	}

	return m, nil
}

func (m model) updateNamespaceDeleteConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.namespaceDeleteConfirm = ui.DeleteConfirm{}
	case "enter":
		return m, m.deleteNamespace()
	}

	return m, nil
}

func (m model) updateDependencyDeleteConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.dependencyDeleteConfirm = ui.DeleteConfirm{}
	case "enter":
		return m, m.deleteDependency()
	}

	return m, nil
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
			err := m.store.Namespaces().Update(context.Background(), form.StackID, icon, name, color)
			return namespaceSavedMsg{namespaceID: form.StackID, err: err}
		}

		namespace, err := m.store.Namespaces().Create(context.Background(), icon, name, color)
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
			ID:    namespace.ID,
			Icon:  namespace.Icon,
			Name:  namespace.Name,
			Color: namespace.Color,
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

func newStackForm() ui.StackForm {
	return ui.StackForm{
		Open:  true,
		Mode:  ui.StackFormModeCreate,
		Focus: ui.StackFormFieldIcon,
	}
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

func (m *model) openEditStackForm() {
	stack, ok := m.selectedStack()
	if !ok {
		return
	}

	m.form = normalizeStackForm(ui.StackForm{
		Open:    true,
		Mode:    ui.StackFormModeEdit,
		StackID: stack.ID,
		Focus:   ui.StackFormFieldIcon,
		Icon:    stack.Icon,
		Color:   stack.Color,
		Name:    stack.Name,
	})
}

func (m *model) openDeleteConfirm() {
	stack, ok := m.selectedStack()
	if !ok {
		return
	}

	m.deleteConfirm = ui.DeleteConfirm{
		Open:    true,
		StackID: stack.ID,
		Name:    stack.Name,
	}
}

func (m *model) openEditNamespaceForm() {
	namespace, ok := m.selectedNamespace()
	if !ok {
		return
	}

	m.namespaceForm = normalizeStackForm(ui.StackForm{
		Open:    true,
		Mode:    ui.StackFormModeEdit,
		StackID: namespace.ID,
		Focus:   ui.StackFormFieldIcon,
		Icon:    namespace.Icon,
		Color:   namespace.Color,
		Name:    namespace.Name,
	})
}

func (m *model) openNamespaceDeleteConfirm() {
	namespace, ok := m.selectedNamespace()
	if !ok {
		return
	}

	m.namespaceDeleteConfirm = ui.DeleteConfirm{
		Open:    true,
		StackID: namespace.ID,
		Name:    namespace.Name,
	}
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

	m.dependencyDeleteConfirm = ui.DeleteConfirm{
		Open:    true,
		StackID: dependency.ID,
		Name:    dependency.Name,
	}
}

func (m *model) ensureSelectedStack() {
	if len(m.stacks) == 0 {
		m.selectedStackID = 0
		return
	}

	for _, stack := range m.stacks {
		if stack.ID == m.selectedStackID {
			return
		}
	}

	m.selectedStackID = m.stacks[0].ID
}

func (m *model) ensureSelectedNamespace() {
	if len(m.namespaces) == 0 {
		m.selectedNamespaceID = 0
		return
	}

	for _, namespace := range m.namespaces {
		if namespace.ID == m.selectedNamespaceID {
			return
		}
	}

	m.selectedNamespaceID = m.namespaces[0].ID
}

func (m *model) ensureSelectedDependency() {
	if len(m.dependencies) == 0 {
		m.selectedDependencyID = 0
		return
	}

	for _, dependency := range m.dependencies {
		if dependency.ID == m.selectedDependencyID {
			return
		}
	}

	m.selectedDependencyID = m.dependencies[0].ID
}

func (m *model) selectPreviousStack() {
	if len(m.stacks) == 0 {
		return
	}

	index := m.selectedStackIndex()
	if index <= 0 {
		index = len(m.stacks)
	}
	m.selectedStackID = m.stacks[index-1].ID
}

func (m *model) selectNextStack() {
	if len(m.stacks) == 0 {
		return
	}

	index := m.selectedStackIndex()
	m.selectedStackID = m.stacks[(index+1)%len(m.stacks)].ID
}

func (m *model) selectPreviousNamespace() {
	if len(m.namespaces) == 0 {
		return
	}

	index := m.selectedNamespaceIndex()
	if index <= 0 {
		index = len(m.namespaces)
	}
	m.selectedNamespaceID = m.namespaces[index-1].ID
}

func (m *model) selectNextNamespace() {
	if len(m.namespaces) == 0 {
		return
	}

	index := m.selectedNamespaceIndex()
	m.selectedNamespaceID = m.namespaces[(index+1)%len(m.namespaces)].ID
}

func (m *model) selectPreviousDependency() {
	if len(m.dependencies) == 0 {
		return
	}

	index := m.selectedDependencyIndex()
	if index <= 0 {
		index = len(m.dependencies)
	}
	m.selectedDependencyID = m.dependencies[index-1].ID
}

func (m *model) selectNextDependency() {
	if len(m.dependencies) == 0 {
		return
	}

	index := m.selectedDependencyIndex()
	m.selectedDependencyID = m.dependencies[(index+1)%len(m.dependencies)].ID
}

func (m model) selectedStack() (ui.Stack, bool) {
	for _, stack := range m.stacks {
		if stack.ID == m.selectedStackID {
			return stack, true
		}
	}

	return ui.Stack{}, false
}

func (m model) selectedNamespace() (ui.Namespace, bool) {
	for _, namespace := range m.namespaces {
		if namespace.ID == m.selectedNamespaceID {
			return namespace, true
		}
	}

	return ui.Namespace{}, false
}

func (m model) selectedDependency() (ui.Dependency, bool) {
	for _, dependency := range m.dependencies {
		if dependency.ID == m.selectedDependencyID {
			return dependency, true
		}
	}

	return ui.Dependency{}, false
}

func (m model) selectedStackIndex() int {
	for index, stack := range m.stacks {
		if stack.ID == m.selectedStackID {
			return index
		}
	}

	return 0
}

func (m model) selectedNamespaceIndex() int {
	for index, namespace := range m.namespaces {
		if namespace.ID == m.selectedNamespaceID {
			return index
		}
	}

	return 0
}

func (m model) selectedDependencyIndex() int {
	for index, dependency := range m.dependencies {
		if dependency.ID == m.selectedDependencyID {
			return index
		}
	}

	return 0
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

func normalizeStackForm(form ui.StackForm) ui.StackForm {
	form.CanSave = strings.TrimSpace(form.Icon) != "" &&
		strings.TrimSpace(form.Color) != "" &&
		strings.TrimSpace(form.Name) != ""
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

func hasStackID(stacks []ui.Stack, stackID int64) bool {
	for _, stack := range stacks {
		if stack.ID == stackID {
			return true
		}
	}

	return false
}

func previousStackID(stacks []ui.Stack, stackID int64) int64 {
	if len(stacks) == 0 {
		return 0
	}

	index := stackIndex(stacks, stackID)
	if index <= 0 {
		index = len(stacks)
	}

	return stacks[index-1].ID
}

func nextStackID(stacks []ui.Stack, stackID int64) int64 {
	if len(stacks) == 0 {
		return 0
	}

	index := stackIndex(stacks, stackID)
	return stacks[(index+1)%len(stacks)].ID
}

func stackIndex(stacks []ui.Stack, stackID int64) int {
	for index, stack := range stacks {
		if stack.ID == stackID {
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
