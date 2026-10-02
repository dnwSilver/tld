package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type iconItem struct {
	ID        int64
	Icon      string
	Name      string
	Color     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func listIconItems(ctx context.Context, db *sql.DB, table string, label string) ([]iconItem, error) {
	rows, err := db.QueryContext(ctx, fmt.Sprintf(`
		SELECT id, icon, name, color, created_at, updated_at
		FROM %s
		ORDER BY name ASC
	`, table))
	if err != nil {
		return nil, fmt.Errorf("list %ss: %w", label, err)
	}
	defer func() {
		_ = rows.Close()
	}()

	items := make([]iconItem, 0)
	for rows.Next() {
		item, err := scanIconItem(rows)
		if err != nil {
			return nil, fmt.Errorf("scan %s: %w", label, err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate %ss: %w", label, err)
	}

	return items, nil
}

func countTable(ctx context.Context, db *sql.DB, table string, label string) (int, error) {
	var count int
	if err := db.QueryRowContext(ctx, fmt.Sprintf("SELECT count(*) FROM %s", table)).Scan(&count); err != nil {
		return 0, fmt.Errorf("count %ss: %w", label, err)
	}

	return count, nil
}

func createIconItem(ctx context.Context, db *sql.DB, table string, label string, icon string, name string, color string) (iconItem, error) {
	if err := validateIconItem(label, icon, name, color); err != nil {
		return iconItem{}, err
	}

	now := time.Now().Unix()
	result, err := db.ExecContext(ctx, fmt.Sprintf(`
		INSERT INTO %s (icon, name, color, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`, table), icon, name, color, now, now)
	if err != nil {
		return iconItem{}, fmt.Errorf("create %s: %w", label, err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return iconItem{}, fmt.Errorf("read %s id: %w", label, err)
	}

	return iconItem{
		ID:        id,
		Icon:      icon,
		Name:      name,
		Color:     color,
		CreatedAt: time.Unix(now, 0).UTC(),
		UpdatedAt: time.Unix(now, 0).UTC(),
	}, nil
}

func updateIconItem(ctx context.Context, db *sql.DB, table string, label string, id int64, icon string, name string, color string) error {
	if err := requireID(label, id); err != nil {
		return err
	}
	if err := validateIconItem(label, icon, name, color); err != nil {
		return err
	}

	result, err := db.ExecContext(ctx, fmt.Sprintf(`
		UPDATE %s
		SET icon = ?, name = ?, color = ?, updated_at = ?
		WHERE id = ?
	`, table), icon, name, color, time.Now().Unix(), id)
	if err != nil {
		return fmt.Errorf("update %s: %w", label, err)
	}

	return requireRowsAffected(result, fmt.Sprintf("read updated %ss count", label))
}

func deleteIconItem(ctx context.Context, db *sql.DB, table string, label string, id int64) error {
	if err := requireID(label, id); err != nil {
		return err
	}

	result, err := db.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s WHERE id = ?", table), id)
	if err != nil {
		return fmt.Errorf("delete %s: %w", label, err)
	}

	return requireRowsAffected(result, fmt.Sprintf("read deleted %ss count", label))
}

func validateIconItem(label string, icon string, name string, color string) error {
	if icon == "" {
		return fmt.Errorf("%s icon is empty", label)
	}
	if name == "" {
		return fmt.Errorf("%s name is empty", label)
	}
	if color == "" {
		return fmt.Errorf("%s color is empty", label)
	}

	return nil
}

func scanIconItem(rows *sql.Rows) (iconItem, error) {
	var item iconItem
	var createdAt int64
	var updatedAt int64
	if err := rows.Scan(&item.ID, &item.Icon, &item.Name, &item.Color, &createdAt, &updatedAt); err != nil {
		return iconItem{}, err
	}
	item.CreatedAt = time.Unix(createdAt, 0).UTC()
	item.UpdatedAt = time.Unix(updatedAt, 0).UTC()

	return item, nil
}

func requireRowsAffected(result sql.Result, label string) error {
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", label, err)
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func requireID(label string, id int64) error {
	if id == 0 {
		return errors.New(label + " id is empty")
	}

	return nil
}
