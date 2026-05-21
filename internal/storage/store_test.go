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
