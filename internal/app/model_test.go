package app

import (
	"context"
	"path/filepath"
	"strings"
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

	updated = applyBatch(t, updated, cmd)
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

func TestCreateDependencyFromForm(t *testing.T) {
	ctx := context.Background()
	store, err := storage.Open(ctx, filepath.Join(t.TempDir(), "tld.db"), "secret")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer func() {
		_ = store.Close()
	}()

	stack, err := store.Stacks().Create(ctx, "", "Git", "#84BA64")
	if err != nil {
		t.Fatalf("create stack: %v", err)
	}

	m := newModel(store)
	m.screen = ui.ScreenDependencies
	m.stacks = []ui.Stack{{ID: stack.ID, Icon: stack.Icon, Name: stack.Name, Color: stack.Color}}
	m.dependencyForm = ui.DependencyForm{
		Open:    true,
		Focus:   ui.DependencyFormFieldName,
		StackID: stack.ID,
		Icon:    "",
		Color:   "#EC9706",
		Name:    "Package",
		CanSave: true,
	}

	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := next.(model)
	if cmd == nil {
		t.Fatal("expected create command")
	}

	next, cmd = updated.Update(cmd().(dependencySavedMsg))
	updated = next.(model)
	if updated.dependencyForm.Open {
		t.Fatal("dependency form should be closed after save")
	}
	if cmd == nil {
		t.Fatal("expected reload command")
	}

	next, _ = updated.Update(cmd().(dependenciesLoadedMsg))
	updated = next.(model)
	if len(updated.dependencies) != 1 {
		t.Fatalf("len(dependencies) = %d, want 1", len(updated.dependencies))
	}
	if updated.dependencies[0].Name != "Package" {
		t.Fatalf("dependency name = %q, want Package", updated.dependencies[0].Name)
	}
	if updated.dependencies[0].StackID != stack.ID {
		t.Fatalf("dependency stack = %d, want %d", updated.dependencies[0].StackID, stack.ID)
	}
	if updated.selectedDependencyID != updated.dependencies[0].ID {
		t.Fatalf("selected dependency = %d, want %d", updated.selectedDependencyID, updated.dependencies[0].ID)
	}
}

func TestCreateSourceFromForm(t *testing.T) {
	ctx := context.Background()
	store, err := storage.Open(ctx, filepath.Join(t.TempDir(), "tld.db"), "secret")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer func() {
		_ = store.Close()
	}()

	m := newModel(store)
	m.screen = ui.ScreenSources
	m.sourceForm = ui.SourceForm{
		Open:     true,
		Focus:    ui.SourceFormFieldName,
		Name:     "GitLab",
		URL:      "https://gitlab.com",
		PATToken: "glpat-secret",
		Type:     storage.SourceTypeGitLab,
		CanSave:  true,
	}

	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := next.(model)
	if cmd == nil {
		t.Fatal("expected create command")
	}

	next, cmd = updated.Update(cmd().(sourceSavedMsg))
	updated = next.(model)
	if updated.sourceForm.Open {
		t.Fatal("source form should be closed after save")
	}
	if cmd == nil {
		t.Fatal("expected reload command")
	}

	updated = applyBatch(t, updated, cmd)
	if len(updated.sources) != 1 {
		t.Fatalf("len(sources) = %d, want 1", len(updated.sources))
	}
	if updated.sources[0].Name != "GitLab" || updated.sources[0].Type != storage.SourceTypeGitLab {
		t.Fatalf("source = %#v", updated.sources[0])
	}
	if updated.sources[0].Icon != "" || updated.sources[0].Color != "FC6D26" {
		t.Fatalf("source appearance = %#v", updated.sources[0])
	}
	if updated.selectedSourceID != updated.sources[0].ID {
		t.Fatalf("selected source = %d, want %d", updated.selectedSourceID, updated.sources[0].ID)
	}
}

func TestCreateProjectFromForm(t *testing.T) {
	ctx := context.Background()
	store, err := storage.Open(ctx, filepath.Join(t.TempDir(), "tld.db"), "secret")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer func() {
		_ = store.Close()
	}()

	namespace, err := store.Namespaces().Create(ctx, "󱃾", "Production", "#25799F")
	if err != nil {
		t.Fatalf("create namespace: %v", err)
	}
	stack, err := store.Stacks().Create(ctx, "", "Git", "#84BA64")
	if err != nil {
		t.Fatalf("create stack: %v", err)
	}
	source, err := store.Sources().Create(ctx, "GitLab", "glpat-secret", "https://gitlab.com", storage.SourceTypeGitLab)
	if err != nil {
		t.Fatalf("create source: %v", err)
	}

	m := newModel(store)
	m.screen = ui.ScreenProjects
	m.namespaces = []ui.Namespace{{ID: namespace.ID, Icon: namespace.Icon, Name: namespace.Name, Color: namespace.Color}}
	m.stacks = []ui.Stack{{ID: stack.ID, Icon: stack.Icon, Name: stack.Name, Color: stack.Color}}
	m.sources = []ui.Source{{ID: source.ID, Icon: source.Icon, Name: source.Name, Color: source.Color, URL: source.URL, PATToken: source.PATToken, Type: source.Type}}
	m.projectForm = ui.ProjectForm{
		Open:        true,
		Focus:       ui.ProjectFormFieldName,
		ProjectID:   "autobase-main",
		NamespaceID: namespace.ID,
		SourceID:    source.ID,
		StackID:     stack.ID,
		Icon:        "󰏖",
		Color:       "#EC9706",
		Name:        "TLD",
		CanSave:     true,
	}

	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := next.(model)
	if cmd == nil {
		t.Fatal("expected create command")
	}

	next, cmd = updated.Update(cmd().(projectSavedMsg))
	updated = next.(model)
	if updated.projectForm.Open {
		t.Fatal("project form should be closed after save")
	}
	if cmd == nil {
		t.Fatal("expected reload command")
	}

	next, _ = updated.Update(cmd().(projectsLoadedMsg))
	updated = next.(model)
	if len(updated.projects) != 1 {
		t.Fatalf("len(projects) = %d, want 1", len(updated.projects))
	}
	if updated.projects[0].ProjectID != "autobase-main" || updated.projects[0].Name != "TLD" || updated.projects[0].NamespaceID != namespace.ID || updated.projects[0].SourceID != source.ID || updated.projects[0].StackID != stack.ID {
		t.Fatalf("project = %#v", updated.projects[0])
	}
	if updated.selectedProjectID != updated.projects[0].ID {
		t.Fatalf("selected project = %d, want %d", updated.selectedProjectID, updated.projects[0].ID)
	}
}

func TestEditProjectCanChangeSelectors(t *testing.T) {
	ctx := context.Background()
	store, err := storage.Open(ctx, filepath.Join(t.TempDir(), "tld.db"), "secret")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer func() {
		_ = store.Close()
	}()

	namespaceA, err := store.Namespaces().Create(ctx, "󱃾", "A", "#25799F")
	if err != nil {
		t.Fatalf("create namespace A: %v", err)
	}
	namespaceB, err := store.Namespaces().Create(ctx, "󱃾", "B", "#EC9706")
	if err != nil {
		t.Fatalf("create namespace B: %v", err)
	}
	stackA, err := store.Stacks().Create(ctx, "", "Go", "#84BA64")
	if err != nil {
		t.Fatalf("create stack A: %v", err)
	}
	stackB, err := store.Stacks().Create(ctx, "󰙯", "Rust", "#F74C00")
	if err != nil {
		t.Fatalf("create stack B: %v", err)
	}
	sourceA, err := store.Sources().Create(ctx, "GitLab", "glpat-secret", "https://gitlab.com", storage.SourceTypeGitLab)
	if err != nil {
		t.Fatalf("create source A: %v", err)
	}
	sourceB, err := store.Sources().Create(ctx, "GitHub", "ghp-secret", "https://github.com", storage.SourceTypeGitHub)
	if err != nil {
		t.Fatalf("create source B: %v", err)
	}
	project, err := store.Projects().Create(ctx, "autobase-main", namespaceA.ID, sourceA.ID, stackA.ID, "󰏖", "TLD", "#EC9706")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	m := newModel(store)
	m.screen = ui.ScreenProjects
	m.namespaces = []ui.Namespace{
		{ID: namespaceA.ID, Icon: namespaceA.Icon, Name: namespaceA.Name, Color: namespaceA.Color},
		{ID: namespaceB.ID, Icon: namespaceB.Icon, Name: namespaceB.Name, Color: namespaceB.Color},
	}
	m.stacks = []ui.Stack{
		{ID: stackA.ID, Icon: stackA.Icon, Name: stackA.Name, Color: stackA.Color},
		{ID: stackB.ID, Icon: stackB.Icon, Name: stackB.Name, Color: stackB.Color},
	}
	m.sources = []ui.Source{
		{ID: sourceA.ID, Icon: sourceA.Icon, Name: sourceA.Name, Color: sourceA.Color, URL: sourceA.URL, PATToken: sourceA.PATToken, Type: sourceA.Type},
		{ID: sourceB.ID, Icon: sourceB.Icon, Name: sourceB.Name, Color: sourceB.Color, URL: sourceB.URL, PATToken: sourceB.PATToken, Type: sourceB.Type},
	}
	m.projects = []ui.Project{{
		ID:            project.ID,
		ProjectID:     project.ProjectID,
		NamespaceID:   namespaceA.ID,
		NamespaceName: namespaceA.Name,
		SourceID:      sourceA.ID,
		SourceName:    sourceA.Name,
		StackID:       stackA.ID,
		StackName:     stackA.Name,
		Icon:          project.Icon,
		Name:          project.Name,
		Color:         project.Color,
	}}
	m.selectedProjectID = project.ID

	next, _ := m.Update(key("e"))
	updated := next.(model)
	if !updated.projectForm.Open {
		t.Fatal("project edit form should be open")
	}

	next, _ = updated.Update(tea.KeyMsg{Type: tea.KeyTab})
	updated = next.(model)
	next, _ = updated.Update(tea.KeyMsg{Type: tea.KeyTab})
	updated = next.(model)
	next, _ = updated.Update(tea.KeyMsg{Type: tea.KeyTab})
	updated = next.(model)
	if updated.projectForm.Focus != ui.ProjectFormFieldNamespace {
		t.Fatalf("focus = %v, want namespace", updated.projectForm.Focus)
	}
	next, _ = updated.Update(key("l"))
	updated = next.(model)
	if updated.projectForm.NamespaceID != namespaceB.ID {
		t.Fatalf("namespace = %d, want %d", updated.projectForm.NamespaceID, namespaceB.ID)
	}

	next, _ = updated.Update(tea.KeyMsg{Type: tea.KeyTab})
	updated = next.(model)
	if updated.projectForm.Focus != ui.ProjectFormFieldSource {
		t.Fatalf("focus = %v, want source", updated.projectForm.Focus)
	}
	next, _ = updated.Update(key("l"))
	updated = next.(model)
	if updated.projectForm.SourceID != sourceB.ID {
		t.Fatalf("source = %d, want %d", updated.projectForm.SourceID, sourceB.ID)
	}

	next, _ = updated.Update(tea.KeyMsg{Type: tea.KeyTab})
	updated = next.(model)
	if updated.projectForm.Focus != ui.ProjectFormFieldStack {
		t.Fatalf("focus = %v, want stack", updated.projectForm.Focus)
	}
	next, _ = updated.Update(key("l"))
	updated = next.(model)
	if updated.projectForm.StackID != stackB.ID {
		t.Fatalf("stack = %d, want %d", updated.projectForm.StackID, stackB.ID)
	}
}

func TestCreatePolicyFromForm(t *testing.T) {
	ctx := context.Background()
	store, err := storage.Open(ctx, filepath.Join(t.TempDir(), "tld.db"), "secret")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer func() {
		_ = store.Close()
	}()

	namespace, err := store.Namespaces().Create(ctx, "󱃾", "Production", "#25799F")
	if err != nil {
		t.Fatalf("create namespace: %v", err)
	}

	m := newModel(store)
	m.screen = ui.ScreenPolicies
	m.namespaces = []ui.Namespace{{ID: namespace.ID, Icon: namespace.Icon, Name: namespace.Name, Color: namespace.Color}}
	m.policyForm = ui.PolicyForm{
		Open:    true,
		Focus:   ui.PolicyFormFieldName,
		Name:    "Production policy",
		CanSave: true,
	}

	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := next.(model)
	if cmd == nil {
		t.Fatal("expected create command")
	}

	next, cmd = updated.Update(cmd().(policySavedMsg))
	updated = next.(model)
	if updated.policyForm.Open {
		t.Fatal("policy form should be closed after save")
	}
	if cmd == nil {
		t.Fatal("expected reload command")
	}

	next, cmd = updated.Update(cmd().(policiesLoadedMsg))
	updated = next.(model)
	if len(updated.policies) != 1 {
		t.Fatalf("len(policies) = %d, want 1", len(updated.policies))
	}
	if updated.policies[0].Name != "Production policy" {
		t.Fatalf("policy name = %q, want Production policy", updated.policies[0].Name)
	}
	if cmd == nil {
		t.Fatal("expected policy values reload command")
	}
}

func TestCreatePolicyValueFromForm(t *testing.T) {
	ctx := context.Background()
	store, err := storage.Open(ctx, filepath.Join(t.TempDir(), "tld.db"), "secret")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer func() {
		_ = store.Close()
	}()

	stack, err := store.Stacks().Create(ctx, "", "Git", "#84BA64")
	if err != nil {
		t.Fatalf("create stack: %v", err)
	}
	dependency, err := store.Dependencies().Create(ctx, stack.ID, "", "Package", "#EC9706")
	if err != nil {
		t.Fatalf("create dependency: %v", err)
	}
	policy, err := store.Policies().Create(ctx, "Production policy")
	if err != nil {
		t.Fatalf("create policy: %v", err)
	}

	m := newModel(store)
	m.screen = ui.ScreenPolicies
	m.policyFocus = ui.PolicyPaneValues
	m.selectedPolicyID = policy.ID
	m.dependencies = []ui.Dependency{{ID: dependency.ID, StackID: stack.ID, StackName: stack.Name, Icon: dependency.Icon, Name: dependency.Name, Color: dependency.Color}}
	m.policyValueForm = ui.PolicyValueForm{
		Open:         true,
		Focus:        ui.PolicyValueFormFieldVersion,
		PolicyID:     policy.ID,
		DependencyID: dependency.ID,
		Version:      "1.2.3",
		CanSave:      true,
	}

	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := next.(model)
	if cmd == nil {
		t.Fatal("expected create command")
	}

	next, cmd = updated.Update(cmd().(policyValueSavedMsg))
	updated = next.(model)
	if updated.policyValueForm.Open {
		t.Fatal("policy value form should be closed after save")
	}
	if cmd == nil {
		t.Fatal("expected reload command")
	}

	next, _ = updated.Update(cmd().(policyValuesLoadedMsg))
	updated = next.(model)
	if len(updated.policyValues) != 1 {
		t.Fatalf("len(policyValues) = %d, want 1", len(updated.policyValues))
	}
	if updated.policyValues[0].Version != "1.2.3" {
		t.Fatalf("policy value version = %q, want 1.2.3", updated.policyValues[0].Version)
	}
}

func TestSourcePATIsMaskedInView(t *testing.T) {
	m := newModel(nil)
	m.width = 100
	m.height = 30
	m.screen = ui.ScreenSources
	m.sourceForm = ui.SourceForm{
		Open:     true,
		Focus:    ui.SourceFormFieldPATToken,
		Name:     "GitLab",
		URL:      "https://gitlab.com",
		PATToken: "glpat-secret",
		Type:     storage.SourceTypeGitLab,
		CanSave:  true,
	}

	view := m.View()
	if strings.Contains(view, "glpat-secret") {
		t.Fatal("view should not contain raw PAT token")
	}
	if !strings.Contains(view, "************") {
		t.Fatal("view should contain masked PAT token")
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

	next, _ = updated.Update(key("k"))
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

func TestDeleteConfirmEscClosesModal(t *testing.T) {
	m := newModel(nil)
	m.screen = ui.ScreenStacks
	m.stacks = []ui.Stack{{ID: 1, Name: "Git"}}
	m.selectedStackID = 1

	next, _ := m.Update(key("d"))
	updated := next.(model)
	if !updated.deleteConfirm.Open {
		t.Fatal("delete confirm should be open")
	}

	next, _ = updated.Update(tea.KeyMsg{Type: tea.KeyEsc})
	updated = next.(model)
	if updated.deleteConfirm.Open {
		t.Fatal("delete confirm should be closed")
	}
}

func key(value string) tea.KeyMsg {
	return tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune(value),
	}
}

func applyBatch(t *testing.T, m model, cmd tea.Cmd) model {
	t.Helper()

	msg := cmd()
	batch, ok := msg.(tea.BatchMsg)
	if !ok {
		next, _ := m.Update(msg)
		return next.(model)
	}

	updated := m
	for _, batchCmd := range batch {
		next, _ := updated.Update(batchCmd())
		updated = next.(model)
	}

	return updated
}
