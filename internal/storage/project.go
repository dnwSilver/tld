package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type ProjectRepository struct {
	db *sql.DB
}

type Project struct {
	ID              int64
	ProjectID       string
	NamespaceID     int64
	NamespaceName   string
	SourceID        int64
	SourceName      string
	StackID         int64
	StackName       string
	StackIcon       string
	StackColor      string
	Icon            string
	Name            string
	Color           string
	Freezing        bool
	EndOfLife       bool
	DependencyCount int
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (r ProjectRepository) List(ctx context.Context) ([]Project, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT p.id, p.project_id, p.namespace_id, n.name, p.source_id, src.name, p.stack_id, st.name, st.icon, st.color, p.icon, p.name, p.color, p.freezing, p.endoflife, COALESCE(pd.dependency_count, 0), p.created_at, p.updated_at
		FROM projects p
		JOIN namespaces n ON n.id = p.namespace_id
		JOIN sources src ON src.id = p.source_id
		JOIN stacks st ON st.id = p.stack_id
		LEFT JOIN (
			SELECT project_id, count(*) AS dependency_count
			FROM project_dependencies
			GROUP BY project_id
		) pd ON pd.project_id = p.id
		ORDER BY n.name ASC, p.name ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	projects := make([]Project, 0)
	for rows.Next() {
		var project Project
		var createdAt int64
		var updatedAt int64
		var freezing int
		var endOfLife int
		if err := rows.Scan(
			&project.ID,
			&project.ProjectID,
			&project.NamespaceID,
			&project.NamespaceName,
			&project.SourceID,
			&project.SourceName,
			&project.StackID,
			&project.StackName,
			&project.StackIcon,
			&project.StackColor,
			&project.Icon,
			&project.Name,
			&project.Color,
			&freezing,
			&endOfLife,
			&project.DependencyCount,
			&createdAt,
			&updatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan project: %w", err)
		}
		project.Freezing = freezing != 0
		project.EndOfLife = endOfLife != 0
		project.CreatedAt = time.Unix(createdAt, 0).UTC()
		project.UpdatedAt = time.Unix(updatedAt, 0).UTC()
		projects = append(projects, project)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate projects: %w", err)
	}

	return projects, nil
}

func (r ProjectRepository) Count(ctx context.Context) (int, error) {
	var count int
	if err := r.db.QueryRowContext(ctx, "SELECT count(*) FROM projects").Scan(&count); err != nil {
		return 0, fmt.Errorf("count projects: %w", err)
	}

	return count, nil
}

func (r ProjectRepository) Create(ctx context.Context, projectID string, namespaceID int64, sourceID int64, stackID int64, icon string, name string, color string, freezing bool, endOfLife bool) (Project, error) {
	if projectID == "" {
		return Project{}, errors.New("project PROJECT_ID is empty")
	}
	if namespaceID == 0 {
		return Project{}, errors.New("project namespace is empty")
	}
	if sourceID == 0 {
		return Project{}, errors.New("project source is empty")
	}
	if stackID == 0 {
		return Project{}, errors.New("project stack is empty")
	}
	if icon == "" {
		return Project{}, errors.New("project icon is empty")
	}
	if name == "" {
		return Project{}, errors.New("project name is empty")
	}
	if color == "" {
		return Project{}, errors.New("project color is empty")
	}

	now := time.Now().Unix()
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO projects (project_id, namespace_id, source_id, stack_id, icon, name, color, freezing, endoflife, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, projectID, namespaceID, sourceID, stackID, icon, name, color, boolToInt(freezing), boolToInt(endOfLife), now, now)
	if err != nil {
		return Project{}, fmt.Errorf("create project: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return Project{}, fmt.Errorf("read project id: %w", err)
	}

	return Project{
		ID:          id,
		ProjectID:   projectID,
		NamespaceID: namespaceID,
		SourceID:    sourceID,
		StackID:     stackID,
		Icon:        icon,
		Name:        name,
		Color:       color,
		Freezing:    freezing,
		EndOfLife:   endOfLife,
		CreatedAt:   time.Unix(now, 0).UTC(),
		UpdatedAt:   time.Unix(now, 0).UTC(),
	}, nil
}

func (r ProjectRepository) Update(ctx context.Context, id int64, projectID string, namespaceID int64, sourceID int64, stackID int64, icon string, name string, color string, freezing bool, endOfLife bool) error {
	if id == 0 {
		return errors.New("project id is empty")
	}
	if projectID == "" {
		return errors.New("project PROJECT_ID is empty")
	}
	if namespaceID == 0 {
		return errors.New("project namespace is empty")
	}
	if sourceID == 0 {
		return errors.New("project source is empty")
	}
	if stackID == 0 {
		return errors.New("project stack is empty")
	}
	if icon == "" {
		return errors.New("project icon is empty")
	}
	if name == "" {
		return errors.New("project name is empty")
	}
	if color == "" {
		return errors.New("project color is empty")
	}

	result, err := r.db.ExecContext(ctx, `
		UPDATE projects
		SET project_id = ?, namespace_id = ?, source_id = ?, stack_id = ?, icon = ?, name = ?, color = ?, freezing = ?, endoflife = ?, updated_at = ?
		WHERE id = ?
	`, projectID, namespaceID, sourceID, stackID, icon, name, color, boolToInt(freezing), boolToInt(endOfLife), time.Now().Unix(), id)
	if err != nil {
		return fmt.Errorf("update project: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read updated projects count: %w", err)
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r ProjectRepository) Delete(ctx context.Context, id int64) error {
	if id == 0 {
		return errors.New("project id is empty")
	}

	result, err := r.db.ExecContext(ctx, "DELETE FROM projects WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete project: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read deleted projects count: %w", err)
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
