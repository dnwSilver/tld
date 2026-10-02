package storage

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

func TestDependencyViewDoesNotMergeDistinctPackageNames(t *testing.T) {
	ctx := context.Background()
	store, err := Open(ctx, filepath.Join(t.TempDir(), "identity.db"), "secret")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	stack, err := store.Stacks().Create(ctx, "S", "JavaScript", "#FFFFFF")
	if err != nil {
		t.Fatal(err)
	}
	policy, err := store.Policies().Create(ctx, "policy")
	if err != nil {
		t.Fatal(err)
	}
	namespace, err := store.Namespaces().SaveWithPolicy(ctx, 0, "N", "namespace", "#FFFFFF", policy.ID)
	if err != nil {
		t.Fatal(err)
	}
	source, err := store.Sources().Create(ctx, "GitHub", "secret", "https://github.com", SourceTypeGitHub)
	if err != nil {
		t.Fatal(err)
	}
	project, err := store.Projects().Create(ctx, "owner/repo", namespace.ID, source.ID, stack.ID, "P", "repo", "#FFFFFF", false, false)
	if err != nil {
		t.Fatal(err)
	}
	ids := make(map[string]int64)
	for _, name := range []string{"foo-bar", "foo_bar"} {
		dependency, err := store.Dependencies().CreateWithRegistryName(ctx, stack.ID, "D", name, "#FFFFFF", name)
		if err != nil {
			t.Fatal(err)
		}
		ids[name] = dependency.ID
		if _, err := store.Policies().CreateValue(ctx, policy.ID, dependency.ID, "2.0.0"); err != nil {
			t.Fatal(err)
		}
	}
	now := time.Now().UTC()
	_, err = store.ProjectDependencies().ReplaceForProjectRun(ctx, project.ID, ProjectDependencyRun{
		CommitSHA: "abcdef1234567890", CommitShortSHA: "abcdef12", Status: ProjectDependencyRunStatusSuccess,
		StartedAt: now, FinishedAt: &now,
	}, []ProjectDependency{{Name: "foo_bar", Version: "1.0.0", DependencyType: "dependencies", SourceFile: "package.json"}})
	if err != nil {
		t.Fatal(err)
	}
	view, err := store.ProjectDependencies().ViewByStack(ctx, stack.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(view.Rows) != 1 || view.Rows[0].Versions[ids["foo-bar"]] != "" || view.Rows[0].Versions[ids["foo_bar"]] != "1.0.0" {
		t.Fatalf("versions = %#v", view.Rows)
	}
	comparisons, err := store.ProjectDependencies().ListPolicyVersionComparisons(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(comparisons) != 1 {
		t.Fatalf("comparisons = %#v", comparisons)
	}
}

func TestDependencyViewSeparatesSameNameAcrossEcosystems(t *testing.T) {
	ctx := context.Background()
	store, err := Open(ctx, filepath.Join(t.TempDir(), "ecosystem.db"), "secret")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	stack, _ := store.Stacks().Create(ctx, "S", "Mixed", "#FFFFFF")
	policy, _ := store.Policies().Create(ctx, "policy")
	namespace, _ := store.Namespaces().SaveWithPolicy(ctx, 0, "N", "namespace", "#FFFFFF", policy.ID)
	projectSource, _ := store.Sources().Create(ctx, "GitHub", "secret", "https://github.com", SourceTypeGitHub)
	npm, _ := store.Sources().CreateWithRegistryKind(ctx, "NPM", "", "https://registry.npmjs.org", SourceTypeRegistry, "npm")
	maven, _ := store.Sources().CreateWithRegistryKind(ctx, "Maven", "", "https://repo.example", SourceTypeRegistry, "maven")
	project, _ := store.Projects().Create(ctx, "owner/repo", namespace.ID, projectSource.ID, stack.ID, "P", "repo", "#FFFFFF", false, false)
	npmDependency, _ := store.Dependencies().CreateWithRegistry(ctx, stack.ID, "D", "NPM shared", "#FFFFFF", "shared", npm.ID)
	mavenDependency, _ := store.Dependencies().CreateWithRegistry(ctx, stack.ID, "D", "Maven shared", "#FFFFFF", "shared", maven.ID)
	_, _ = store.Policies().CreateValue(ctx, policy.ID, npmDependency.ID, "2.0.0")
	_, _ = store.Policies().CreateValue(ctx, policy.ID, mavenDependency.ID, "3.0.0")
	now := time.Now().UTC()
	_, err = store.ProjectDependencies().ReplaceForProjectRun(ctx, project.ID, ProjectDependencyRun{
		CommitSHA: "abcdef1234567890", CommitShortSHA: "abcdef12", Status: ProjectDependencyRunStatusSuccess, StartedAt: now, FinishedAt: &now,
	}, []ProjectDependency{
		{Name: "shared", Ecosystem: "npm", Version: "1.0.0", DependencyType: "dependencies", SourceFile: "package.json"},
		{Name: "shared", Ecosystem: "maven", Version: "4.0.0", DependencyType: "library", SourceFile: "build.gradle"},
	})
	if err != nil {
		t.Fatal(err)
	}
	view, err := store.ProjectDependencies().ViewByStack(ctx, stack.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got := view.Rows[0].Versions[npmDependency.ID]; got != "1.0.0" {
		t.Fatalf("npm version = %q", got)
	}
	if got := view.Rows[0].Versions[mavenDependency.ID]; got != "4.0.0" {
		t.Fatalf("maven version = %q", got)
	}
}
