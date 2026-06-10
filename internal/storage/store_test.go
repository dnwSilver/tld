package storage

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestOpenCreatesEncryptedDatabase(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "tld.db")

	store, err := Open(ctx, path, "secret")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer func() {
		_ = store.Close()
	}()

	if store.Path() != path {
		t.Fatalf("path = %q, want %q", store.Path(), path)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read db file: %v", err)
	}

	if bytes.Contains(raw, []byte("kv_cache")) {
		t.Fatal("database file contains plaintext schema")
	}
}

func TestOpenRejectsWrongPassphrase(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "tld.db")

	store, err := Open(ctx, path, "secret")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("close store: %v", err)
	}

	if _, err := Open(ctx, path, "wrong-secret"); err == nil {
		t.Fatal("expected wrong passphrase to fail")
	}
}

func TestCachePersistsAcrossOpen(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "tld.db")

	store, err := Open(ctx, path, "secret")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}

	if err := store.Cache().Set(ctx, "domain", "item", []byte(`{"ok":true}`), "application/json", 0); err != nil {
		t.Fatalf("set cache: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("close store: %v", err)
	}

	store, err = Open(ctx, path, "secret")
	if err != nil {
		t.Fatalf("reopen store: %v", err)
	}
	defer func() {
		_ = store.Close()
	}()

	entry, err := store.Cache().Get(ctx, "domain", "item")
	if err != nil {
		t.Fatalf("get cache: %v", err)
	}
	if string(entry.Value) != `{"ok":true}` {
		t.Fatalf("value = %q", string(entry.Value))
	}
	if entry.ContentType != "application/json" {
		t.Fatalf("content type = %q", entry.ContentType)
	}
}

func TestCacheTTL(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "tld.db")

	store, err := Open(ctx, path, "secret")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer func() {
		_ = store.Close()
	}()

	cache := store.Cache()
	if err := cache.Set(ctx, "domain", "short", []byte("value"), "", time.Millisecond); err != nil {
		t.Fatalf("set cache: %v", err)
	}

	time.Sleep(10 * time.Millisecond)

	if _, err := cache.Get(ctx, "domain", "short"); !errors.Is(err, ErrCacheMiss) {
		t.Fatalf("get expired cache error = %v, want %v", err, ErrCacheMiss)
	}

	if err := cache.DeleteExpired(ctx); err != nil {
		t.Fatalf("delete expired cache: %v", err)
	}
}

func TestStacks(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "tld.db")

	store, err := Open(ctx, path, "secret")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer func() {
		_ = store.Close()
	}()

	repository := store.Stacks()
	stack, err := repository.Create(ctx, "", "Git", "#84BA64")
	if err != nil {
		t.Fatalf("create stack: %v", err)
	}
	if stack.Icon != "" || stack.Name != "Git" || stack.Color != "#84BA64" {
		t.Fatalf("stack = %#v", stack)
	}

	count, err := repository.Count(ctx)
	if err != nil {
		t.Fatalf("count stacks: %v", err)
	}
	if count != 1 {
		t.Fatalf("count = %d, want 1", count)
	}

	stacks, err := repository.List(ctx)
	if err != nil {
		t.Fatalf("list stacks: %v", err)
	}
	if len(stacks) != 1 {
		t.Fatalf("len(stacks) = %d, want 1", len(stacks))
	}
	if stacks[0].Icon != "" || stacks[0].Name != "Git" || stacks[0].Color != "#84BA64" {
		t.Fatalf("stacks[0] = %#v", stacks[0])
	}

	if err := repository.Update(ctx, stack.ID, "", "Java", "#EC9706"); err != nil {
		t.Fatalf("update stack: %v", err)
	}

	stacks, err = repository.List(ctx)
	if err != nil {
		t.Fatalf("list stacks after update: %v", err)
	}
	if stacks[0].Icon != "" || stacks[0].Name != "Java" || stacks[0].Color != "#EC9706" {
		t.Fatalf("updated stack = %#v", stacks[0])
	}

	if err := repository.Delete(ctx, stack.ID); err != nil {
		t.Fatalf("delete stack: %v", err)
	}

	count, err = repository.Count(ctx)
	if err != nil {
		t.Fatalf("count stacks after delete: %v", err)
	}
	if count != 0 {
		t.Fatalf("count = %d, want 0", count)
	}
}

func TestNamespaces(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "tld.db")

	store, err := Open(ctx, path, "secret")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer func() {
		_ = store.Close()
	}()

	repository := store.Namespaces()
	namespace, err := repository.Create(ctx, "󱃾", "Production", "#25799F")
	if err != nil {
		t.Fatalf("create namespace: %v", err)
	}
	if namespace.Icon != "󱃾" || namespace.Name != "Production" || namespace.Color != "#25799F" {
		t.Fatalf("namespace = %#v", namespace)
	}

	count, err := repository.Count(ctx)
	if err != nil {
		t.Fatalf("count namespaces: %v", err)
	}
	if count != 1 {
		t.Fatalf("count = %d, want 1", count)
	}

	namespaces, err := repository.List(ctx)
	if err != nil {
		t.Fatalf("list namespaces: %v", err)
	}
	if len(namespaces) != 1 {
		t.Fatalf("len(namespaces) = %d, want 1", len(namespaces))
	}

	if err := repository.Update(ctx, namespace.ID, "󰒋", "Staging", "#EC9706"); err != nil {
		t.Fatalf("update namespace: %v", err)
	}

	namespaces, err = repository.List(ctx)
	if err != nil {
		t.Fatalf("list namespaces after update: %v", err)
	}
	if namespaces[0].Icon != "󰒋" || namespaces[0].Name != "Staging" || namespaces[0].Color != "#EC9706" {
		t.Fatalf("updated namespace = %#v", namespaces[0])
	}

	if err := repository.Delete(ctx, namespace.ID); err != nil {
		t.Fatalf("delete namespace: %v", err)
	}

	count, err = repository.Count(ctx)
	if err != nil {
		t.Fatalf("count namespaces after delete: %v", err)
	}
	if count != 0 {
		t.Fatalf("count = %d, want 0", count)
	}
}

func TestDependencies(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "tld.db")

	store, err := Open(ctx, path, "secret")
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

	repository := store.Dependencies()
	dependency, err := repository.Create(ctx, stack.ID, "", "Package", "#EC9706")
	if err != nil {
		t.Fatalf("create dependency: %v", err)
	}
	if dependency.StackID != stack.ID || dependency.Icon != "" || dependency.Name != "Package" || dependency.Color != "#EC9706" {
		t.Fatalf("dependency = %#v", dependency)
	}

	count, err := repository.Count(ctx)
	if err != nil {
		t.Fatalf("count dependencies: %v", err)
	}
	if count != 1 {
		t.Fatalf("count = %d, want 1", count)
	}

	dependencies, err := repository.List(ctx)
	if err != nil {
		t.Fatalf("list dependencies: %v", err)
	}
	if len(dependencies) != 1 {
		t.Fatalf("len(dependencies) = %d, want 1", len(dependencies))
	}
	if dependencies[0].StackName != "Git" {
		t.Fatalf("dependency stack name = %q, want Git", dependencies[0].StackName)
	}

	if err := repository.Update(ctx, dependency.ID, stack.ID, "󰎙", "Runtime", "#25799F"); err != nil {
		t.Fatalf("update dependency: %v", err)
	}

	dependencies, err = repository.List(ctx)
	if err != nil {
		t.Fatalf("list dependencies after update: %v", err)
	}
	if dependencies[0].Icon != "󰎙" || dependencies[0].Name != "Runtime" || dependencies[0].Color != "#25799F" {
		t.Fatalf("updated dependency = %#v", dependencies[0])
	}

	if err := repository.Delete(ctx, dependency.ID); err != nil {
		t.Fatalf("delete dependency: %v", err)
	}

	count, err = repository.Count(ctx)
	if err != nil {
		t.Fatalf("count dependencies after delete: %v", err)
	}
	if count != 0 {
		t.Fatalf("count = %d, want 0", count)
	}
}

func TestSources(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "tld.db")

	store, err := Open(ctx, path, "secret")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer func() {
		_ = store.Close()
	}()

	repository := store.Sources()
	source, err := repository.Create(ctx, "GitLab", "glpat-secret", "https://gitlab.com", SourceTypeGitLab)
	if err != nil {
		t.Fatalf("create source: %v", err)
	}
	if source.Icon != "" || source.Color != "FC6D26" || source.PATToken != "glpat-secret" || source.Type != SourceTypeGitLab {
		t.Fatalf("source = %#v", source)
	}

	count, err := repository.Count(ctx)
	if err != nil {
		t.Fatalf("count sources: %v", err)
	}
	if count != 1 {
		t.Fatalf("count = %d, want 1", count)
	}

	sources, err := repository.List(ctx)
	if err != nil {
		t.Fatalf("list sources: %v", err)
	}
	if len(sources) != 1 {
		t.Fatalf("len(sources) = %d, want 1", len(sources))
	}
	if sources[0].Name != "GitLab" || sources[0].URL != "https://gitlab.com" {
		t.Fatalf("sources[0] = %#v", sources[0])
	}

	if err := repository.Update(ctx, source.ID, "GitHub", "ghp-secret", "https://github.com", SourceTypeGitHub); err != nil {
		t.Fatalf("update source: %v", err)
	}

	sources, err = repository.List(ctx)
	if err != nil {
		t.Fatalf("list sources after update: %v", err)
	}
	if sources[0].Icon != "" || sources[0].Color != "FFFFFF" || sources[0].PATToken != "ghp-secret" || sources[0].Type != SourceTypeGitHub {
		t.Fatalf("updated source = %#v", sources[0])
	}

	if err := repository.Delete(ctx, source.ID); err != nil {
		t.Fatalf("delete source: %v", err)
	}

	count, err = repository.Count(ctx)
	if err != nil {
		t.Fatalf("count sources after delete: %v", err)
	}
	if count != 0 {
		t.Fatalf("count = %d, want 0", count)
	}
}

func TestProjects(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "tld.db")

	store, err := Open(ctx, path, "secret")
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
	source, err := store.Sources().Create(ctx, "GitLab", "glpat-secret", "https://gitlab.com", SourceTypeGitLab)
	if err != nil {
		t.Fatalf("create source: %v", err)
	}

	repository := store.Projects()
	project, err := repository.Create(ctx, "autobase-main", namespace.ID, source.ID, stack.ID, "󰏖", "TLD", "#EC9706", false, false)
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	if project.ProjectID != "autobase-main" || project.NamespaceID != namespace.ID || project.SourceID != source.ID || project.StackID != stack.ID {
		t.Fatalf("project = %#v", project)
	}

	count, err := repository.Count(ctx)
	if err != nil {
		t.Fatalf("count projects: %v", err)
	}
	if count != 1 {
		t.Fatalf("count = %d, want 1", count)
	}

	projects, err := repository.List(ctx)
	if err != nil {
		t.Fatalf("list projects: %v", err)
	}
	if len(projects) != 1 {
		t.Fatalf("len(projects) = %d, want 1", len(projects))
	}
	if projects[0].NamespaceName != "Production" || projects[0].SourceName != "GitLab" || projects[0].StackName != "Git" {
		t.Fatalf("projects[0] = %#v", projects[0])
	}
	if projects[0].DependencyCount != 0 {
		t.Fatalf("dependency count = %d, want 0", projects[0].DependencyCount)
	}

	now := time.Now().UTC()
	if _, err := store.ProjectDependencies().ReplaceForProjectRun(ctx, project.ID, ProjectDependencyRun{
		CommitShortSHA: "abcdef12",
		CommitSHA:      "abcdef1234567890",
		Status:         ProjectDependencyRunStatusSuccess,
		StartedAt:      now,
		FinishedAt:     &now,
	}, []ProjectDependency{
		{Name: "react", Version: "19.2.0", DependencyType: "dependencies", SourceFile: "package.json"},
		{Name: "typescript", Version: "5.9.3", DependencyType: "devDependencies", SourceFile: "package.json"},
	}); err != nil {
		t.Fatalf("replace project dependencies: %v", err)
	}

	projects, err = repository.List(ctx)
	if err != nil {
		t.Fatalf("list projects with dependencies: %v", err)
	}
	if projects[0].DependencyCount != 2 {
		t.Fatalf("dependency count = %d, want 2", projects[0].DependencyCount)
	}

	if err := repository.Update(ctx, project.ID, "autobase-ui", namespace.ID, source.ID, stack.ID, "󰏖", "TLD UI", "#84BA64", false, false); err != nil {
		t.Fatalf("update project: %v", err)
	}

	projects, err = repository.List(ctx)
	if err != nil {
		t.Fatalf("list projects after update: %v", err)
	}
	if projects[0].ProjectID != "autobase-ui" || projects[0].Name != "TLD UI" || projects[0].Color != "#84BA64" {
		t.Fatalf("updated project = %#v", projects[0])
	}

	if err := repository.Delete(ctx, project.ID); err != nil {
		t.Fatalf("delete project: %v", err)
	}

	count, err = repository.Count(ctx)
	if err != nil {
		t.Fatalf("count projects after delete: %v", err)
	}
	if count != 0 {
		t.Fatalf("count = %d, want 0", count)
	}
}

func TestProjectDependenciesReplaceAndSkip(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "tld.db")

	store, err := Open(ctx, path, "secret")
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
	stack, err := store.Stacks().Create(ctx, "", "JavaScript", "#84BA64")
	if err != nil {
		t.Fatalf("create stack: %v", err)
	}
	source, err := store.Sources().Create(ctx, "GitHub", "ghp-secret", "https://github.com", SourceTypeGitHub)
	if err != nil {
		t.Fatalf("create source: %v", err)
	}
	project, err := store.Projects().Create(ctx, "owner/repo", namespace.ID, source.ID, stack.ID, "󰏖", "TLD", "#EC9706", false, false)
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	now := time.Now().UTC()
	repository := store.ProjectDependencies()
	run, err := repository.ReplaceForProjectRun(ctx, project.ID, ProjectDependencyRun{
		CommitShortSHA: "abcdef12",
		CommitSHA:      "abcdef1234567890",
		Status:         ProjectDependencyRunStatusSuccess,
		StartedAt:      now,
		FinishedAt:     &now,
	}, []ProjectDependency{
		{Name: "react", Version: "^19.0.0", DependencyType: "dependencies", SourceFile: "package.json"},
		{Name: "typescript", Version: "^5.0.0", DependencyType: "devDependencies", SourceFile: "package.json"},
	})
	if err != nil {
		t.Fatalf("replace project dependencies: %v", err)
	}
	if run.ID == 0 {
		t.Fatal("run id should be set")
	}

	hasRun, err := repository.HasSuccessfulRun(ctx, project.ID, "abcdef12")
	if err != nil {
		t.Fatalf("has successful run: %v", err)
	}
	if !hasRun {
		t.Fatal("expected successful run")
	}

	dependencies, err := repository.ListByProject(ctx, project.ID)
	if err != nil {
		t.Fatalf("list project dependencies: %v", err)
	}
	if len(dependencies) != 2 {
		t.Fatalf("len(dependencies) = %d, want 2", len(dependencies))
	}

	_, err = repository.ReplaceForProjectRun(ctx, project.ID, ProjectDependencyRun{
		CommitShortSHA: "abcdef12",
		CommitSHA:      "abcdef1234567890",
		Status:         ProjectDependencyRunStatusSuccess,
		StartedAt:      now,
		FinishedAt:     &now,
	}, []ProjectDependency{
		{Name: "react", Version: "^19.1.0", DependencyType: "dependencies", SourceFile: "package.json"},
	})
	if err != nil {
		t.Fatalf("replace project dependencies again: %v", err)
	}

	dependencies, err = repository.ListByProject(ctx, project.ID)
	if err != nil {
		t.Fatalf("list replaced project dependencies: %v", err)
	}
	if len(dependencies) != 1 {
		t.Fatalf("len(dependencies) = %d, want 1", len(dependencies))
	}
	if dependencies[0].Version != "^19.1.0" {
		t.Fatalf("dependency version = %q, want ^19.1.0", dependencies[0].Version)
	}

	latestRun, err := repository.LatestRun(ctx, project.ID)
	if err != nil {
		t.Fatalf("latest run: %v", err)
	}
	if latestRun.CommitShortSHA != "abcdef12" || latestRun.Status != ProjectDependencyRunStatusSuccess {
		t.Fatalf("latest run = %#v", latestRun)
	}
}

func TestDependencyViewByStack(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "tld.db")

	store, err := Open(ctx, path, "secret")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer func() {
		_ = store.Close()
	}()

	stack, err := store.Stacks().Create(ctx, "", "JavaScript", "#84BA64")
	if err != nil {
		t.Fatalf("create stack: %v", err)
	}
	namespace, err := store.Namespaces().Create(ctx, "󱃾", "Production", "#25799F")
	if err != nil {
		t.Fatalf("create namespace: %v", err)
	}
	policy, err := store.Policies().Create(ctx, "Production policy")
	if err != nil {
		t.Fatalf("create policy: %v", err)
	}
	if err := store.Namespaces().SetPolicy(ctx, namespace.ID, policy.ID); err != nil {
		t.Fatalf("set policy: %v", err)
	}
	source, err := store.Sources().Create(ctx, "GitHub", "ghp-secret", "https://github.com", SourceTypeGitHub)
	if err != nil {
		t.Fatalf("create source: %v", err)
	}
	project, err := store.Projects().Create(ctx, "owner/repo", namespace.ID, source.ID, stack.ID, "󰏖", "TLD", "#EC9706", false, false)
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	dependency, err := store.Dependencies().Create(ctx, stack.ID, "", "React Native", "#61DAFB")
	if err != nil {
		t.Fatalf("create dependency: %v", err)
	}
	if _, err := store.Policies().CreateValue(ctx, policy.ID, dependency.ID, "19.2.0"); err != nil {
		t.Fatalf("create policy value: %v", err)
	}

	now := time.Now().UTC()
	if _, err := store.ProjectDependencies().ReplaceForProjectRun(ctx, project.ID, ProjectDependencyRun{
		CommitShortSHA: "abcdef12",
		CommitSHA:      "abcdef1234567890",
		Status:         ProjectDependencyRunStatusSuccess,
		StartedAt:      now,
		FinishedAt:     &now,
	}, []ProjectDependency{
		{Name: "react-native", Version: "0.82.1", DependencyType: "dependencies", SourceFile: "package.json"},
	}); err != nil {
		t.Fatalf("replace project dependencies: %v", err)
	}

	view, err := store.ProjectDependencies().ViewByStack(ctx, stack.ID)
	if err != nil {
		t.Fatalf("view by stack: %v", err)
	}
	if view.StackName != "JavaScript" {
		t.Fatalf("stack name = %q, want JavaScript", view.StackName)
	}
	if len(view.Columns) != 1 || view.Columns[0].Name != "React Native" {
		t.Fatalf("columns = %#v", view.Columns)
	}
	if view.Columns[0].PolicyVersion != "19.2.0" {
		t.Fatalf("policy version = %q, want 19.2.0", view.Columns[0].PolicyVersion)
	}
	if len(view.Rows) != 1 || view.Rows[0].ProjectName != "TLD" {
		t.Fatalf("rows = %#v", view.Rows)
	}
	if view.Rows[0].ProjectIcon != "󰏖" || view.Rows[0].ProjectColor != "#EC9706" {
		t.Fatalf("row project appearance = %#v", view.Rows[0])
	}
	if view.Rows[0].Versions[dependency.ID] != "0.82.1" {
		t.Fatalf("version = %q, want 0.82.1", view.Rows[0].Versions[dependency.ID])
	}
}

func TestDependencyViewByStackMatchesScopedJavaScriptPackages(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "tld.db")

	store, err := Open(ctx, path, "secret")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer func() {
		_ = store.Close()
	}()

	stack, err := store.Stacks().Create(ctx, "", "JavaScript", "#84BA64")
	if err != nil {
		t.Fatalf("create stack: %v", err)
	}
	namespace, err := store.Namespaces().Create(ctx, "󱃾", "Production", "#25799F")
	if err != nil {
		t.Fatalf("create namespace: %v", err)
	}
	policy, err := store.Policies().Create(ctx, "Production policy")
	if err != nil {
		t.Fatalf("create policy: %v", err)
	}
	if err := store.Namespaces().SetPolicy(ctx, namespace.ID, policy.ID); err != nil {
		t.Fatalf("set policy: %v", err)
	}
	source, err := store.Sources().Create(ctx, "GitHub", "ghp-secret", "https://github.com", SourceTypeGitHub)
	if err != nil {
		t.Fatalf("create source: %v", err)
	}
	project, err := store.Projects().Create(ctx, "owner/repo", namespace.ID, source.ID, stack.ID, "󰏖", "TLD", "#EC9706", false, false)
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	dependency, err := store.Dependencies().Create(ctx, stack.ID, "", "Spectrum UI Kit", "#84BA64")
	if err != nil {
		t.Fatalf("create dependency: %v", err)
	}
	if _, err := store.Policies().CreateValue(ctx, policy.ID, dependency.ID, "1.0.0"); err != nil {
		t.Fatalf("create policy value: %v", err)
	}

	now := time.Now().UTC()
	if _, err := store.ProjectDependencies().ReplaceForProjectRun(ctx, project.ID, ProjectDependencyRun{
		CommitShortSHA: "abcdef12",
		CommitSHA:      "abcdef1234567890",
		Status:         ProjectDependencyRunStatusSuccess,
		StartedAt:      now,
		FinishedAt:     &now,
	}, []ProjectDependency{
		{Name: "@spectrum/ui-kit", Version: "1.2.3", DependencyType: "dependencies", SourceFile: "package.json"},
	}); err != nil {
		t.Fatalf("replace project dependencies: %v", err)
	}

	view, err := store.ProjectDependencies().ViewByStack(ctx, stack.ID)
	if err != nil {
		t.Fatalf("view by stack: %v", err)
	}
	if view.Rows[0].Versions[dependency.ID] != "1.2.3" {
		t.Fatalf("version = %q, want 1.2.3", view.Rows[0].Versions[dependency.ID])
	}
}

func TestDependencyViewByStackMatchesNodeFromNvmrc(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "tld.db")

	store, err := Open(ctx, path, "secret")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer func() {
		_ = store.Close()
	}()

	stack, err := store.Stacks().Create(ctx, "", "JavaScript", "#84BA64")
	if err != nil {
		t.Fatalf("create stack: %v", err)
	}
	namespace, err := store.Namespaces().Create(ctx, "󱃾", "Production", "#25799F")
	if err != nil {
		t.Fatalf("create namespace: %v", err)
	}
	policy, err := store.Policies().Create(ctx, "Production policy")
	if err != nil {
		t.Fatalf("create policy: %v", err)
	}
	if err := store.Namespaces().SetPolicy(ctx, namespace.ID, policy.ID); err != nil {
		t.Fatalf("set policy: %v", err)
	}
	source, err := store.Sources().Create(ctx, "GitHub", "ghp-secret", "https://github.com", SourceTypeGitHub)
	if err != nil {
		t.Fatalf("create source: %v", err)
	}
	project, err := store.Projects().Create(ctx, "owner/repo", namespace.ID, source.ID, stack.ID, "󰏖", "TLD", "#EC9706", false, false)
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	dependency, err := store.Dependencies().Create(ctx, stack.ID, "", "Node.js", "#339933")
	if err != nil {
		t.Fatalf("create dependency: %v", err)
	}
	if _, err := store.Policies().CreateValue(ctx, policy.ID, dependency.ID, "24.11.1"); err != nil {
		t.Fatalf("create policy value: %v", err)
	}

	now := time.Now().UTC()
	if _, err := store.ProjectDependencies().ReplaceForProjectRun(ctx, project.ID, ProjectDependencyRun{
		CommitShortSHA: "abcdef12",
		CommitSHA:      "abcdef1234567890",
		Status:         ProjectDependencyRunStatusSuccess,
		StartedAt:      now,
		FinishedAt:     &now,
	}, []ProjectDependency{
		{Name: "node", Version: "24.11.1", DependencyType: "nvmrc", SourceFile: ".nvmrc"},
	}); err != nil {
		t.Fatalf("replace project dependencies: %v", err)
	}

	view, err := store.ProjectDependencies().ViewByStack(ctx, stack.ID)
	if err != nil {
		t.Fatalf("view by stack: %v", err)
	}
	if view.Rows[0].Versions[dependency.ID] != "24.11.1" {
		t.Fatalf("version = %q, want 24.11.1", view.Rows[0].Versions[dependency.ID])
	}
}

func TestPolicies(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "tld.db")

	store, err := Open(ctx, path, "secret")
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
	dependency, err := store.Dependencies().Create(ctx, stack.ID, "", "Package", "#EC9706")
	if err != nil {
		t.Fatalf("create dependency: %v", err)
	}

	repository := store.Policies()
	policy, err := repository.Create(ctx, "Production policy")
	if err != nil {
		t.Fatalf("create policy: %v", err)
	}
	if policy.Name != "Production policy" {
		t.Fatalf("policy = %#v", policy)
	}
	if _, err := repository.AssignToNamespace(ctx, namespace.ID, policy.Name); err != nil {
		t.Fatalf("assign policy: %v", err)
	}

	policies, err := repository.List(ctx)
	if err != nil {
		t.Fatalf("list policies: %v", err)
	}
	if len(policies) != 1 || policies[0].DependencyCount != 0 {
		t.Fatalf("policies = %#v", policies)
	}

	value, err := repository.CreateValue(ctx, policy.ID, dependency.ID, "1.2.3")
	if err != nil {
		t.Fatalf("create policy value: %v", err)
	}
	if value.PolicyID != policy.ID || value.DependencyID != dependency.ID || value.Version != "1.2.3" {
		t.Fatalf("policy value = %#v", value)
	}
	policies, err = repository.List(ctx)
	if err != nil {
		t.Fatalf("list policies after value: %v", err)
	}
	if len(policies) != 1 || policies[0].DependencyCount != 1 {
		t.Fatalf("policies after value = %#v", policies)
	}

	values, err := repository.ListValues(ctx, policy.ID)
	if err != nil {
		t.Fatalf("list policy values: %v", err)
	}
	if len(values) != 1 || values[0].DependencyName != "Package" || values[0].Version != "1.2.3" {
		t.Fatalf("policy values = %#v", values)
	}

	if err := repository.UpdateValue(ctx, value.ID, dependency.ID, "2.0.0"); err != nil {
		t.Fatalf("update policy value: %v", err)
	}
	values, err = repository.ListValues(ctx, policy.ID)
	if err != nil {
		t.Fatalf("list policy values after update: %v", err)
	}
	if values[0].Version != "2.0.0" {
		t.Fatalf("updated policy value = %#v", values[0])
	}

	if err := repository.DeleteValue(ctx, value.ID); err != nil {
		t.Fatalf("delete policy value: %v", err)
	}
	if err := repository.Delete(ctx, policy.ID); err != nil {
		t.Fatalf("delete policy: %v", err)
	}
}
