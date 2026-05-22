package storage

import (
	"context"
	"database/sql"
	"time"
)

const (
	stackTable = "stacks"
	stackLabel = "stack"
)

type StackRepository struct {
	db *sql.DB
}

type Stack struct {
	ID        int64
	Icon      string
	Name      string
	Color     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (r StackRepository) List(ctx context.Context) ([]Stack, error) {
	items, err := listIconItems(ctx, r.db, stackTable, stackLabel)
	if err != nil {
		return nil, err
	}

	return toStacks(items), nil
}

func (r StackRepository) Count(ctx context.Context) (int, error) {
	return countTable(ctx, r.db, stackTable, stackLabel)
}

func (r StackRepository) Create(ctx context.Context, icon string, name string, color string) (Stack, error) {
	item, err := createIconItem(ctx, r.db, stackTable, stackLabel, icon, name, color)
	if err != nil {
		return Stack{}, err
	}

	return toStack(item), nil
}

func (r StackRepository) Update(ctx context.Context, id int64, icon string, name string, color string) error {
	return updateIconItem(ctx, r.db, stackTable, stackLabel, id, icon, name, color)
}

func (r StackRepository) Delete(ctx context.Context, id int64) error {
	return deleteIconItem(ctx, r.db, stackTable, stackLabel, id)
}

func toStacks(items []iconItem) []Stack {
	stacks := make([]Stack, 0, len(items))
	for _, item := range items {
		stacks = append(stacks, toStack(item))
	}

	return stacks
}

func toStack(item iconItem) Stack {
	return Stack{
		ID:        item.ID,
		Icon:      item.Icon,
		Name:      item.Name,
		Color:     item.Color,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}
