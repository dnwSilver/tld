package storage

import (
	"context"
	"database/sql"
	"fmt"
)

const currentSchemaVersion = 5

func Migrate(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
		return fmt.Errorf("enable foreign keys: %w", err)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin migration: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if _, err := tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
		)
	`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	version, err := schemaVersion(ctx, tx)
	if err != nil {
		return err
	}

	if version < 1 {
		if err := migrateV1(ctx, tx); err != nil {
			return err
		}
	}
	if version < 2 {
		if err := migrateV2(ctx, tx); err != nil {
			return err
		}
	}
	if version < 3 {
		if err := migrateV3(ctx, tx); err != nil {
			return err
		}
	}
	if version < 4 {
		if err := migrateV4(ctx, tx); err != nil {
			return err
		}
	}
	if version < 5 {
		if err := migrateV5(ctx, tx); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration: %w", err)
	}

	return nil
}

func schemaVersion(ctx context.Context, tx *sql.Tx) (int, error) {
	var version sql.NullInt64
	if err := tx.QueryRowContext(ctx, "SELECT max(version) FROM schema_migrations").Scan(&version); err != nil {
		return 0, fmt.Errorf("read schema version: %w", err)
	}

	if !version.Valid {
		return 0, nil
	}

	return int(version.Int64), nil
}

func migrateV1(ctx context.Context, tx *sql.Tx) error {
	statements := []string{
		`
			CREATE TABLE IF NOT EXISTS app_meta (
				key TEXT PRIMARY KEY,
				value BLOB NOT NULL,
				created_at INTEGER NOT NULL DEFAULT (CAST(strftime('%s', 'now') AS INTEGER)),
				updated_at INTEGER NOT NULL DEFAULT (CAST(strftime('%s', 'now') AS INTEGER))
			)
		`,
		`
			CREATE TABLE IF NOT EXISTS kv_cache (
				namespace TEXT NOT NULL,
				key TEXT NOT NULL,
				value BLOB NOT NULL,
				content_type TEXT NOT NULL,
				expires_at INTEGER,
				created_at INTEGER NOT NULL DEFAULT (CAST(strftime('%s', 'now') AS INTEGER)),
				updated_at INTEGER NOT NULL DEFAULT (CAST(strftime('%s', 'now') AS INTEGER)),
				PRIMARY KEY (namespace, key)
			)
		`,
		`
			CREATE INDEX IF NOT EXISTS idx_kv_cache_expires_at
			ON kv_cache (expires_at)
			WHERE expires_at IS NOT NULL
		`,
		"INSERT OR IGNORE INTO schema_migrations (version) VALUES (1)",
	}

	for _, statement := range statements {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("apply schema v1: %w", err)
		}
	}

	return nil
}

func migrateV2(ctx context.Context, tx *sql.Tx) error {
	statements := []string{
		`
			CREATE TABLE IF NOT EXISTS stacks (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				icon TEXT NOT NULL,
				name TEXT NOT NULL,
				color TEXT NOT NULL,
				created_at INTEGER NOT NULL DEFAULT (CAST(strftime('%s', 'now') AS INTEGER)),
				updated_at INTEGER NOT NULL DEFAULT (CAST(strftime('%s', 'now') AS INTEGER)),
				UNIQUE (name)
			)
		`,
		"INSERT OR IGNORE INTO schema_migrations (version) VALUES (2)",
	}

	for _, statement := range statements {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("apply schema v2: %w", err)
		}
	}

	return nil
}

func migrateV3(ctx context.Context, tx *sql.Tx) error {
	statements := []string{
		`
			CREATE TABLE IF NOT EXISTS namespaces (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				icon TEXT NOT NULL,
				name TEXT NOT NULL,
				color TEXT NOT NULL,
				created_at INTEGER NOT NULL DEFAULT (CAST(strftime('%s', 'now') AS INTEGER)),
				updated_at INTEGER NOT NULL DEFAULT (CAST(strftime('%s', 'now') AS INTEGER)),
				UNIQUE (name)
			)
		`,
		"INSERT OR IGNORE INTO schema_migrations (version) VALUES (3)",
	}

	for _, statement := range statements {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("apply schema v3: %w", err)
		}
	}

	return nil
}

func migrateV4(ctx context.Context, tx *sql.Tx) error {
	statements := []string{
		`
			CREATE TABLE IF NOT EXISTS dependencies (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				stack_id INTEGER NOT NULL,
				icon TEXT NOT NULL,
				name TEXT NOT NULL,
				color TEXT NOT NULL,
				created_at INTEGER NOT NULL DEFAULT (CAST(strftime('%s', 'now') AS INTEGER)),
				updated_at INTEGER NOT NULL DEFAULT (CAST(strftime('%s', 'now') AS INTEGER)),
				UNIQUE (stack_id, name),
				FOREIGN KEY (stack_id) REFERENCES stacks (id) ON UPDATE CASCADE ON DELETE RESTRICT
			)
		`,
		`
			CREATE INDEX IF NOT EXISTS idx_dependencies_stack_id
			ON dependencies (stack_id)
		`,
		"INSERT OR IGNORE INTO schema_migrations (version) VALUES (4)",
	}

	for _, statement := range statements {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("apply schema v4: %w", err)
		}
	}

	return nil
}

func migrateV5(ctx context.Context, tx *sql.Tx) error {
	statements := []string{
		`
			CREATE TABLE IF NOT EXISTS sources (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				icon TEXT NOT NULL,
				name TEXT NOT NULL,
				color TEXT NOT NULL,
				pat_token TEXT NOT NULL,
				url TEXT NOT NULL,
				type TEXT NOT NULL,
				created_at INTEGER NOT NULL DEFAULT (CAST(strftime('%s', 'now') AS INTEGER)),
				updated_at INTEGER NOT NULL DEFAULT (CAST(strftime('%s', 'now') AS INTEGER)),
				UNIQUE (name),
				CHECK (type IN ('gitlab', 'github', 'gitea', 'bitbucket'))
			)
		`,
		"INSERT OR IGNORE INTO schema_migrations (version) VALUES (5)",
	}

	for _, statement := range statements {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("apply schema v5: %w", err)
		}
	}

	return nil
}
