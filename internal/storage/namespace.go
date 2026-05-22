package storage

import (
	"context"
	"database/sql"
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
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (r NamespaceRepository) List(ctx context.Context) ([]Namespace, error) {
	items, err := listIconItems(ctx, r.db, namespaceTable, namespaceLabel)
	if err != nil {
		return nil, err
	}

	return toNamespaces(items), nil
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
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}
