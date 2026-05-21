package app

import (
	"context"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dnwSilver/tld/internal/storage"
	"github.com/dnwSilver/tld/internal/ui"
)

type model struct {
	width           int
	height          int
	creator         ui.Creator
	store           *storage.Store
	screen          ui.Screen
	stacks          []ui.Stack
	selectedStackID int64
	form            ui.StackForm
	deleteConfirm   ui.DeleteConfirm
	err             error
}

type stacksLoadedMsg struct {
	stacks []ui.Stack
	err    error
}

type stackSavedMsg struct {
	stackID int64
	err     error
}

type stackDeletedMsg struct {
	stackID int64
	err     error
}

func newModel(store *storage.Store) model {
	return model{
		creator: ui.NewCreator(),
		store:   store,
		screen:  ui.ScreenDefault,
		stacks:  []ui.Stack{},
	}
}

func (m model) Init() tea.Cmd {
	return m.loadStacks()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		if m.deleteConfirm.Open {
			return m.updateDeleteConfirm(msg)
		}
		if m.form.Open {
			return m.updateStackForm(msg)
		}

		switch msg.String() {
		case "0", "cmd+0", "alt+0":
			m.screen = ui.ScreenDefault
		case "1", "cmd+1", "alt+1":
			m.screen = ui.ScreenStacks
		case "a":
			if m.screen == ui.ScreenStacks {
				m.form = newStackForm()
			}
		case "e":
			if m.screen == ui.ScreenStacks {
				m.openEditStackForm()
			}
		case "d":
			if m.screen == ui.ScreenStacks {
				m.openDeleteConfirm()
			}
		case "up", "h", "k":
			if m.screen == ui.ScreenStacks {
				m.selectPreviousStack()
			}
		case "down", "j":
			if m.screen == ui.ScreenStacks {
				m.selectNextStack()
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
	case stackSavedMsg:
		m.err = msg.err
		if msg.err == nil {
			m.selectedStackID = msg.stackID
			m.form = ui.StackForm{}
			return m, m.loadStacks()
		}
		m.form.Error = msg.err.Error()
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
		m.form,
		m.deleteConfirm,
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

func newStackForm() ui.StackForm {
	return ui.StackForm{
		Open:  true,
		Mode:  ui.StackFormModeCreate,
		Focus: ui.StackFormFieldIcon,
	}
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

func (m model) selectedStack() (ui.Stack, bool) {
	for _, stack := range m.stacks {
		if stack.ID == m.selectedStackID {
			return stack, true
		}
	}

	return ui.Stack{}, false
}

func (m model) selectedStackIndex() int {
	for index, stack := range m.stacks {
		if stack.ID == m.selectedStackID {
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

func trimLastRune(value string) string {
	runes := []rune(value)
	if len(runes) == 0 {
		return value
	}

	return string(runes[:len(runes)-1])
}
