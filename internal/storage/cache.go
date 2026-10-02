package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

const defaultContentType = "application/octet-stream"

var ErrCacheMiss = errors.New("cache entry not found")

type CacheRepository struct {
	db       *sql.DB
	maxBytes int64
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

// GetMany reads a small group of keys from one namespace in a single snapshot.
// Missing and expired entries are omitted from the result.
func (r CacheRepository) GetMany(ctx context.Context, namespace string, keys []string) (map[string]CacheEntry, error) {
	entries := make(map[string]CacheEntry, len(keys))
	if len(keys) == 0 {
		return entries, nil
	}
	const maxKeysPerQuery = 500
	for offset := 0; offset < len(keys); offset += maxKeysPerQuery {
		end := min(offset+maxKeysPerQuery, len(keys))
		batch := keys[offset:end]
		args := make([]any, 0, len(batch)+2)
		args = append(args, namespace)
		for _, key := range batch {
			args = append(args, key)
		}
		args = append(args, time.Now().Unix())
		query := `SELECT namespace, key, value, content_type, expires_at, created_at, updated_at
			FROM kv_cache WHERE namespace = ? AND key IN (` + strings.TrimSuffix(strings.Repeat("?,", len(batch)), ",") + `)
			AND (expires_at IS NULL OR expires_at > ?)`
		rows, err := r.db.QueryContext(ctx, query, args...)
		if err != nil {
			return nil, fmt.Errorf("read cache entries: %w", err)
		}
		for rows.Next() {
			var entry CacheEntry
			var expiresAt sql.NullInt64
			var createdAt, updatedAt int64
			if err := rows.Scan(&entry.Namespace, &entry.Key, &entry.Value, &entry.ContentType, &expiresAt, &createdAt, &updatedAt); err != nil {
				_ = rows.Close()
				return nil, fmt.Errorf("scan cache entry: %w", err)
			}
			if expiresAt.Valid {
				expires := time.Unix(expiresAt.Int64, 0).UTC()
				entry.ExpiresAt = &expires
			}
			entry.CreatedAt = time.Unix(createdAt, 0).UTC()
			entry.UpdatedAt = time.Unix(updatedAt, 0).UTC()
			entries[entry.Key] = entry
		}
		if err := rows.Err(); err != nil {
			_ = rows.Close()
			return nil, fmt.Errorf("iterate cache entries: %w", err)
		}
		if err := rows.Close(); err != nil {
			return nil, fmt.Errorf("close cache entries: %w", err)
		}
	}
	return entries, nil
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
	if r.maxBytes > 0 && int64(len(value)) > r.maxBytes {
		return fmt.Errorf("cache entry exceeds %d byte payload budget", r.maxBytes)
	}

	now := time.Now().Unix()
	var expiresAt any
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl).Unix()
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin cache write: %w", err)
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `
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
	if r.maxBytes > 0 {
		if err := enforceByteBudget(ctx, tx, r.maxBytes, namespace, key); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit cache write: %w", err)
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

// EnforceByteBudget removes the oldest cache entries until stored payloads fit
// in the budget. Metadata and SQLite page overhead are not included.
func (r CacheRepository) EnforceByteBudget(ctx context.Context, maxBytes int64) error {
	if maxBytes < 0 {
		return errors.New("cache byte budget is negative")
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin cache eviction: %w", err)
	}
	defer tx.Rollback()
	if err := enforceByteBudget(ctx, tx, maxBytes, "", ""); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit cache eviction: %w", err)
	}
	return nil
}

func enforceByteBudget(ctx context.Context, tx *sql.Tx, maxBytes int64, protectedNamespace, protectedKey string) error {
	var total int64
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(SUM(length(value)), 0) FROM kv_cache`).Scan(&total); err != nil {
		return fmt.Errorf("read cache payload size: %w", err)
	}
	if total <= maxBytes {
		return nil
	}
	rows, err := tx.QueryContext(ctx, `
		SELECT namespace, key, length(value)
		FROM kv_cache
		WHERE NOT (namespace = ? AND key = ?)
		ORDER BY updated_at ASC, created_at ASC, namespace ASC, key ASC
	`, protectedNamespace, protectedKey)
	if err != nil {
		return fmt.Errorf("list cache sizes: %w", err)
	}
	type item struct {
		namespace string
		key       string
		size      int64
	}
	items := make([]item, 0)
	for rows.Next() {
		var entry item
		if err := rows.Scan(&entry.namespace, &entry.key, &entry.size); err != nil {
			_ = rows.Close()
			return fmt.Errorf("scan cache size: %w", err)
		}
		items = append(items, entry)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return fmt.Errorf("iterate cache sizes: %w", err)
	}
	if err := rows.Close(); err != nil {
		return fmt.Errorf("close cache sizes: %w", err)
	}
	for _, entry := range items {
		if total <= maxBytes {
			break
		}
		if _, err := tx.ExecContext(ctx, "DELETE FROM kv_cache WHERE namespace = ? AND key = ?", entry.namespace, entry.key); err != nil {
			return fmt.Errorf("evict cache entry: %w", err)
		}
		total -= entry.size
	}
	if total > maxBytes {
		return errors.New("cache payload budget cannot be satisfied")
	}
	return nil
}

func (r CacheRepository) DeleteAmbiguousProjectKeys(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM kv_cache
		WHERE namespace IN ('project-files', 'project-checks', 'project-releases', 'project-release-status', 'project-vulnerabilities')
		AND key NOT LIKE 'v2:%'
	`)
	if err != nil {
		return fmt.Errorf("delete legacy project cache keys: %w", err)
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
