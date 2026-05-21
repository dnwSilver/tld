package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

const defaultContentType = "application/octet-stream"

var ErrCacheMiss = errors.New("cache entry not found")

type CacheRepository struct {
	db *sql.DB
}

type CacheEntry struct {
	Namespace   string
	Key         string
	Value       []byte
	ContentType string
	ExpiresAt   *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (r CacheRepository) Get(ctx context.Context, namespace string, key string) (CacheEntry, error) {
	now := time.Now().Unix()

	var entry CacheEntry
	var expiresAt sql.NullInt64
	var createdAt int64
	var updatedAt int64
	err := r.db.QueryRowContext(ctx, `
		SELECT namespace, key, value, content_type, expires_at, created_at, updated_at
		FROM kv_cache
		WHERE namespace = ?
			AND key = ?
			AND (expires_at IS NULL OR expires_at > ?)
	`, namespace, key, now).Scan(
		&entry.Namespace,
		&entry.Key,
		&entry.Value,
		&entry.ContentType,
		&expiresAt,
		&createdAt,
		&updatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return CacheEntry{}, ErrCacheMiss
	}
	if err != nil {
		return CacheEntry{}, fmt.Errorf("read cache entry: %w", err)
	}

	if expiresAt.Valid {
		expires := time.Unix(expiresAt.Int64, 0).UTC()
		entry.ExpiresAt = &expires
	}
	entry.CreatedAt = time.Unix(createdAt, 0).UTC()
	entry.UpdatedAt = time.Unix(updatedAt, 0).UTC()

	return entry, nil
}

func (r CacheRepository) Set(ctx context.Context, namespace string, key string, value []byte, contentType string, ttl time.Duration) error {
	if namespace == "" {
		return errors.New("cache namespace is empty")
	}
	if key == "" {
		return errors.New("cache key is empty")
	}
	if contentType == "" {
		contentType = defaultContentType
	}

	now := time.Now().Unix()
	var expiresAt any
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl).Unix()
	}

	if _, err := r.db.ExecContext(ctx, `
		INSERT INTO kv_cache (
			namespace,
			key,
			value,
			content_type,
			expires_at,
			created_at,
			updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(namespace, key) DO UPDATE SET
			value = excluded.value,
			content_type = excluded.content_type,
			expires_at = excluded.expires_at,
			updated_at = excluded.updated_at
	`, namespace, key, value, contentType, expiresAt, now, now); err != nil {
		return fmt.Errorf("write cache entry: %w", err)
	}

	return nil
}

func (r CacheRepository) Delete(ctx context.Context, namespace string, key string) error {
	if _, err := r.db.ExecContext(ctx, `
		DELETE FROM kv_cache
		WHERE namespace = ? AND key = ?
	`, namespace, key); err != nil {
		return fmt.Errorf("delete cache entry: %w", err)
	}

	return nil
}

func (r CacheRepository) DeleteExpired(ctx context.Context) error {
	if _, err := r.db.ExecContext(ctx, `
		DELETE FROM kv_cache
		WHERE expires_at IS NOT NULL AND expires_at <= ?
	`, time.Now().Unix()); err != nil {
		return fmt.Errorf("delete expired cache entries: %w", err)
	}

	return nil
}

func (r CacheRepository) ClearNamespace(ctx context.Context, namespace string) error {
	if _, err := r.db.ExecContext(ctx, `
		DELETE FROM kv_cache
		WHERE namespace = ?
	`, namespace); err != nil {
		return fmt.Errorf("clear cache namespace: %w", err)
	}

	return nil
}
