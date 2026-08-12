package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type PolicyRepository struct {
	db *sql.DB
}

type Policy struct {
	ID              int64
	Name            string
	DependencyCount int
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type PolicyValue struct {
	ID                 int64
	PolicyID           int64
	DependencyID       int64
	DependencyIcon     string
	DependencyName     string
	DependencyColor    string
	StackID            int64
	StackName          string
	RegistryName       string
	RegistrySourceID   int64
	RegistrySourceName string
	Version            string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

func (r PolicyRepository) List(ctx context.Context) ([]Policy, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT p.id, p.name, count(pv.id), p.created_at, p.updated_at
		FROM policies p
		LEFT JOIN policy_values pv ON pv.policy_id = p.id
		GROUP BY p.id, p.name, p.created_at, p.updated_at
		ORDER BY p.name ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("list policies: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	policies := make([]Policy, 0)
	for rows.Next() {
		var policy Policy
		var createdAt int64
		var updatedAt int64
		if err := rows.Scan(&policy.ID, &policy.Name, &policy.DependencyCount, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("scan policy: %w", err)
		}
		policy.CreatedAt = time.Unix(createdAt, 0).UTC()
		policy.UpdatedAt = time.Unix(updatedAt, 0).UTC()
		policies = append(policies, policy)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate policies: %w", err)
	}

	return policies, nil
}

func (r PolicyRepository) Count(ctx context.Context) (int, error) {
	var count int
	if err := r.db.QueryRowContext(ctx, "SELECT count(*) FROM policies").Scan(&count); err != nil {
		return 0, fmt.Errorf("count policies: %w", err)
	}

	return count, nil
}

func (r PolicyRepository) Create(ctx context.Context, name string) (Policy, error) {
	if name == "" {
		return Policy{}, errors.New("policy name is empty")
	}

	now := time.Now().Unix()
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO policies (name, created_at, updated_at)
		VALUES (?, ?, ?)
	`, name, now, now)
	if err != nil {
		return Policy{}, fmt.Errorf("create policy: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return Policy{}, fmt.Errorf("read policy id: %w", err)
	}

	return Policy{
		ID:        id,
		Name:      name,
		CreatedAt: time.Unix(now, 0).UTC(),
		UpdatedAt: time.Unix(now, 0).UTC(),
	}, nil
}

func (r PolicyRepository) AssignToNamespace(ctx context.Context, namespaceID int64, name string) (Policy, error) {
	if namespaceID == 0 {
		return Policy{}, errors.New("policy namespace is empty")
	}
	if name == "" {
		return Policy{}, errors.New("policy name is empty")
	}

	now := time.Now().Unix()
	if _, err := r.db.ExecContext(ctx, `
		INSERT OR IGNORE INTO policies (name, created_at, updated_at)
		VALUES (?, ?, ?)
	`, name, now, now); err != nil {
		return Policy{}, fmt.Errorf("upsert policy: %w", err)
	}

	policy, err := r.GetByName(ctx, name)
	if err != nil {
		return Policy{}, err
	}
	if _, err := r.db.ExecContext(ctx, `
		UPDATE namespaces
		SET policy_id = ?, updated_at = ?
		WHERE id = ?
	`, policy.ID, now, namespaceID); err != nil {
		return Policy{}, fmt.Errorf("assign namespace policy: %w", err)
	}

	return policy, nil
}

func (r PolicyRepository) GetByName(ctx context.Context, name string) (Policy, error) {
	var policy Policy
	var createdAt int64
	var updatedAt int64
	if err := r.db.QueryRowContext(ctx, `
		SELECT p.id, p.name, count(pv.id), p.created_at, p.updated_at
		FROM policies p
		LEFT JOIN policy_values pv ON pv.policy_id = p.id
		WHERE p.name = ?
		GROUP BY p.id, p.name, p.created_at, p.updated_at
	`, name).Scan(&policy.ID, &policy.Name, &policy.DependencyCount, &createdAt, &updatedAt); err != nil {
		return Policy{}, fmt.Errorf("get policy by name: %w", err)
	}
	policy.CreatedAt = time.Unix(createdAt, 0).UTC()
	policy.UpdatedAt = time.Unix(updatedAt, 0).UTC()

	return policy, nil
}

func (r PolicyRepository) Update(ctx context.Context, id int64, name string) error {
	if id == 0 {
		return errors.New("policy id is empty")
	}
	if name == "" {
		return errors.New("policy name is empty")
	}

	result, err := r.db.ExecContext(ctx, `
		UPDATE policies
		SET name = ?, updated_at = ?
		WHERE id = ?
	`, name, time.Now().Unix(), id)
	if err != nil {
		return fmt.Errorf("update policy: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read updated policies count: %w", err)
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r PolicyRepository) Delete(ctx context.Context, id int64) error {
	if id == 0 {
		return errors.New("policy id is empty")
	}

	result, err := r.db.ExecContext(ctx, "DELETE FROM policies WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete policy: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read deleted policies count: %w", err)
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r PolicyRepository) ListValues(ctx context.Context, policyID int64) ([]PolicyValue, error) {
	if policyID == 0 {
		return []PolicyValue{}, nil
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT pv.id, pv.policy_id, pv.dependency_id, d.stack_id, s.name, d.icon, d.name, d.color, d.registry_name, d.registry_source_id, COALESCE(src.name, ''), pv.version, pv.created_at, pv.updated_at
		FROM policy_values pv
		JOIN dependencies d ON d.id = pv.dependency_id
		JOIN stacks s ON s.id = d.stack_id
		LEFT JOIN sources src ON src.id = d.registry_source_id AND src.type = 'registry'
		WHERE pv.policy_id = ?
		ORDER BY d.name ASC
	`, policyID)
	if err != nil {
		return nil, fmt.Errorf("list policy values: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	values := make([]PolicyValue, 0)
	for rows.Next() {
		var value PolicyValue
		var createdAt int64
		var updatedAt int64
		if err := rows.Scan(
			&value.ID,
			&value.PolicyID,
			&value.DependencyID,
			&value.StackID,
			&value.StackName,
			&value.DependencyIcon,
			&value.DependencyName,
			&value.DependencyColor,
			&value.RegistryName,
			&value.RegistrySourceID,
			&value.RegistrySourceName,
			&value.Version,
			&createdAt,
			&updatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan policy value: %w", err)
		}
		value.CreatedAt = time.Unix(createdAt, 0).UTC()
		value.UpdatedAt = time.Unix(updatedAt, 0).UTC()
		values = append(values, value)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate policy values: %w", err)
	}

	return values, nil
}

func (r PolicyRepository) ListValuesByStack(ctx context.Context, policyID int64, stackID int64) ([]PolicyValue, error) {
	values, err := r.ListValues(ctx, policyID)
	if err != nil {
		return nil, err
	}
	filtered := make([]PolicyValue, 0, len(values))
	for _, value := range values {
		if value.StackID == stackID {
			filtered = append(filtered, value)
		}
	}
	return filtered, nil
}

func (r PolicyRepository) UpdateValueVersion(ctx context.Context, id int64, version string) error {
	if id == 0 {
		return errors.New("policy value id is empty")
	}
	if version == "" {
		return errors.New("policy value version is empty")
	}
	result, err := r.db.ExecContext(ctx, `UPDATE policy_values SET version = ?, updated_at = ? WHERE id = ?`, version, time.Now().Unix(), id)
	if err != nil {
		return fmt.Errorf("update policy value version: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read updated policy values count: %w", err)
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r PolicyRepository) CreateValue(ctx context.Context, policyID int64, dependencyID int64, version string) (PolicyValue, error) {
	if policyID == 0 {
		return PolicyValue{}, errors.New("policy value policy is empty")
	}
	if dependencyID == 0 {
		return PolicyValue{}, errors.New("policy value dependency is empty")
	}
	if version == "" {
		return PolicyValue{}, errors.New("policy value version is empty")
	}

	now := time.Now().Unix()
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO policy_values (policy_id, dependency_id, version, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`, policyID, dependencyID, version, now, now)
	if err != nil {
		return PolicyValue{}, fmt.Errorf("create policy value: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return PolicyValue{}, fmt.Errorf("read policy value id: %w", err)
	}

	return PolicyValue{
		ID:           id,
		PolicyID:     policyID,
		DependencyID: dependencyID,
		Version:      version,
		CreatedAt:    time.Unix(now, 0).UTC(),
		UpdatedAt:    time.Unix(now, 0).UTC(),
	}, nil
}

func (r PolicyRepository) UpdateValue(ctx context.Context, id int64, dependencyID int64, version string) error {
	if id == 0 {
		return errors.New("policy value id is empty")
	}
	if dependencyID == 0 {
		return errors.New("policy value dependency is empty")
	}
	if version == "" {
		return errors.New("policy value version is empty")
	}

	result, err := r.db.ExecContext(ctx, `
		UPDATE policy_values
		SET dependency_id = ?, version = ?, updated_at = ?
		WHERE id = ?
	`, dependencyID, version, time.Now().Unix(), id)
	if err != nil {
		return fmt.Errorf("update policy value: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read updated policy values count: %w", err)
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r PolicyRepository) DeleteValue(ctx context.Context, id int64) error {
	if id == 0 {
		return errors.New("policy value id is empty")
	}

	result, err := r.db.ExecContext(ctx, "DELETE FROM policy_values WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete policy value: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read deleted policy values count: %w", err)
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}
