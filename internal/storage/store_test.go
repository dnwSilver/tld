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
