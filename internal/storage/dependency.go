package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type DependencyRepository struct {
	db *sql.DB
}

type Dependency struct {
	ID                 int64
	StackID            int64
	StackName          string
	Icon               string
	Name               string
	Color              string
	RegistryName       string
	RegistrySourceID   int64
	RegistrySourceName string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

func (r DependencyRepository) List(ctx context.Context) ([]Dependency, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT d.id, d.stack_id, s.name, d.icon, d.name, d.color, d.registry_name, d.registry_source_id, COALESCE(src.name, ''), d.created_at, d.updated_at
		FROM dependencies d
		JOIN stacks s ON s.id = d.stack_id
		LEFT JOIN sources src ON src.id = d.registry_source_id AND src.type = 'registry'
		ORDER BY s.name ASC, d.name ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("list dependencies: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	dependencies := make([]Dependency, 0)
	for rows.Next() {
		var dependency Dependency
		var createdAt int64
		var updatedAt int64
		if err := rows.Scan(
			&dependency.ID,
			&dependency.StackID,
			&dependency.StackName,
			&dependency.Icon,
			&dependency.Name,
			&dependency.Color,
			&dependency.RegistryName,
			&dependency.RegistrySourceID,
			&dependency.RegistrySourceName,
			&createdAt,
			&updatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan dependency: %w", err)
		}
		dependency.CreatedAt = time.Unix(createdAt, 0).UTC()
		dependency.UpdatedAt = time.Unix(updatedAt, 0).UTC()
		dependencies = append(dependencies, dependency)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate dependencies: %w", err)
	}

	return dependencies, nil
}

func (r DependencyRepository) Count(ctx context.Context) (int, error) {
	var count int
	if err := r.db.QueryRowContext(ctx, "SELECT count(*) FROM dependencies").Scan(&count); err != nil {
		return 0, fmt.Errorf("count dependencies: %w", err)
	}

	return count, nil
}

func (r DependencyRepository) Create(ctx context.Context, stackID int64, icon string, name string, color string) (Dependency, error) {
	return r.CreateWithRegistryName(ctx, stackID, icon, name, color, name)
}

func (r DependencyRepository) CreateWithRegistryName(ctx context.Context, stackID int64, icon string, name string, color string, registryName string) (Dependency, error) {
	return r.CreateWithRegistry(ctx, stackID, icon, name, color, registryName, 0)
}

func (r DependencyRepository) CreateWithRegistry(ctx context.Context, stackID int64, icon string, name string, color string, registryName string, registrySourceID int64) (Dependency, error) {
	if stackID == 0 {
		return Dependency{}, errors.New("dependency stack is empty")
	}
	if icon == "" {
		return Dependency{}, errors.New("dependency icon is empty")
	}
	if name == "" {
		return Dependency{}, errors.New("dependency name is empty")
	}
	if color == "" {
		return Dependency{}, errors.New("dependency color is empty")
	}
	if registryName == "" {
		registryName = name
	}

	now := time.Now().Unix()
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO dependencies (stack_id, icon, name, color, registry_name, registry_source_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, stackID, icon, name, color, registryName, registrySourceID, now, now)
	if err != nil {
		return Dependency{}, fmt.Errorf("create dependency: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return Dependency{}, fmt.Errorf("read dependency id: %w", err)
	}

	return Dependency{
		ID:               id,
		StackID:          stackID,
		Icon:             icon,
		Name:             name,
		Color:            color,
		RegistryName:     registryName,
		RegistrySourceID: registrySourceID,
		CreatedAt:        time.Unix(now, 0).UTC(),
		UpdatedAt:        time.Unix(now, 0).UTC(),
	}, nil
}

func (r DependencyRepository) Update(ctx context.Context, id int64, stackID int64, icon string, name string, color string) error {
	return r.update(ctx, id, stackID, icon, name, color, "", 0, false)
}

func (r DependencyRepository) UpdateWithRegistryName(ctx context.Context, id int64, stackID int64, icon string, name string, color string, registryName string) error {
	return r.update(ctx, id, stackID, icon, name, color, registryName, 0, true)
}

func (r DependencyRepository) UpdateWithRegistry(ctx context.Context, id int64, stackID int64, icon string, name string, color string, registryName string, registrySourceID int64) error {
	return r.update(ctx, id, stackID, icon, name, color, registryName, registrySourceID, true)
}

func (r DependencyRepository) update(ctx context.Context, id int64, stackID int64, icon string, name string, color string, registryName string, registrySourceID int64, updateRegistry bool) error {
	if id == 0 {
		return errors.New("dependency id is empty")
	}
	if stackID == 0 {
		return errors.New("dependency stack is empty")
	}
	if icon == "" {
		return errors.New("dependency icon is empty")
	}
	if name == "" {
		return errors.New("dependency name is empty")
	}
	if color == "" {
		return errors.New("dependency color is empty")
	}

	query := `UPDATE dependencies SET stack_id = ?, icon = ?, name = ?, color = ?, updated_at = ? WHERE id = ?`
	args := []any{stackID, icon, name, color, time.Now().Unix(), id}
	if updateRegistry {
		if registryName == "" {
			registryName = name
		}
		query = `UPDATE dependencies SET stack_id = ?, icon = ?, name = ?, color = ?, registry_name = ?, registry_source_id = ?, updated_at = ? WHERE id = ?`
		args = []any{stackID, icon, name, color, registryName, registrySourceID, time.Now().Unix(), id}
	}
	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update dependency: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read updated dependencies count: %w", err)
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r DependencyRepository) Delete(ctx context.Context, id int64) error {
	if id == 0 {
		return errors.New("dependency id is empty")
	}

	result, err := r.db.ExecContext(ctx, "DELETE FROM dependencies WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete dependency: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read deleted dependencies count: %w", err)
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}
