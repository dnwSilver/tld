package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

const (
	namespaceTable = "namespaces"
	namespaceLabel = "namespace"
)

type NamespaceRepository struct {
	db *sql.DB
}

type Namespace struct {
	ID        int64
	Icon      string
	Name      string
	Color     string
	PolicyID  int64
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (r NamespaceRepository) List(ctx context.Context) ([]Namespace, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, icon, name, color, COALESCE(policy_id, 0), created_at, updated_at
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
		if err := rows.Scan(&namespace.ID, &namespace.Icon, &namespace.Name, &namespace.Color, &namespace.PolicyID, &createdAt, &updatedAt); err != nil {
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
	return countTable(ctx, r.db, namespaceTable, namespaceLabel)
}

func (r NamespaceRepository) Create(ctx context.Context, icon string, name string, color string) (Namespace, error) {
	item, err := createIconItem(ctx, r.db, namespaceTable, namespaceLabel, icon, name, color)
	if err != nil {
		return Namespace{}, err
	}

	return toNamespace(item), nil
}

func (r NamespaceRepository) Update(ctx context.Context, id int64, icon string, name string, color string) error {
	return updateIconItem(ctx, r.db, namespaceTable, namespaceLabel, id, icon, name, color)
}

func (r NamespaceRepository) Delete(ctx context.Context, id int64) error {
	return deleteIconItem(ctx, r.db, namespaceTable, namespaceLabel, id)
}

func (r NamespaceRepository) SetPolicy(ctx context.Context, id int64, policyID int64) error {
	if id == 0 {
		return errors.New("namespace id is empty")
	}
	if policyID == 0 {
		return errors.New("namespace policy is empty")
	}

	result, err := r.db.ExecContext(ctx, `
		UPDATE namespaces
		SET policy_id = ?, updated_at = ?
		WHERE id = ?
	`, policyID, time.Now().Unix(), id)
	if err != nil {
		return fmt.Errorf("set namespace policy: %w", err)
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

func toNamespaces(items []iconItem) []Namespace {
	namespaces := make([]Namespace, 0, len(items))
	for _, item := range items {
		namespaces = append(namespaces, toNamespace(item))
	}

	return namespaces
}

func toNamespace(item iconItem) Namespace {
	return Namespace{
		ID:        item.ID,
		Icon:      item.Icon,
		Name:      item.Name,
		Color:     item.Color,
		PolicyID:  0,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}
