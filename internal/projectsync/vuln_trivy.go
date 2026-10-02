package projectsync

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

var errTrivyNotFound = errors.New("trivy is not installed")

func runMobileVulnScan(dir string, files []File, strategy StackStrategy) (VulnReport, error) {
	return runMobileVulnScanContext(context.Background(), dir, files, strategy)
}

func runMobileVulnScanContext(ctx context.Context, dir string, files []File, strategy StackStrategy) (VulnReport, error) {
	osvReport, osvErr := runOsvScanContext(ctx, dir, files, strategy)
	if osvErr != nil {
		return VulnReport{}, osvErr
	}

	trivyReport, trivyErr := tryTrivyLockfileScanContext(ctx, dir, files)
	if trivyErr != nil {
		return osvReport, trivyErr
	}

	merged := mergeVulnReports(osvReport, trivyReport)
	return merged, nil
}

func tryTrivyLockfileScan(dir string, files []File) (VulnReport, error) {
	return tryTrivyLockfileScanContext(context.Background(), dir, files)
}

func tryTrivyLockfileScanContext(ctx context.Context, dir string, files []File) (VulnReport, error) {
	if !hasTrivyLockfileInFiles(files) {
		return VulnReport{}, nil
	}

	report, err := runTrivyLockfileScanContext(ctx, dir)
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
	return runTrivyLockfileScanContext(context.Background(), dir)
}

func runTrivyLockfileScanContext(ctx context.Context, dir string) (VulnReport, error) {
	binary, err := exec.LookPath("trivy")
	if err != nil {
		return VulnReport{}, errTrivyNotFound
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, binary, "fs", "--scanners", "vuln", "--format", "json", dir)
	var stdout, stderr scannerOutput
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if errors.Is(err, errScannerOutputLimit) {
		return VulnReport{}, err
	}
	output := stdout.Bytes()
	if len(bytes.TrimSpace(output)) == 0 {
		if err != nil {
			message := strings.TrimSpace(string(stderr.Bytes()))
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
		Items: make([]Vulnerability, 0),
	}
	seen := make(map[string]struct{})
	targets := 0

	for _, result := range payload.Results {
		if !isTrivyLockfileTarget(result.Target) {
			continue
		}
		targets++
		report.Scanned = true
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
	if !report.Scanned {
		return VulnReport{}, errors.New("trivy returned no lockfile results")
	}
	report.Coverage = []ScannerCoverage{{Scanner: "Trivy", Method: "lockfile results", Targets: targets}}

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
		merged.Coverage = append(merged.Coverage, report.Coverage...)
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
