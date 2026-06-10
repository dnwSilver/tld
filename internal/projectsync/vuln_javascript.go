package projectsync

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type JavaScriptVulnStrategy struct{}

func (JavaScriptVulnStrategy) Files() []string {
	return []string{"package.json", "package-lock.json"}
}

func (JavaScriptVulnStrategy) Scan(files []File) (VulnReport, error) {
	var packageJSON []byte
	var packageLock []byte
	for _, file := range files {
		switch file.Path {
		case "package.json":
			packageJSON = file.Content
		case "package-lock.json":
			packageLock = file.Content
		}
	}
	if len(packageJSON) == 0 {
		return VulnReport{}, errors.New("package.json is required for npm audit")
	}

	output, err := runNpmAudit(packageJSON, packageLock)
	if err != nil {
		return VulnReport{}, err
	}

	return parseNpmAudit(output)
}

func runNpmAudit(packageJSON []byte, packageLock []byte) ([]byte, error) {
	dir, err := os.MkdirTemp("", "tld-npm-audit-*")
	if err != nil {
		return nil, fmt.Errorf("create npm audit temp dir: %w", err)
	}
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	if err := os.WriteFile(filepath.Join(dir, "package.json"), packageJSON, 0o600); err != nil {
		return nil, fmt.Errorf("write package.json: %w", err)
	}
	if len(packageLock) > 0 {
		if err := os.WriteFile(filepath.Join(dir, "package-lock.json"), packageLock, 0o600); err != nil {
			return nil, fmt.Errorf("write package-lock.json: %w", err)
		}
	}

	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "npm", "audit", "--json")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if len(bytes.TrimSpace(output)) == 0 {
		if err != nil {
			return nil, fmt.Errorf("npm audit failed: %w", err)
		}
		return nil, errors.New("npm audit returned empty output")
	}
	if err != nil {
		var auditErr *exec.ExitError
		if !errors.As(err, &auditErr) {
			return nil, fmt.Errorf("npm audit failed: %w", err)
		}
	}

	return output, nil
}

func parseNpmAudit(output []byte) (VulnReport, error) {
	var payload struct {
		Vulnerabilities map[string]struct {
			Name     string            `json:"name"`
			Severity string            `json:"severity"`
			Range    string            `json:"range"`
			Via      []json.RawMessage `json:"via"`
		} `json:"vulnerabilities"`
		Metadata struct {
			Vulnerabilities struct {
				Info     int `json:"info"`
				Low      int `json:"low"`
				Moderate int `json:"moderate"`
				High     int `json:"high"`
				Critical int `json:"critical"`
			} `json:"vulnerabilities"`
		} `json:"metadata"`
	}
	if err := json.Unmarshal(output, &payload); err != nil {
		return VulnReport{}, fmt.Errorf("parse npm audit output: %w", err)
	}

	report := VulnReport{
		Scanned: true,
		Counts: VulnCounts{
			Critical: payload.Metadata.Vulnerabilities.Critical,
			High:     payload.Metadata.Vulnerabilities.High,
			Medium:   payload.Metadata.Vulnerabilities.Moderate,
			Low:      payload.Metadata.Vulnerabilities.Low,
			None:     payload.Metadata.Vulnerabilities.Info,
		},
		Items: make([]Vulnerability, 0),
	}

	seen := make(map[string]struct{})
	for _, entry := range payload.Vulnerabilities {
		for _, raw := range entry.Via {
			var viaName string
			if err := json.Unmarshal(raw, &viaName); err == nil {
				continue
			}

			var advisory struct {
				Source      int    `json:"source"`
				Name        string `json:"name"`
				Dependency  string `json:"dependency"`
				Title       string `json:"title"`
				URL         string `json:"url"`
				Severity    string `json:"severity"`
				Cwe         []string `json:"cwe"`
				Cvss        struct {
					Score        float64 `json:"score"`
					VectorString string  `json:"vectorString"`
				} `json:"cvss"`
				Range       string `json:"range"`
				Description string `json:"overview"`
			}
			if err := json.Unmarshal(raw, &advisory); err != nil {
				continue
			}

			key := fmt.Sprintf("%d:%s", advisory.Source, advisory.Title)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}

			pkg := strings.TrimSpace(advisory.Dependency)
			if pkg == "" {
				pkg = strings.TrimSpace(advisory.Name)
			}
			if pkg == "" {
				pkg = strings.TrimSpace(entry.Name)
			}

			severity := normalizeVulnSeverity(advisory.Severity)
			if severity == VulnSeverityNone {
				severity = normalizeVulnSeverity(entry.Severity)
			}

			description := strings.TrimSpace(advisory.Description)
			if description == "" && advisory.URL != "" {
				description = advisory.URL
			}

			report.Items = append(report.Items, Vulnerability{
				Package:     pkg,
				Severity:    severity,
				Title:       strings.TrimSpace(advisory.Title),
				Description: description,
				Range:       firstNonEmpty(strings.TrimSpace(advisory.Range), strings.TrimSpace(entry.Range)),
			})
		}
	}

	return report, nil
}

func normalizeVulnSeverity(value string) VulnSeverity {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "critical":
		return VulnSeverityCritical
	case "high":
		return VulnSeverityHigh
	case "moderate", "medium":
		return VulnSeverityMedium
	case "low":
		return VulnSeverityLow
	case "info", "none", "":
		return VulnSeverityNone
	default:
		return VulnSeverityNone
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
