package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
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
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, icon, name, color, created_at, updated_at
		FROM stacks
		ORDER BY name ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("list stacks: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	stacks := make([]Stack, 0)
	for rows.Next() {
		var stack Stack
		var createdAt int64
		var updatedAt int64
		if err := rows.Scan(&stack.ID, &stack.Icon, &stack.Name, &stack.Color, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("scan stack: %w", err)
		}
		stack.CreatedAt = time.Unix(createdAt, 0).UTC()
		stack.UpdatedAt = time.Unix(updatedAt, 0).UTC()
		stacks = append(stacks, stack)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate stacks: %w", err)
	}

	return stacks, nil
}

func (r StackRepository) Count(ctx context.Context) (int, error) {
	var count int
	if err := r.db.QueryRowContext(ctx, "SELECT count(*) FROM stacks").Scan(&count); err != nil {
		return 0, fmt.Errorf("count stacks: %w", err)
	}

	return count, nil
}

func (r StackRepository) Create(ctx context.Context, icon string, name string, color string) (Stack, error) {
	if icon == "" {
		return Stack{}, errors.New("stack icon is empty")
	}
	if name == "" {
		return Stack{}, errors.New("stack name is empty")
	}
	if color == "" {
		return Stack{}, errors.New("stack color is empty")
	}

	now := time.Now().Unix()
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO stacks (icon, name, color, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`, icon, name, color, now, now)
	if err != nil {
		return Stack{}, fmt.Errorf("create stack: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return Stack{}, fmt.Errorf("read stack id: %w", err)
	}

	return Stack{
		ID:        id,
		Icon:      icon,
		Name:      name,
		Color:     color,
		CreatedAt: time.Unix(now, 0).UTC(),
		UpdatedAt: time.Unix(now, 0).UTC(),
	}, nil
}

func (r StackRepository) Update(ctx context.Context, id int64, icon string, name string, color string) error {
	if id == 0 {
		return errors.New("stack id is empty")
	}
	if icon == "" {
		return errors.New("stack icon is empty")
	}
	if name == "" {
		return errors.New("stack name is empty")
	}
	if color == "" {
		return errors.New("stack color is empty")
	}

	result, err := r.db.ExecContext(ctx, `
		UPDATE stacks
		SET icon = ?, name = ?, color = ?, updated_at = ?
		WHERE id = ?
	`, icon, name, color, time.Now().Unix(), id)
	if err != nil {
		return fmt.Errorf("update stack: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read updated stacks count: %w", err)
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r StackRepository) Delete(ctx context.Context, id int64) error {
	if id == 0 {
		return errors.New("stack id is empty")
	}

	result, err := r.db.ExecContext(ctx, "DELETE FROM stacks WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete stack: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read deleted stacks count: %w", err)
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}
