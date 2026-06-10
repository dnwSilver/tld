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
	"strconv"
	"strings"
)

var androidVulnFiles = append(
	[]string{"settings.gradle.kts", "app/gradle.lockfile", "gradle.lockfile"},
	KotlinStrategy{}.Files()...,
)

type osvVulnerability struct {
	ID       string `json:"id"`
	Summary  string `json:"summary"`
	Details  string `json:"details"`
	DatabaseSpecific struct {
		Severity string `json:"severity"`
	} `json:"database_specific"`
	Severity []struct {
		Type  string `json:"type"`
		Score string `json:"score"`
	} `json:"severity"`
	Affected []struct {
		Ranges []struct {
			Type   string `json:"type"`
			Events []struct {
				Introduced   string `json:"introduced"`
				Fixed        string `json:"fixed"`
				LastAffected string `json:"last_affected"`
			} `json:"events"`
		} `json:"ranges"`
	} `json:"affected"`
}

func runOsvScan(dir string, files []File, strategy StackStrategy) (VulnReport, error) {
	if report, err := scanOsvAPI(context.Background(), files, strategy); err == nil {
		return report, nil
	}

	output, cliErr := runOsvScannerCLI(dir)
	if cliErr == nil && len(bytes.TrimSpace(output)) > 0 {
		report, parseErr := parseOsvScanner(output)
		if parseErr == nil && !shouldUseOsvAPIFallback(output) {
			return report, nil
		}
	}
	if cliErr != nil {
		return VulnReport{}, cliErr
	}
	return VulnReport{}, errors.New("osv-scanner returned empty output")
}

func runOsvScannerCLI(dir string) ([]byte, error) {
	binary, err := exec.LookPath("osv-scanner")
	if err != nil {
		return nil, errors.New("osv-scanner is required: go install github.com/google/osv-scanner/v2/cmd/osv-scanner@latest")
	}

	ctx := context.Background()
	for _, args := range [][]string{
		{"scan", "source", "--format", "json", "-r", dir},
		{"scan", "--format", "json", "-r", dir},
	} {
		cmd := exec.CommandContext(ctx, binary, args...)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		output, err := cmd.Output()
		if len(bytes.TrimSpace(output)) > 0 {
			if err != nil {
				var scanErr *exec.ExitError
				if !errors.As(err, &scanErr) {
					return nil, fmt.Errorf("osv-scanner failed: %w", err)
				}
			}
			return extractJSONPayload(output), nil
		}
		if err != nil {
			message := strings.TrimSpace(stderr.String())
			if isBenignOsvScannerMessage(message) {
				return nil, nil
			}
			if message != "" {
				return nil, fmt.Errorf("osv-scanner failed: %s", message)
			}
			continue
		}
	}

	return nil, errors.New("osv-scanner returned empty output")
}

func isBenignOsvScannerMessage(message string) bool {
	lower := strings.ToLower(message)
	return strings.Contains(lower, "no package sources found")
}

func extractJSONPayload(output []byte) []byte {
	trimmed := bytes.TrimSpace(output)
	start := bytes.IndexByte(trimmed, '{')
	if start >= 0 {
		return trimmed[start:]
	}
	return trimmed
}

func shouldUseOsvAPIFallback(output []byte) bool {
	var payload struct {
		Results []struct {
			Packages []json.RawMessage `json:"packages"`
		} `json:"results"`
	}
	if err := json.Unmarshal(extractJSONPayload(output), &payload); err != nil {
		return true
	}
	if len(payload.Results) == 0 {
		return true
	}
	totalPackages := 0
	for _, result := range payload.Results {
		totalPackages += len(result.Packages)
	}
	return totalPackages == 0
}

func writeOsvScanFiles(dir string, files []File) error {
	for _, file := range files {
		path := filepath.Join(dir, file.Path)
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			return fmt.Errorf("create scan dir for %s: %w", file.Path, err)
		}
		if err := os.WriteFile(path, file.Content, 0o600); err != nil {
			return fmt.Errorf("write %s: %w", file.Path, err)
		}
	}
	return nil
}

func parseOsvScanner(output []byte) (VulnReport, error) {
	var payload struct {
		Results []struct {
			Packages []struct {
				Package struct {
					Name      string `json:"name"`
					Version   string `json:"version"`
					Ecosystem string `json:"ecosystem"`
				} `json:"package"`
				Vulnerabilities []osvVulnerability `json:"vulnerabilities"`
				Groups          []struct {
					IDs         []string `json:"ids"`
					MaxSeverity string   `json:"max_severity"`
				} `json:"groups"`
			} `json:"packages"`
		} `json:"results"`
	}
	if err := json.Unmarshal(output, &payload); err != nil {
		return VulnReport{}, fmt.Errorf("parse osv-scanner output: %w", err)
	}

	report := VulnReport{
		Scanned: true,
		Items:   make([]Vulnerability, 0),
	}
	seen := make(map[string]struct{})

	for _, result := range payload.Results {
		for _, pkg := range result.Packages {
			if len(pkg.Vulnerabilities) == 0 {
				continue
			}

			byID := make(map[string]osvVulnerability, len(pkg.Vulnerabilities))
			for _, vuln := range pkg.Vulnerabilities {
				if vuln.ID != "" {
					byID[vuln.ID] = vuln
				}
			}

			packageName := formatOsvPackage(pkg.Package.Ecosystem, pkg.Package.Name, pkg.Package.Version)
			groups := pkg.Groups
			if len(groups) == 0 {
				for _, vuln := range pkg.Vulnerabilities {
					groups = append(groups, struct {
						IDs         []string `json:"ids"`
						MaxSeverity string   `json:"max_severity"`
					}{IDs: []string{vuln.ID}})
				}
			}

			for _, group := range groups {
				if len(group.IDs) == 0 {
					continue
				}
				key := packageName + ":" + strings.Join(group.IDs, ",")
				if _, ok := seen[key]; ok {
					continue
				}
				seen[key] = struct{}{}

				vuln := byID[group.IDs[0]]
				title := strings.TrimSpace(vuln.Summary)
				if title == "" {
					title = group.IDs[0]
				}

				description := strings.TrimSpace(vuln.Details)
				if description == "" {
					description = group.IDs[0]
				}

				report.Items = append(report.Items, Vulnerability{
					Package:     packageName,
					Severity:    severityForOsvGroup(vuln, group.MaxSeverity),
					Title:       title,
					Description: description,
					Range:       osvAffectedRange(vuln),
				})
			}
		}
	}

	report.Counts = countVulnSeverities(report.Items)
	return report, nil
}

func formatOsvPackage(ecosystem, name, version string) string {
	name = strings.TrimSpace(name)
	version = strings.TrimSpace(version)
	ecosystem = strings.TrimSpace(ecosystem)

	switch {
	case name == "":
		return ""
	case version == "":
		return name
	case ecosystem != "":
		return ecosystem + ":" + name + "@" + version
	default:
		return name + "@" + version
	}
}

func severityForOsvGroup(vuln osvVulnerability, maxSeverity string) VulnSeverity {
	if severity := severityFromOsvVuln(vuln); severity != VulnSeverityMedium || hasOsvSeveritySignal(vuln) {
		return severity
	}
	if strings.TrimSpace(maxSeverity) != "" {
		return severityFromMaxSeverity(maxSeverity)
	}
	return VulnSeverityMedium
}

func severityFromOsvVuln(vuln osvVulnerability) VulnSeverity {
	if severity := normalizeVulnSeverity(vuln.DatabaseSpecific.Severity); severity != VulnSeverityNone {
		return severity
	}

	maxNumeric := 0.0
	maxVector := VulnSeverityNone
	for _, entry := range vuln.Severity {
		score := strings.TrimSpace(entry.Score)
		if numeric, err := strconv.ParseFloat(score, 64); err == nil {
			if numeric > maxNumeric {
				maxNumeric = numeric
			}
			continue
		}
		if vectorSeverity := severityFromCVSSVector(score); vulnSeverityRank(vectorSeverity) > vulnSeverityRank(maxVector) {
			maxVector = vectorSeverity
		}
	}
	if maxNumeric > 0 {
		return severityFromCVSS(maxNumeric)
	}
	if maxVector != VulnSeverityNone {
		return maxVector
	}
	return VulnSeverityMedium
}

func hasOsvSeveritySignal(vuln osvVulnerability) bool {
	return strings.TrimSpace(vuln.DatabaseSpecific.Severity) != "" || len(vuln.Severity) > 0
}

func severityFromCVSSVector(vector string) VulnSeverity {
	maxImpact := 0
	for _, part := range strings.Split(vector, "/") {
		if len(part) < 4 {
			continue
		}
		metric := part[:3]
		if metric != "VC:" && metric != "VI:" && metric != "VA:" {
			continue
		}
		maxImpact = max(maxImpact, impactRating(part[3:]))
	}

	switch {
	case maxImpact >= 3:
		return VulnSeverityHigh
	case maxImpact == 2:
		return VulnSeverityMedium
	case maxImpact == 1:
		return VulnSeverityLow
	default:
		return VulnSeverityNone
	}
}

func impactRating(value string) int {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "H":
		return 3
	case "M":
		return 2
	case "L":
		return 1
	default:
		return 0
	}
}

func vulnSeverityRank(severity VulnSeverity) int {
	switch severity {
	case VulnSeverityCritical:
		return 4
	case VulnSeverityHigh:
		return 3
	case VulnSeverityMedium:
		return 2
	case VulnSeverityLow:
		return 1
	default:
		return 0
	}
}

func maxCVSSFromOsvVuln(vuln osvVulnerability) float64 {
	maxScore := 0.0
	for _, entry := range vuln.Severity {
		score, err := strconv.ParseFloat(strings.TrimSpace(entry.Score), 64)
		if err != nil {
			continue
		}
		if score > maxScore {
			maxScore = score
		}
	}
	return maxScore
}

func severityFromMaxSeverity(value string) VulnSeverity {
	score, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil {
		return VulnSeverityMedium
	}
	return severityFromCVSS(score)
}

func severityFromCVSS(score float64) VulnSeverity {
	switch {
	case score >= 9.0:
		return VulnSeverityCritical
	case score >= 7.0:
		return VulnSeverityHigh
	case score >= 4.0:
		return VulnSeverityMedium
	case score > 0:
		return VulnSeverityLow
	default:
		return VulnSeverityNone
	}
}

func osvAffectedRange(vuln osvVulnerability) string {
	for _, affected := range vuln.Affected {
		for _, rng := range affected.Ranges {
			var introduced string
			var fixed string
			var lastAffected string
			for _, event := range rng.Events {
				if event.Introduced != "" {
					introduced = event.Introduced
				}
				if event.Fixed != "" {
					fixed = event.Fixed
				}
				if event.LastAffected != "" {
					lastAffected = event.LastAffected
				}
			}
			if fixed != "" {
				if introduced != "" && introduced != "0" {
					return introduced + " – < " + fixed
				}
				return "< " + fixed
			}
			if lastAffected != "" {
				return "<= " + lastAffected
			}
		}
	}
	return ""
}

func countVulnSeverities(items []Vulnerability) VulnCounts {
	counts := VulnCounts{}
	for _, item := range items {
		switch item.Severity {
		case VulnSeverityCritical:
			counts.Critical++
		case VulnSeverityHigh:
			counts.High++
		case VulnSeverityMedium:
			counts.Medium++
		case VulnSeverityLow:
			counts.Low++
		default:
			counts.None++
		}
	}
	return counts
}

func hasGradleBuildFile(files []File) bool {
	for _, file := range files {
		if file.Path == "settings.gradle.kts" || file.Path == "settings.gradle" {
			continue
		}
		if strings.HasSuffix(file.Path, ".gradle.kts") || strings.HasSuffix(file.Path, ".gradle") {
			return true
		}
	}
	return false
}
