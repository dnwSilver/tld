package projectsync

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

var errTrivyNotFound = errors.New("trivy is not installed")

func runMobileVulnScan(dir string, files []File, strategy StackStrategy) (VulnReport, error) {
	osvReport, osvErr := runOsvScan(dir, files, strategy)

	trivyReport, trivyErr := tryTrivyLockfileScan(dir, files)
	if trivyErr != nil {
		if osvErr != nil {
			return VulnReport{}, osvErr
		}
		return osvReport, trivyErr
	}

	merged := mergeVulnReports(osvReport, trivyReport)
	if !merged.Scanned && osvErr != nil {
		return VulnReport{}, osvErr
	}
	return merged, nil
}

func tryTrivyLockfileScan(dir string, files []File) (VulnReport, error) {
	if !hasTrivyLockfileInFiles(files) {
		return VulnReport{}, nil
	}

	report, err := runTrivyLockfileScan(dir)
	if errors.Is(err, errTrivyNotFound) {
		return VulnReport{}, nil
	}
	return report, err
}

func hasTrivyLockfileInFiles(files []File) bool {
	for _, file := range files {
		if len(file.Content) == 0 {
			continue
		}
		for _, name := range trivyLockfileNames {
			if file.Path == name || file.Path == "app/"+name {
				return true
			}
		}
	}
	return false
}

var trivyLockfileNames = []string{
	"Podfile.lock",
	"Gemfile.lock",
	"gradle.lockfile",
}

func runTrivyLockfileScan(dir string) (VulnReport, error) {
	binary, err := exec.LookPath("trivy")
	if err != nil {
		return VulnReport{}, errTrivyNotFound
	}

	ctx := context.Background()
	cmd := exec.CommandContext(ctx, binary, "fs", "--scanners", "vuln", "--format", "json", dir)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if len(bytes.TrimSpace(output)) == 0 {
		if err != nil {
			message := strings.TrimSpace(stderr.String())
			if message != "" {
				return VulnReport{}, fmt.Errorf("trivy failed: %s", message)
			}
			return VulnReport{}, fmt.Errorf("trivy failed: %w", err)
		}
		return VulnReport{}, errors.New("trivy returned empty output")
	}
	if err != nil {
		var scanErr *exec.ExitError
		if !errors.As(err, &scanErr) {
			return VulnReport{}, fmt.Errorf("trivy failed: %w", err)
		}
	}

	return parseTrivyReport(output)
}

func parseTrivyReport(output []byte) (VulnReport, error) {
	var payload struct {
		Results []struct {
			Target          string `json:"Target"`
			Vulnerabilities []struct {
				VulnerabilityID  string `json:"VulnerabilityID"`
				PkgName          string `json:"PkgName"`
				InstalledVersion string `json:"InstalledVersion"`
				Severity         string `json:"Severity"`
				Title            string `json:"Title"`
				Description      string `json:"Description"`
			} `json:"Vulnerabilities"`
		} `json:"Results"`
	}
	if err := json.Unmarshal(bytes.TrimSpace(output), &payload); err != nil {
		return VulnReport{}, fmt.Errorf("parse trivy output: %w", err)
	}

	report := VulnReport{
		Scanned: true,
		Items:   make([]Vulnerability, 0),
	}
	seen := make(map[string]struct{})

	for _, result := range payload.Results {
		if !isTrivyLockfileTarget(result.Target) {
			continue
		}
		for _, vuln := range result.Vulnerabilities {
			pkg := formatOsvPackage("", vuln.PkgName, vuln.InstalledVersion)
			title := strings.TrimSpace(vuln.Title)
			if title == "" {
				title = strings.TrimSpace(vuln.VulnerabilityID)
			}
			key := pkg + ":" + title
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}

			description := strings.TrimSpace(vuln.Description)
			if description == "" {
				description = strings.TrimSpace(vuln.VulnerabilityID)
			}

			report.Items = append(report.Items, Vulnerability{
				Package:     pkg,
				Severity:    normalizeVulnSeverity(vuln.Severity),
				Title:       title,
				Description: description,
				Range:       vuln.InstalledVersion,
			})
		}
	}

	report.Counts = countVulnSeverities(report.Items)
	return report, nil
}

func isTrivyLockfileTarget(target string) bool {
	target = strings.TrimSpace(target)
	for _, name := range trivyLockfileNames {
		if target == name || strings.HasSuffix(target, "/"+name) {
			return true
		}
	}
	return false
}

func mergeVulnReports(reports ...VulnReport) VulnReport {
	merged := VulnReport{
		Items: make([]Vulnerability, 0),
	}
	seen := make(map[string]struct{})

	for _, report := range reports {
		merged.Scanned = merged.Scanned || report.Scanned
		for _, item := range report.Items {
			title := item.Title
			if title == "" {
				title = item.Package
			}
			key := item.Package + ":" + title
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			merged.Items = append(merged.Items, item)
		}
	}

	merged.Counts = countVulnSeverities(merged.Items)
	return merged
}
