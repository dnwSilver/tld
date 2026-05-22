package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type NamespaceRepository struct {
	db *sql.DB
}

type Namespace struct {
	ID        int64
	Icon      string
	Name      string
	Color     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (r NamespaceRepository) List(ctx context.Context) ([]Namespace, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, icon, name, color, created_at, updated_at
		FROM namespaces
		ORDER BY name ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("list namespaces: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	namespaces := make([]Namespace, 0)
	for rows.Next() {
		var namespace Namespace
		var createdAt int64
		var updatedAt int64
		if err := rows.Scan(&namespace.ID, &namespace.Icon, &namespace.Name, &namespace.Color, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("scan namespace: %w", err)
		}
		namespace.CreatedAt = time.Unix(createdAt, 0).UTC()
		namespace.UpdatedAt = time.Unix(updatedAt, 0).UTC()
		namespaces = append(namespaces, namespace)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate namespaces: %w", err)
	}

	return namespaces, nil
}

func (r NamespaceRepository) Count(ctx context.Context) (int, error) {
	var count int
	if err := r.db.QueryRowContext(ctx, "SELECT count(*) FROM namespaces").Scan(&count); err != nil {
		return 0, fmt.Errorf("count namespaces: %w", err)
	}

	return count, nil
}

func (r NamespaceRepository) Create(ctx context.Context, icon string, name string, color string) (Namespace, error) {
	if icon == "" {
		return Namespace{}, errors.New("namespace icon is empty")
	}
	if name == "" {
		return Namespace{}, errors.New("namespace name is empty")
	}
	if color == "" {
		return Namespace{}, errors.New("namespace color is empty")
	}

	now := time.Now().Unix()
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO namespaces (icon, name, color, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`, icon, name, color, now, now)
	if err != nil {
		return Namespace{}, fmt.Errorf("create namespace: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return Namespace{}, fmt.Errorf("read namespace id: %w", err)
	}

	return Namespace{
		ID:        id,
		Icon:      icon,
		Name:      name,
		Color:     color,
		CreatedAt: time.Unix(now, 0).UTC(),
		UpdatedAt: time.Unix(now, 0).UTC(),
	}, nil
}

func (r NamespaceRepository) Update(ctx context.Context, id int64, icon string, name string, color string) error {
	if id == 0 {
		return errors.New("namespace id is empty")
	}
	if icon == "" {
		return errors.New("namespace icon is empty")
	}
	if name == "" {
		return errors.New("namespace name is empty")
	}
	if color == "" {
		return errors.New("namespace color is empty")
	}

	result, err := r.db.ExecContext(ctx, `
		UPDATE namespaces
		SET icon = ?, name = ?, color = ?, updated_at = ?
		WHERE id = ?
	`, icon, name, color, time.Now().Unix(), id)
	if err != nil {
		return fmt.Errorf("update namespace: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read updated namespaces count: %w", err)
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r NamespaceRepository) Delete(ctx context.Context, id int64) error {
	if id == 0 {
		return errors.New("namespace id is empty")
	}

	result, err := r.db.ExecContext(ctx, "DELETE FROM namespaces WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete namespace: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read deleted namespaces count: %w", err)
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}
