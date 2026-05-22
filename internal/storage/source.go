package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

const (
	SourceTypeGitLab    = "gitlab"
	SourceTypeGitHub    = "github"
	SourceTypeGitea     = "gitea"
	SourceTypeBitbucket = "bitbucket"
)

type SourceRepository struct {
	db *sql.DB
}

type Source struct {
	ID        int64
	Icon      string
	Name      string
	Color     string
	PATToken  string
	URL       string
	Type      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type sourceAppearance struct {
	icon  string
	color string
}

var sourceAppearances = map[string]sourceAppearance{
	SourceTypeGitLab:    {icon: "", color: "FC6D26"},
	SourceTypeGitHub:    {icon: "", color: "FFFFFF"},
	SourceTypeBitbucket: {icon: "", color: "1A74ED"},
	SourceTypeGitea:     {icon: "", color: "609926"},
}

func (r SourceRepository) List(ctx context.Context) ([]Source, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, icon, name, color, pat_token, url, type, created_at, updated_at
		FROM sources
		ORDER BY name ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("list sources: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	sources := make([]Source, 0)
	for rows.Next() {
		var source Source
		var createdAt int64
		var updatedAt int64
		if err := rows.Scan(
			&source.ID,
			&source.Icon,
			&source.Name,
			&source.Color,
			&source.PATToken,
			&source.URL,
			&source.Type,
			&createdAt,
			&updatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan source: %w", err)
		}
		source.CreatedAt = time.Unix(createdAt, 0).UTC()
		source.UpdatedAt = time.Unix(updatedAt, 0).UTC()
		sources = append(sources, source)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate sources: %w", err)
	}

	return sources, nil
}

func (r SourceRepository) Count(ctx context.Context) (int, error) {
	var count int
	if err := r.db.QueryRowContext(ctx, "SELECT count(*) FROM sources").Scan(&count); err != nil {
		return 0, fmt.Errorf("count sources: %w", err)
	}

	return count, nil
}

func (r SourceRepository) Create(ctx context.Context, name string, patToken string, sourceURL string, sourceType string) (Source, error) {
	if name == "" {
		return Source{}, errors.New("source name is empty")
	}
	if patToken == "" {
		return Source{}, errors.New("source PAT token is empty")
	}
	if sourceURL == "" {
		return Source{}, errors.New("source url is empty")
	}
	appearance, err := getSourceAppearance(sourceType)
	if err != nil {
		return Source{}, err
	}

	now := time.Now().Unix()
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO sources (icon, name, color, pat_token, url, type, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, appearance.icon, name, appearance.color, patToken, sourceURL, sourceType, now, now)
	if err != nil {
		return Source{}, fmt.Errorf("create source: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return Source{}, fmt.Errorf("read source id: %w", err)
	}

	return Source{
		ID:        id,
		Icon:      appearance.icon,
		Name:      name,
		Color:     appearance.color,
		PATToken:  patToken,
		URL:       sourceURL,
		Type:      sourceType,
		CreatedAt: time.Unix(now, 0).UTC(),
		UpdatedAt: time.Unix(now, 0).UTC(),
	}, nil
}

func (r SourceRepository) Update(ctx context.Context, id int64, name string, patToken string, sourceURL string, sourceType string) error {
	if id == 0 {
		return errors.New("source id is empty")
	}
	if name == "" {
		return errors.New("source name is empty")
	}
	if patToken == "" {
		return errors.New("source PAT token is empty")
	}
	if sourceURL == "" {
		return errors.New("source url is empty")
	}
	appearance, err := getSourceAppearance(sourceType)
	if err != nil {
		return err
	}

	result, err := r.db.ExecContext(ctx, `
		UPDATE sources
		SET icon = ?, name = ?, color = ?, pat_token = ?, url = ?, type = ?, updated_at = ?
		WHERE id = ?
	`, appearance.icon, name, appearance.color, patToken, sourceURL, sourceType, time.Now().Unix(), id)
	if err != nil {
		return fmt.Errorf("update source: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read updated sources count: %w", err)
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r SourceRepository) Delete(ctx context.Context, id int64) error {
	if id == 0 {
		return errors.New("source id is empty")
	}

	result, err := r.db.ExecContext(ctx, "DELETE FROM sources WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete source: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read deleted sources count: %w", err)
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func getSourceAppearance(sourceType string) (sourceAppearance, error) {
	appearance, ok := sourceAppearances[sourceType]
	if !ok {
		return sourceAppearance{}, errors.New("source type is invalid")
	}

	return appearance, nil
}
