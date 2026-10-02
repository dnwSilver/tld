package storage

import (
	"context"
	"path/filepath"
	"testing"
)

func TestDependencyViewUsesProjectNamespacePolicy(t *testing.T) {
	ctx := context.Background()
	store, err := Open(ctx, filepath.Join(t.TempDir(), "view.db"), "secret")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	stack, err := store.Stacks().Create(ctx, "S", "Go", "#FFFFFF")
	if err != nil {
		t.Fatal(err)
	}
	source, err := store.Sources().Create(ctx, "GitHub", "secret", "https://github.com", SourceTypeGitHub)
	if err != nil {
		t.Fatal(err)
	}
	dependency, err := store.Dependencies().Create(ctx, stack.ID, "D", "module", "#FFFFFF")
	if err != nil {
		t.Fatal(err)
	}
	want := make(map[int64]string)
	for index, version := range []string{"1.0.0", "2.0.0"} {
		policy, err := store.Policies().Create(ctx, "Policy "+version)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := store.Policies().CreateValue(ctx, policy.ID, dependency.ID, version); err != nil {
			t.Fatal(err)
		}
		namespace, err := store.Namespaces().SaveWithPolicy(ctx, 0, "N", "Namespace "+version, "#FFFFFF", policy.ID)
		if err != nil {
			t.Fatal(err)
		}
		project, err := store.Projects().Create(ctx, "owner/repo"+string(rune('A'+index)), namespace.ID, source.ID, stack.ID, "P", "Project "+version, "#FFFFFF", false, false)
		if err != nil {
			t.Fatal(err)
		}
		want[project.ID] = version
	}
	view, err := store.ProjectDependencies().ViewByStack(ctx, stack.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(view.Columns) != 1 || view.Columns[0].PolicyVersion != "" {
		t.Fatalf("mixed policy columns: %#v", view.Columns)
	}
	for _, row := range view.Rows {
		if row.PolicyVersions[dependency.ID] != want[row.ProjectID] {
			t.Fatalf("project %d expected %q, got %q", row.ProjectID, want[row.ProjectID], row.PolicyVersions[dependency.ID])
		}
	}
}
