package app

import (
	"context"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dnwSilver/tld/internal/storage"
	"github.com/dnwSilver/tld/internal/ui"
)

func TestScreenNavigation(t *testing.T) {
	m := newModel(nil)

	next, _ := m.Update(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'1'},
	})
	updated := next.(model)
	if updated.screen != ui.ScreenStacks {
		t.Fatalf("screen = %v, want %v", updated.screen, ui.ScreenStacks)
	}

	next, _ = updated.Update(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'0'},
	})
	updated = next.(model)
	if updated.screen != ui.ScreenDefault {
		t.Fatalf("screen = %v, want %v", updated.screen, ui.ScreenDefault)
	}
}

func TestAltScreenNavigation(t *testing.T) {
	m := newModel(nil)

	next, _ := m.Update(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'1'},
		Alt:   true,
	})
	updated := next.(model)
	if updated.screen != ui.ScreenStacks {
		t.Fatalf("screen = %v, want %v", updated.screen, ui.ScreenStacks)
	}

	next, _ = updated.Update(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'0'},
		Alt:   true,
	})
	updated = next.(model)
	if updated.screen != ui.ScreenDefault {
		t.Fatalf("screen = %v, want %v", updated.screen, ui.ScreenDefault)
	}
}

func TestOpenStackForm(t *testing.T) {
	m := newModel(nil)
	m.screen = ui.ScreenStacks

	next, _ := m.Update(key("a"))
	updated := next.(model)
	if !updated.form.Open {
		t.Fatal("form should be open")
	}
	if updated.form.Focus != ui.StackFormFieldIcon {
		t.Fatalf("focus = %v, want %v", updated.form.Focus, ui.StackFormFieldIcon)
	}
}

func TestStackFormInputAndCancel(t *testing.T) {
	m := newModel(nil)
	m.screen = ui.ScreenStacks
	next, _ := m.Update(key("a"))
	updated := next.(model)

	next, _ = updated.Update(key(""))
	updated = next.(model)
	next, _ = updated.Update(tea.KeyMsg{Type: tea.KeyTab})
	updated = next.(model)
	next, _ = updated.Update(key("#84BA64"))
	updated = next.(model)
	next, _ = updated.Update(tea.KeyMsg{Type: tea.KeyTab})
	updated = next.(model)
	next, _ = updated.Update(key("Git"))
	updated = next.(model)
	if !updated.form.CanSave {
		t.Fatal("form should be valid")
	}

	next, _ = updated.Update(tea.KeyMsg{Type: tea.KeyEsc})
	updated = next.(model)
	if updated.form.Open {
		t.Fatal("form should be closed")
	}
}

func TestCreateStackFromForm(t *testing.T) {
	ctx := context.Background()
	store, err := storage.Open(ctx, filepath.Join(t.TempDir(), "tld.db"), "secret")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer func() {
		_ = store.Close()
	}()

	m := newModel(store)
	m.screen = ui.ScreenStacks
	m.form = ui.StackForm{
		Open:    true,
		Focus:   ui.StackFormFieldName,
		Icon:    "",
		Color:   "#84BA64",
		Name:    "Git",
		CanSave: true,
	}

	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := next.(model)
	if cmd == nil {
		t.Fatal("expected create command")
	}

	next, cmd = updated.Update(cmd().(stackSavedMsg))
	updated = next.(model)
	if updated.form.Open {
		t.Fatal("form should be closed after save")
	}
	if cmd == nil {
		t.Fatal("expected reload command")
	}

	next, _ = updated.Update(cmd().(stacksLoadedMsg))
	updated = next.(model)
	if len(updated.stacks) != 1 {
		t.Fatalf("len(stacks) = %d, want 1", len(updated.stacks))
	}
	if updated.stacks[0].Name != "Git" {
		t.Fatalf("stack name = %q, want Git", updated.stacks[0].Name)
	}
	if updated.selectedStackID != updated.stacks[0].ID {
		t.Fatalf("selected stack = %d, want %d", updated.selectedStackID, updated.stacks[0].ID)
	}
}

func TestStackSelectionNavigation(t *testing.T) {
	m := newModel(nil)
	m.screen = ui.ScreenStacks
	m.stacks = []ui.Stack{
		{ID: 1, Name: "Go"},
		{ID: 2, Name: "Java"},
	}
	m.selectedStackID = 1

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	updated := next.(model)
	if updated.selectedStackID != 2 {
		t.Fatalf("selected stack = %d, want 2", updated.selectedStackID)
	}

	next, _ = updated.Update(key("h"))
	updated = next.(model)
	if updated.selectedStackID != 1 {
		t.Fatalf("selected stack = %d, want 1", updated.selectedStackID)
	}
}

func TestOpenEditStackForm(t *testing.T) {
	m := newModel(nil)
	m.screen = ui.ScreenStacks
	m.stacks = []ui.Stack{{ID: 1, Icon: "", Name: "Git", Color: "#84BA64"}}
	m.selectedStackID = 1

	next, _ := m.Update(key("e"))
	updated := next.(model)
	if !updated.form.Open {
		t.Fatal("form should be open")
	}
	if updated.form.Mode != ui.StackFormModeEdit {
		t.Fatalf("form mode = %v, want edit", updated.form.Mode)
	}
	if updated.form.StackID != 1 || updated.form.Name != "Git" {
		t.Fatalf("form = %#v", updated.form)
	}
}

func TestOpenDeleteConfirm(t *testing.T) {
	m := newModel(nil)
	m.screen = ui.ScreenStacks
	m.stacks = []ui.Stack{{ID: 1, Name: "Git"}}
	m.selectedStackID = 1

	next, _ := m.Update(key("d"))
	updated := next.(model)
	if !updated.deleteConfirm.Open {
		t.Fatal("delete confirm should be open")
	}
	if updated.deleteConfirm.StackID != 1 || updated.deleteConfirm.Name != "Git" {
		t.Fatalf("delete confirm = %#v", updated.deleteConfirm)
	}
}

func key(value string) tea.KeyMsg {
	return tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune(value),
	}
}
