package storage

import (
	"context"
	"database/sql"
	"fmt"
)

const currentSchemaVersion = 15

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
	if version < 8 {
		if err := migrateV8(ctx, tx); err != nil {
			return err
		}
	}
	if version < 9 {
		if err := migrateV9(ctx, tx); err != nil {
			return err
		}
	}
	if version < 10 {
		if err := migrateV10(ctx, tx); err != nil {
			return err
		}
	}
	if version < 11 {
		if err := migrateV11(ctx, tx); err != nil {
			return err
		}
	}
	if version < 12 {
		if err := migrateV12(ctx, tx); err != nil {
			return err
		}
	}
	if version < 13 {
		if err := migrateV13(ctx, tx); err != nil {
			return err
		}
	}
	if version < 14 {
		if err := migrateV14(ctx, tx); err != nil {
			return err
		}
	}
	if version < 15 {
		if err := migrateV15(ctx, tx); err != nil {
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

func migrateV8(ctx context.Context, tx *sql.Tx) error {
	statements := []string{
		`
			CREATE TABLE IF NOT EXISTS projects (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				namespace_id INTEGER NOT NULL,
				source_id INTEGER NOT NULL,
				stack_id INTEGER NOT NULL,
				icon TEXT NOT NULL,
				name TEXT NOT NULL,
				color TEXT NOT NULL,
				created_at INTEGER NOT NULL DEFAULT (CAST(strftime('%s', 'now') AS INTEGER)),
				updated_at INTEGER NOT NULL DEFAULT (CAST(strftime('%s', 'now') AS INTEGER)),
				UNIQUE (namespace_id, name),
				FOREIGN KEY (namespace_id) REFERENCES namespaces (id) ON UPDATE CASCADE ON DELETE RESTRICT,
				FOREIGN KEY (source_id) REFERENCES sources (id) ON UPDATE CASCADE ON DELETE RESTRICT,
				FOREIGN KEY (stack_id) REFERENCES stacks (id) ON UPDATE CASCADE ON DELETE RESTRICT
			)
		`,
		`
			CREATE INDEX IF NOT EXISTS idx_projects_namespace_id
			ON projects (namespace_id)
		`,
		`
			CREATE INDEX IF NOT EXISTS idx_projects_source_id
			ON projects (source_id)
		`,
		`
			CREATE INDEX IF NOT EXISTS idx_projects_stack_id
			ON projects (stack_id)
		`,
		"INSERT OR IGNORE INTO schema_migrations (version) VALUES (8)",
	}

	for _, statement := range statements {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("apply schema v8: %w", err)
		}
	}

	return nil
}

func migrateV9(ctx context.Context, tx *sql.Tx) error {
	hasProjectID, err := hasColumn(ctx, tx, "projects", "project_id")
	if err != nil {
		return err
	}
	if !hasProjectID {
		statements := []string{
			"ALTER TABLE projects ADD COLUMN project_id INTEGER NOT NULL DEFAULT 0",
			"UPDATE projects SET project_id = id WHERE project_id = 0",
		}

		for _, statement := range statements {
			if _, err := tx.ExecContext(ctx, statement); err != nil {
				return fmt.Errorf("apply schema v9: %w", err)
			}
		}
	}

	if _, err := tx.ExecContext(ctx, "INSERT OR IGNORE INTO schema_migrations (version) VALUES (9)"); err != nil {
		return fmt.Errorf("apply schema v9: %w", err)
	}

	return nil
}

func migrateV10(ctx context.Context, tx *sql.Tx) error {
	statements := []string{
		`
			CREATE TABLE projects_v10 (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				project_id TEXT NOT NULL,
				namespace_id INTEGER NOT NULL,
				source_id INTEGER NOT NULL,
				stack_id INTEGER NOT NULL,
				icon TEXT NOT NULL,
				name TEXT NOT NULL,
				color TEXT NOT NULL,
				created_at INTEGER NOT NULL DEFAULT (CAST(strftime('%s', 'now') AS INTEGER)),
				updated_at INTEGER NOT NULL DEFAULT (CAST(strftime('%s', 'now') AS INTEGER)),
				UNIQUE (namespace_id, name),
				FOREIGN KEY (namespace_id) REFERENCES namespaces (id) ON UPDATE CASCADE ON DELETE RESTRICT,
				FOREIGN KEY (source_id) REFERENCES sources (id) ON UPDATE CASCADE ON DELETE RESTRICT,
				FOREIGN KEY (stack_id) REFERENCES stacks (id) ON UPDATE CASCADE ON DELETE RESTRICT
			)
		`,
		`
			INSERT INTO projects_v10 (id, project_id, namespace_id, source_id, stack_id, icon, name, color, created_at, updated_at)
			SELECT id, CAST(project_id AS TEXT), namespace_id, source_id, stack_id, icon, name, color, created_at, updated_at
			FROM projects
		`,
		"DROP TABLE projects",
		"ALTER TABLE projects_v10 RENAME TO projects",
		`
			CREATE INDEX IF NOT EXISTS idx_projects_namespace_id
			ON projects (namespace_id)
		`,
		`
			CREATE INDEX IF NOT EXISTS idx_projects_source_id
			ON projects (source_id)
		`,
		`
			CREATE INDEX IF NOT EXISTS idx_projects_stack_id
			ON projects (stack_id)
		`,
		"INSERT OR IGNORE INTO schema_migrations (version) VALUES (10)",
	}

	for _, statement := range statements {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("apply schema v10: %w", err)
		}
	}

	return nil
}

func migrateV11(ctx context.Context, tx *sql.Tx) error {
	statements := []string{
		`
			CREATE TABLE IF NOT EXISTS project_dependency_runs (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				project_id INTEGER NOT NULL,
				commit_short_sha TEXT NOT NULL,
				commit_sha TEXT NOT NULL,
				status TEXT NOT NULL,
				started_at INTEGER NOT NULL,
				finished_at INTEGER,
				error TEXT NOT NULL DEFAULT '',
				created_at INTEGER NOT NULL DEFAULT (CAST(strftime('%s', 'now') AS INTEGER)),
				updated_at INTEGER NOT NULL DEFAULT (CAST(strftime('%s', 'now') AS INTEGER)),
				UNIQUE (project_id, commit_short_sha),
				FOREIGN KEY (project_id) REFERENCES projects (id) ON UPDATE CASCADE ON DELETE CASCADE
			)
		`,
		`
			CREATE TABLE IF NOT EXISTS project_dependencies (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				project_id INTEGER NOT NULL,
				run_id INTEGER NOT NULL,
				name TEXT NOT NULL,
				version TEXT NOT NULL,
				dependency_type TEXT NOT NULL,
				source_file TEXT NOT NULL,
				created_at INTEGER NOT NULL DEFAULT (CAST(strftime('%s', 'now') AS INTEGER)),
				updated_at INTEGER NOT NULL DEFAULT (CAST(strftime('%s', 'now') AS INTEGER)),
				FOREIGN KEY (project_id) REFERENCES projects (id) ON UPDATE CASCADE ON DELETE CASCADE,
				FOREIGN KEY (run_id) REFERENCES project_dependency_runs (id) ON UPDATE CASCADE ON DELETE CASCADE
			)
		`,
		`
			CREATE INDEX IF NOT EXISTS idx_project_dependency_runs_project_id
			ON project_dependency_runs (project_id)
		`,
		`
			CREATE INDEX IF NOT EXISTS idx_project_dependencies_project_id
			ON project_dependencies (project_id)
		`,
		`
			CREATE INDEX IF NOT EXISTS idx_project_dependencies_run_id
			ON project_dependencies (run_id)
		`,
		`
			CREATE INDEX IF NOT EXISTS idx_project_dependencies_project_name
			ON project_dependencies (project_id, name)
		`,
		"INSERT OR IGNORE INTO schema_migrations (version) VALUES (11)",
	}

	for _, statement := range statements {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("apply schema v11: %w", err)
		}
	}

	return nil
}

func migrateV12(ctx context.Context, tx *sql.Tx) error {
	hasFreezing, err := hasColumn(ctx, tx, "projects", "freezing")
	if err != nil {
		return err
	}
	if !hasFreezing {
		statements := []string{
			"ALTER TABLE projects ADD COLUMN freezing INTEGER NOT NULL DEFAULT 0",
			"ALTER TABLE projects ADD COLUMN endoflife INTEGER NOT NULL DEFAULT 0",
		}
		for _, statement := range statements {
			if _, err := tx.ExecContext(ctx, statement); err != nil {
				return fmt.Errorf("apply schema v12: %w", err)
			}
		}
	}

	if _, err := tx.ExecContext(ctx, "INSERT OR IGNORE INTO schema_migrations (version) VALUES (12)"); err != nil {
		return fmt.Errorf("apply schema v12: %w", err)
	}

	return nil
}

func migrateV13(ctx context.Context, tx *sql.Tx) error {
	statements := []string{
		"PRAGMA defer_foreign_keys = ON",
		"ALTER TABLE projects RENAME TO projects_v12",
		"ALTER TABLE sources RENAME TO sources_v12",
		`CREATE TABLE sources (id INTEGER PRIMARY KEY AUTOINCREMENT, icon TEXT NOT NULL, name TEXT NOT NULL, color TEXT NOT NULL, pat_token TEXT NOT NULL, url TEXT NOT NULL, type TEXT NOT NULL, registry_kind TEXT NOT NULL DEFAULT '', created_at INTEGER NOT NULL DEFAULT (CAST(strftime('%s', 'now') AS INTEGER)), updated_at INTEGER NOT NULL DEFAULT (CAST(strftime('%s', 'now') AS INTEGER)), UNIQUE (name), CHECK (type IN ('gitlab', 'github', 'gitea', 'bitbucket', 'registry')))`,
		`INSERT INTO sources (id, icon, name, color, pat_token, url, type, registry_kind, created_at, updated_at) SELECT id, icon, name, color, pat_token, url, type, '', created_at, updated_at FROM sources_v12`,
		`CREATE TABLE projects (id INTEGER PRIMARY KEY AUTOINCREMENT, project_id TEXT NOT NULL, namespace_id INTEGER NOT NULL, source_id INTEGER NOT NULL, stack_id INTEGER NOT NULL, icon TEXT NOT NULL, name TEXT NOT NULL, color TEXT NOT NULL, created_at INTEGER NOT NULL DEFAULT (CAST(strftime('%s', 'now') AS INTEGER)), updated_at INTEGER NOT NULL DEFAULT (CAST(strftime('%s', 'now') AS INTEGER)), freezing INTEGER NOT NULL DEFAULT 0, endoflife INTEGER NOT NULL DEFAULT 0, UNIQUE (namespace_id, name), FOREIGN KEY (namespace_id) REFERENCES namespaces (id) ON UPDATE CASCADE ON DELETE RESTRICT, FOREIGN KEY (source_id) REFERENCES sources (id) ON UPDATE CASCADE ON DELETE RESTRICT, FOREIGN KEY (stack_id) REFERENCES stacks (id) ON UPDATE CASCADE ON DELETE RESTRICT)`,
		`INSERT INTO projects (id, project_id, namespace_id, source_id, stack_id, icon, name, color, created_at, updated_at, freezing, endoflife) SELECT id, project_id, namespace_id, source_id, stack_id, icon, name, color, created_at, updated_at, freezing, endoflife FROM projects_v12`,
		`CREATE TABLE project_dependency_runs_v13 (id INTEGER PRIMARY KEY AUTOINCREMENT, project_id INTEGER NOT NULL, commit_short_sha TEXT NOT NULL, commit_sha TEXT NOT NULL, status TEXT NOT NULL, started_at INTEGER NOT NULL, finished_at INTEGER, error TEXT NOT NULL DEFAULT '', created_at INTEGER NOT NULL DEFAULT (CAST(strftime('%s', 'now') AS INTEGER)), updated_at INTEGER NOT NULL DEFAULT (CAST(strftime('%s', 'now') AS INTEGER)), UNIQUE (project_id, commit_short_sha), FOREIGN KEY (project_id) REFERENCES projects (id) ON UPDATE CASCADE ON DELETE CASCADE)`,
		`INSERT INTO project_dependency_runs_v13 SELECT id, project_id, commit_short_sha, commit_sha, status, started_at, finished_at, error, created_at, updated_at FROM project_dependency_runs`,
		`CREATE TABLE project_dependencies_v13 (id INTEGER PRIMARY KEY AUTOINCREMENT, project_id INTEGER NOT NULL, run_id INTEGER NOT NULL, name TEXT NOT NULL, version TEXT NOT NULL, dependency_type TEXT NOT NULL, source_file TEXT NOT NULL, created_at INTEGER NOT NULL DEFAULT (CAST(strftime('%s', 'now') AS INTEGER)), updated_at INTEGER NOT NULL DEFAULT (CAST(strftime('%s', 'now') AS INTEGER)), FOREIGN KEY (project_id) REFERENCES projects (id) ON UPDATE CASCADE ON DELETE CASCADE, FOREIGN KEY (run_id) REFERENCES project_dependency_runs_v13 (id) ON UPDATE CASCADE ON DELETE CASCADE)`,
		`INSERT INTO project_dependencies_v13 SELECT id, project_id, run_id, name, version, dependency_type, source_file, created_at, updated_at FROM project_dependencies`,
		"ALTER TABLE dependencies ADD COLUMN registry_name TEXT NOT NULL DEFAULT ''",
		"UPDATE dependencies SET registry_name = name WHERE registry_name = ''",
		"DROP TABLE project_dependencies",
		"DROP TABLE project_dependency_runs",
		"DROP TABLE projects_v12",
		"DROP TABLE sources_v12",
		"ALTER TABLE project_dependency_runs_v13 RENAME TO project_dependency_runs",
		"ALTER TABLE project_dependencies_v13 RENAME TO project_dependencies",
		"CREATE INDEX IF NOT EXISTS idx_projects_namespace_id ON projects (namespace_id)",
		"CREATE INDEX IF NOT EXISTS idx_projects_source_id ON projects (source_id)",
		"CREATE INDEX IF NOT EXISTS idx_projects_stack_id ON projects (stack_id)",
		"CREATE INDEX IF NOT EXISTS idx_project_dependency_runs_project_id ON project_dependency_runs (project_id)",
		"CREATE INDEX IF NOT EXISTS idx_project_dependencies_project_id ON project_dependencies (project_id)",
		"CREATE INDEX IF NOT EXISTS idx_project_dependencies_run_id ON project_dependencies (run_id)",
		"CREATE INDEX IF NOT EXISTS idx_project_dependencies_project_name ON project_dependencies (project_id, name)",
		"INSERT OR IGNORE INTO schema_migrations (version) VALUES (13)",
	}
	for _, statement := range statements {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("apply schema v13: %w", err)
		}
	}
	return nil
}

func migrateV14(ctx context.Context, tx *sql.Tx) error {
	statements := []string{
		"ALTER TABLE dependencies ADD COLUMN registry_source_id INTEGER NOT NULL DEFAULT 0",
		"INSERT OR IGNORE INTO schema_migrations (version) VALUES (14)",
	}
	for _, statement := range statements {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("apply schema v14: %w", err)
		}
	}
	return nil
}

func migrateV15(ctx context.Context, tx *sql.Tx) error {
	statements := []string{
		"ALTER TABLE project_dependencies ADD COLUMN ecosystem TEXT NOT NULL DEFAULT ''",
		`UPDATE project_dependencies SET ecosystem = CASE
			WHEN dependency_type IN ('dependencies', 'devDependencies', 'peerDependencies', 'optionalDependencies') THEN 'npm'
			WHEN dependency_type IN ('engines', 'nvmrc') THEN 'node'
			WHEN dependency_type = 'require' THEN 'go'
			WHEN dependency_type IN ('library', 'plugin') THEN 'maven'
			WHEN dependency_type = 'cocoapods' THEN 'cocoapods'
			WHEN dependency_type = 'bundler' THEN 'rubygems'
			ELSE '' END`,
		"CREATE INDEX IF NOT EXISTS idx_project_dependencies_identity ON project_dependencies (project_id, ecosystem, name)",
		"INSERT OR IGNORE INTO schema_migrations (version) VALUES (15)",
	}
	for _, statement := range statements {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("apply schema v15: %w", err)
		}
	}
	return nil
}
