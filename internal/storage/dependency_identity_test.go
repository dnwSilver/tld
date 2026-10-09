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
	for _, name := range []string{"foo-bar", "foo_bar", "React", "spectrum/auth"} {
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
	}, []ProjectDependency{
		{Name: "foo_bar", Ecosystem: "npm", Version: "1.0.0", DependencyType: "dependencies", SourceFile: "package.json"},
		{Name: "react", Ecosystem: "npm", Version: "19.3.0", DependencyType: "dependencies", SourceFile: "package.json"},
		{Name: "@spectrum/auth", Ecosystem: "npm", Version: "7.1.0", DependencyType: "dependencies", SourceFile: "package.json"},
	})
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
	for name, want := range map[string]string{"React": "19.3.0", "spectrum/auth": "7.1.0"} {
		if got := view.Rows[0].Versions[ids[name]]; got != want {
			t.Fatalf("%s version = %q, want %q", name, got, want)
		}
	}
	comparisons, err := store.ProjectDependencies().ListPolicyVersionComparisons(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(comparisons) != 3 {
		t.Fatalf("comparisons = %#v", comparisons)
	}
	actual := make(map[string]string)
	for _, comparison := range comparisons {
		actual[comparison.Actual] = comparison.Policy
	}
	for _, version := range []string{"1.0.0", "19.3.0", "7.1.0"} {
		if actual[version] != "2.0.0" {
			t.Fatalf("missing policy comparison for %s: %#v", version, comparisons)
		}
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

func TestProjectVersionIndexPackageNames(t *testing.T) {
	index := newProjectVersionIndex()
	index.add("npm", "react", "19.3.0")
	index.add("npm", "@spectrum/auth", "7.1.0")
	index.add("npm", "foo_bar", "1.0.0")
	index.add("go", "github.com/Acme/module", "2.0.0")
	index.add("maven", "React", "3.0.0")
	index.add("node", "node", "24.18.0")
	for _, tt := range []struct{ ecosystem, name, want string }{
		{"npm", "React", "19.3.0"},
		{"npm", " spectrum/auth ", "7.1.0"},
		{"", "spectrum/auth", "7.1.0"},
		{"", "REACT", "19.3.0"},
		{"", "React", ""}, // Ambiguous between npm and Maven.
		{"npm", "foo-bar", ""},
		{"", "foo-bar", ""},
		{"go", "github.com/acme/module", ""},
		{"go", "github.com/Acme/module", "2.0.0"},
		{"maven", "react", ""},
		{"maven", "React", "3.0.0"},
		{"node", "Node.js", "24.18.0"},
		{"", "NODE", "24.18.0"},
	} {
		t.Run(tt.ecosystem+"/"+tt.name, func(t *testing.T) {
			if got := index.version(tt.ecosystem, tt.name); got != tt.want {
				t.Fatalf("version(%q, %q) = %q, want %q", tt.ecosystem, tt.name, got, tt.want)
			}
		})
	}
}
