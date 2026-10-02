package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

const (
	ProjectDependencyRunStatusRunning = "running"
	ProjectDependencyRunStatusSuccess = "success"
	ProjectDependencyRunStatusError   = "error"
)

type ProjectDependencyRepository struct {
	db *sql.DB
}

type ProjectDependencyRun struct {
	ID             int64
	ProjectID      int64
	CommitShortSHA string
	CommitSHA      string
	Status         string
	StartedAt      time.Time
	FinishedAt     *time.Time
	Error          string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type ProjectDependency struct {
	ID             int64
	ProjectID      int64
	RunID          int64
	Name           string
	Ecosystem      string
	Version        string
	DependencyType string
	SourceFile     string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (r ProjectDependencyRepository) LatestRun(ctx context.Context, projectID int64) (ProjectDependencyRun, error) {
	if projectID == 0 {
		return ProjectDependencyRun{}, errors.New("project dependency run project is empty")
	}

	var run ProjectDependencyRun
	var finishedAt sql.NullInt64
	var startedAt int64
	var createdAt int64
	var updatedAt int64
	err := r.db.QueryRowContext(ctx, `
		SELECT id, project_id, commit_short_sha, commit_sha, status, started_at, finished_at, error, created_at, updated_at
		FROM project_dependency_runs
		WHERE project_id = ?
		ORDER BY COALESCE(finished_at, started_at) DESC, id DESC
		LIMIT 1
	`, projectID).Scan(
		&run.ID,
		&run.ProjectID,
		&run.CommitShortSHA,
		&run.CommitSHA,
		&run.Status,
		&startedAt,
		&finishedAt,
		&run.Error,
		&createdAt,
		&updatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return ProjectDependencyRun{}, nil
	}
	if err != nil {
		return ProjectDependencyRun{}, fmt.Errorf("latest project dependency run: %w", err)
	}

	run.StartedAt = time.Unix(startedAt, 0).UTC()
	if finishedAt.Valid {
		value := time.Unix(finishedAt.Int64, 0).UTC()
		run.FinishedAt = &value
	}
	run.CreatedAt = time.Unix(createdAt, 0).UTC()
	run.UpdatedAt = time.Unix(updatedAt, 0).UTC()

	return run, nil
}

func (r ProjectDependencyRepository) ListByProject(ctx context.Context, projectID int64) ([]ProjectDependency, error) {
	if projectID == 0 {
		return []ProjectDependency{}, nil
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, project_id, run_id, name, ecosystem, version, dependency_type, source_file, created_at, updated_at
		FROM project_dependencies
		WHERE project_id = ?
		ORDER BY dependency_type ASC, name ASC
	`, projectID)
	if err != nil {
		return nil, fmt.Errorf("list project dependencies: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	dependencies := make([]ProjectDependency, 0)
	for rows.Next() {
		var dependency ProjectDependency
		var createdAt int64
		var updatedAt int64
		if err := rows.Scan(
			&dependency.ID,
			&dependency.ProjectID,
			&dependency.RunID,
			&dependency.Name,
			&dependency.Ecosystem,
			&dependency.Version,
			&dependency.DependencyType,
			&dependency.SourceFile,
			&createdAt,
			&updatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan project dependency: %w", err)
		}
		dependency.CreatedAt = time.Unix(createdAt, 0).UTC()
		dependency.UpdatedAt = time.Unix(updatedAt, 0).UTC()
		dependencies = append(dependencies, dependency)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate project dependencies: %w", err)
	}

	return dependencies, nil
}

func (r ProjectDependencyRepository) HasSuccessfulRun(ctx context.Context, projectID int64, shortSHA string) (bool, error) {
	if projectID == 0 {
		return false, errors.New("project dependency run project is empty")
	}
	if shortSHA == "" {
		return false, errors.New("project dependency run short sha is empty")
	}

	var count int
	if err := r.db.QueryRowContext(ctx, `
		SELECT count(*)
		FROM project_dependency_runs
		WHERE project_id = ?
			AND commit_short_sha = ?
			AND status = ?
	`, projectID, shortSHA, ProjectDependencyRunStatusSuccess).Scan(&count); err != nil {
		return false, fmt.Errorf("check project dependency run: %w", err)
	}

	return count > 0, nil
}

func (r ProjectDependencyRepository) ReplaceForProjectRun(ctx context.Context, projectID int64, run ProjectDependencyRun, dependencies []ProjectDependency) (ProjectDependencyRun, error) {
	if projectID == 0 {
		return ProjectDependencyRun{}, errors.New("project dependency project is empty")
	}
	if run.CommitShortSHA == "" {
		return ProjectDependencyRun{}, errors.New("project dependency run short sha is empty")
	}
	if run.CommitSHA == "" {
		return ProjectDependencyRun{}, errors.New("project dependency run sha is empty")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return ProjectDependencyRun{}, fmt.Errorf("begin project dependency replace: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	now := time.Now().Unix()
	startedAt := run.StartedAt.Unix()
	if startedAt <= 0 {
		startedAt = now
	}
	var finishedAt any
	if run.FinishedAt != nil {
		finishedAt = run.FinishedAt.Unix()
	} else if run.Status != ProjectDependencyRunStatusRunning {
		finishedAt = now
	}
	status := run.Status
	if status == "" {
		status = ProjectDependencyRunStatusSuccess
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO project_dependency_runs (
			project_id,
			commit_short_sha,
			commit_sha,
			status,
			started_at,
			finished_at,
			error,
			created_at,
			updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(project_id, commit_short_sha) DO UPDATE SET
			commit_sha = excluded.commit_sha,
			status = excluded.status,
			started_at = excluded.started_at,
			finished_at = excluded.finished_at,
			error = excluded.error,
			updated_at = excluded.updated_at
	`, projectID, run.CommitShortSHA, run.CommitSHA, status, startedAt, finishedAt, run.Error, now, now); err != nil {
		return ProjectDependencyRun{}, fmt.Errorf("upsert project dependency run: %w", err)
	}

	var runID int64
	if err := tx.QueryRowContext(ctx, `
		SELECT id
		FROM project_dependency_runs
		WHERE project_id = ? AND commit_short_sha = ?
	`, projectID, run.CommitShortSHA).Scan(&runID); err != nil {
		return ProjectDependencyRun{}, fmt.Errorf("read project dependency run id: %w", err)
	}

	if _, err := tx.ExecContext(ctx, "DELETE FROM project_dependencies WHERE project_id = ?", projectID); err != nil {
		return ProjectDependencyRun{}, fmt.Errorf("delete previous project dependencies: %w", err)
	}

	for _, dependency := range dependencies {
		if dependency.Name == "" {
			return ProjectDependencyRun{}, errors.New("project dependency name is empty")
		}
		if dependency.Version == "" {
			return ProjectDependencyRun{}, errors.New("project dependency version is empty")
		}
		if dependency.DependencyType == "" {
			return ProjectDependencyRun{}, errors.New("project dependency type is empty")
		}
		if dependency.SourceFile == "" {
			return ProjectDependencyRun{}, errors.New("project dependency source file is empty")
		}

		if _, err := tx.ExecContext(ctx, `
			INSERT INTO project_dependencies (
				project_id,
				run_id,
				name,
				ecosystem,
				version,
				dependency_type,
				source_file,
				created_at,
				updated_at
			)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, projectID, runID, dependency.Name, dependency.Ecosystem, dependency.Version, dependency.DependencyType, dependency.SourceFile, now, now); err != nil {
			return ProjectDependencyRun{}, fmt.Errorf("insert project dependency: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return ProjectDependencyRun{}, fmt.Errorf("commit project dependency replace: %w", err)
	}

	saved := run
	saved.ID = runID
	saved.ProjectID = projectID
	saved.Status = status
	saved.StartedAt = time.Unix(startedAt, 0).UTC()
	if finishedAtUnix, ok := finishedAt.(int64); ok {
		value := time.Unix(finishedAtUnix, 0).UTC()
		saved.FinishedAt = &value
	}
	saved.CreatedAt = time.Unix(now, 0).UTC()
	saved.UpdatedAt = time.Unix(now, 0).UTC()

	return saved, nil
}
