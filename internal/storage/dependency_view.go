package storage

import (
	"context"
	"fmt"
	"strings"
)

type DependencyView struct {
	StackID   int64
	StackName string
	Columns   []DependencyViewColumn
	Rows      []DependencyViewRow
}

type DependencyViewColumn struct {
	DependencyID  int64
	Icon          string
	Name          string
	Color         string
	PolicyVersion string
}

type DependencyViewRow struct {
	ProjectID        int64
	ProjectIcon      string
	ProjectName      string
	ProjectColor     string
	ProjectFreezing  bool
	ProjectEndOfLife bool
	Versions         map[int64]string
}

func (r ProjectDependencyRepository) ViewByStack(ctx context.Context, stackID int64) (DependencyView, error) {
	if stackID == 0 {
		return DependencyView{}, nil
	}

	view := DependencyView{StackID: stackID}
	if err := r.db.QueryRowContext(ctx, "SELECT name FROM stacks WHERE id = ?", stackID).Scan(&view.StackName); err != nil {
		return DependencyView{}, fmt.Errorf("read view stack: %w", err)
	}

	columns, err := r.viewColumns(ctx, stackID)
	if err != nil {
		return DependencyView{}, err
	}
	rows, err := r.viewRows(ctx, stackID)
	if err != nil {
		return DependencyView{}, err
	}
	versions, err := r.viewVersions(ctx, stackID)
	if err != nil {
		return DependencyView{}, err
	}

	for index := range rows {
		rows[index].Versions = map[int64]string{}
		for _, column := range columns {
			rows[index].Versions[column.DependencyID] = versions[rows[index].ProjectID][normalizeDependencyViewName(column.Name)]
		}
	}

	view.Columns = columns
	view.Rows = rows
	return view, nil
}

func (r ProjectDependencyRepository) viewColumns(ctx context.Context, stackID int64) ([]DependencyViewColumn, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT DISTINCT d.id, d.icon, d.name, d.color, pv.version
		FROM projects p
		JOIN namespaces n ON n.id = p.namespace_id
		JOIN policy_values pv ON pv.policy_id = n.policy_id
		JOIN dependencies d ON d.id = pv.dependency_id
		WHERE p.stack_id = ? AND d.stack_id = ? AND p.endoflife = 0
		ORDER BY d.name ASC, pv.version ASC
	`, stackID, stackID)
	if err != nil {
		return nil, fmt.Errorf("list dependency view columns: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	columns := make([]DependencyViewColumn, 0)
	for rows.Next() {
		var column DependencyViewColumn
		if err := rows.Scan(&column.DependencyID, &column.Icon, &column.Name, &column.Color, &column.PolicyVersion); err != nil {
			return nil, fmt.Errorf("scan dependency view column: %w", err)
		}
		columns = append(columns, column)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate dependency view columns: %w", err)
	}

	return columns, nil
}

func (r ProjectDependencyRepository) viewRows(ctx context.Context, stackID int64) ([]DependencyViewRow, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, icon, name, color, freezing, endoflife
		FROM projects
		WHERE stack_id = ? AND endoflife = 0
		ORDER BY name ASC
	`, stackID)
	if err != nil {
		return nil, fmt.Errorf("list dependency view rows: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	viewRows := make([]DependencyViewRow, 0)
	for rows.Next() {
		var row DependencyViewRow
		var freezing int
		var endOfLife int
		if err := rows.Scan(&row.ProjectID, &row.ProjectIcon, &row.ProjectName, &row.ProjectColor, &freezing, &endOfLife); err != nil {
			return nil, fmt.Errorf("scan dependency view row: %w", err)
		}
		row.ProjectFreezing = freezing != 0
		row.ProjectEndOfLife = endOfLife != 0
		viewRows = append(viewRows, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate dependency view rows: %w", err)
	}

	return viewRows, nil
}

func (r ProjectDependencyRepository) viewVersions(ctx context.Context, stackID int64) (map[int64]map[string]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT p.id, pd.name, pd.version
		FROM projects p
		JOIN project_dependencies pd ON pd.project_id = p.id
		WHERE p.stack_id = ? AND p.endoflife = 0
	`, stackID)
	if err != nil {
		return nil, fmt.Errorf("list dependency view versions: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	versions := map[int64]map[string]string{}
	for rows.Next() {
		var projectID int64
		var name string
		var version string
		if err := rows.Scan(&projectID, &name, &version); err != nil {
			return nil, fmt.Errorf("scan dependency view version: %w", err)
		}
		if versions[projectID] == nil {
			versions[projectID] = map[string]string{}
		}
		versions[projectID][normalizeDependencyViewName(name)] = version
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate dependency view versions: %w", err)
	}

	return versions, nil
}

func normalizeDependencyViewName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	replacer := strings.NewReplacer(" ", "", "-", "", "_", "", ".", "", "/", "", ":", "", "@", "")
	normalized := replacer.Replace(name)

	switch normalized {
	case "nodejs":
		return "node"
	default:
		return normalized
	}
}
