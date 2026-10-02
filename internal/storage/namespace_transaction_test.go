package storage

import (
	"context"
	"path/filepath"
	"testing"
)

func TestSaveNamespaceWithPolicyRollsBackOnInvalidPolicy(t *testing.T) {
	ctx := context.Background()
	store, err := Open(ctx, filepath.Join(t.TempDir(), "store.db"), "secret")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	policy, err := store.Policies().Create(ctx, "Production")
	if err != nil {
		t.Fatal(err)
	}
	created, err := store.Namespaces().SaveWithPolicy(ctx, 0, "N", "Production", "#FFFFFF", policy.ID)
	if err != nil || created.PolicyID != policy.ID {
		t.Fatalf("create: %#v, %v", created, err)
	}
	if _, err := store.Namespaces().SaveWithPolicy(ctx, created.ID, "N", "Changed", "#FFFFFF", policy.ID+1000); err == nil {
		t.Fatal("invalid policy update succeeded")
	}
	items, err := store.Namespaces().List(ctx)
	if err != nil || len(items) != 1 || items[0].Name != "Production" || items[0].PolicyID != policy.ID {
		t.Fatalf("partial update remained: %#v, %v", items, err)
	}
}
