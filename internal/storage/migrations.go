package storage

import (
	"context"
	"database/sql"
	"fmt"
)

const currentSchemaVersion = 7

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
	if version < 6 {
		if err := migrateV6(ctx, tx); err != nil {
			return err
		}
	}
	if version < 7 {
		if err := migrateV7(ctx, tx); err != nil {
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

func migrateV6(ctx context.Context, tx *sql.Tx) error {
	statements := []string{
		`
			CREATE TABLE IF NOT EXISTS policies (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				name TEXT NOT NULL,
				created_at INTEGER NOT NULL DEFAULT (CAST(strftime('%s', 'now') AS INTEGER)),
				updated_at INTEGER NOT NULL DEFAULT (CAST(strftime('%s', 'now') AS INTEGER)),
				UNIQUE (name)
			)
		`,
		`
			ALTER TABLE namespaces
			ADD COLUMN policy_id INTEGER REFERENCES policies (id) ON UPDATE CASCADE ON DELETE SET NULL
		`,
		`
			CREATE TABLE IF NOT EXISTS policy_values (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				policy_id INTEGER NOT NULL,
				dependency_id INTEGER NOT NULL,
				version TEXT NOT NULL,
				created_at INTEGER NOT NULL DEFAULT (CAST(strftime('%s', 'now') AS INTEGER)),
				updated_at INTEGER NOT NULL DEFAULT (CAST(strftime('%s', 'now') AS INTEGER)),
				UNIQUE (policy_id, dependency_id),
				FOREIGN KEY (policy_id) REFERENCES policies (id) ON UPDATE CASCADE ON DELETE CASCADE,
				FOREIGN KEY (dependency_id) REFERENCES dependencies (id) ON UPDATE CASCADE ON DELETE RESTRICT
			)
		`,
		`
			CREATE INDEX IF NOT EXISTS idx_policy_values_policy_id
			ON policy_values (policy_id)
		`,
		`
			INSERT OR IGNORE INTO policies (name, created_at, updated_at)
			SELECT name, CAST(strftime('%s', 'now') AS INTEGER), CAST(strftime('%s', 'now') AS INTEGER)
			FROM namespaces
		`,
		`
			UPDATE namespaces
			SET policy_id = (
				SELECT policies.id
				FROM policies
				WHERE policies.name = namespaces.name
			)
			WHERE policy_id IS NULL
		`,
		"INSERT OR IGNORE INTO schema_migrations (version) VALUES (6)",
	}

	for _, statement := range statements {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("apply schema v6: %w", err)
		}
	}

	return nil
}

func migrateV7(ctx context.Context, tx *sql.Tx) error {
	hasNamespacePolicyID, err := hasColumn(ctx, tx, "namespaces", "policy_id")
	if err != nil {
		return err
	}
	if !hasNamespacePolicyID {
		if _, err := tx.ExecContext(ctx, "ALTER TABLE namespaces ADD COLUMN policy_id INTEGER"); err != nil {
			return fmt.Errorf("add namespace policy id: %w", err)
		}
	}

	hasPolicyNamespaceID, err := hasColumn(ctx, tx, "policies", "namespace_id")
	if err != nil {
		return err
	}
	if hasPolicyNamespaceID {
		statements := []string{
			`
				CREATE TABLE policies_v7 (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					name TEXT NOT NULL,
					created_at INTEGER NOT NULL DEFAULT (CAST(strftime('%s', 'now') AS INTEGER)),
					updated_at INTEGER NOT NULL DEFAULT (CAST(strftime('%s', 'now') AS INTEGER)),
					UNIQUE (name)
				)
			`,
			`
				INSERT OR IGNORE INTO policies_v7 (id, name, created_at, updated_at)
				SELECT id, name, created_at, updated_at
				FROM policies
			`,
			`
				UPDATE namespaces
				SET policy_id = (
					SELECT policies.id
					FROM policies
					WHERE policies.namespace_id = namespaces.id
				)
				WHERE policy_id IS NULL
			`,
			"ALTER TABLE policy_values RENAME TO policy_values_v6",
			"ALTER TABLE policies RENAME TO policies_v6",
			"ALTER TABLE policies_v7 RENAME TO policies",
			`
				CREATE TABLE policy_values (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					policy_id INTEGER NOT NULL,
					dependency_id INTEGER NOT NULL,
					version TEXT NOT NULL,
					created_at INTEGER NOT NULL DEFAULT (CAST(strftime('%s', 'now') AS INTEGER)),
					updated_at INTEGER NOT NULL DEFAULT (CAST(strftime('%s', 'now') AS INTEGER)),
					UNIQUE (policy_id, dependency_id),
					FOREIGN KEY (policy_id) REFERENCES policies (id) ON UPDATE CASCADE ON DELETE CASCADE,
					FOREIGN KEY (dependency_id) REFERENCES dependencies (id) ON UPDATE CASCADE ON DELETE RESTRICT
				)
			`,
			`
				INSERT OR IGNORE INTO policy_values (id, policy_id, dependency_id, version, created_at, updated_at)
				SELECT id, policy_id, dependency_id, version, created_at, updated_at
				FROM policy_values_v6
			`,
			`
				CREATE INDEX IF NOT EXISTS idx_policy_values_policy_id
				ON policy_values (policy_id)
			`,
			"DROP TABLE policy_values_v6",
			"DROP TABLE policies_v6",
		}

		for _, statement := range statements {
			if _, err := tx.ExecContext(ctx, statement); err != nil {
				return fmt.Errorf("apply schema v7: %w", err)
			}
		}
	}

	if _, err := tx.ExecContext(ctx, "INSERT OR IGNORE INTO schema_migrations (version) VALUES (7)"); err != nil {
		return fmt.Errorf("apply schema v7: %w", err)
	}

	return nil
}

func hasColumn(ctx context.Context, tx *sql.Tx, table string, column string) (bool, error) {
	rows, err := tx.QueryContext(ctx, "PRAGMA table_info("+table+")")
	if err != nil {
		return false, fmt.Errorf("read table info: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	for rows.Next() {
		var cid int
		var name string
		var columnType string
		var notNull int
		var defaultValue sql.NullString
		var primaryKey int
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			return false, fmt.Errorf("scan table info: %w", err)
		}
		if name == column {
			return true, nil
		}
	}
	if err := rows.Err(); err != nil {
		return false, fmt.Errorf("iterate table info: %w", err)
	}

	return false, nil
}
