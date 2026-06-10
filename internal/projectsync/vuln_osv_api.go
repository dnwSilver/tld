package projectsync

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var osvAPIBaseURL = "https://api.osv.dev"

type osvAPIQuery struct {
	Version string `json:"version"`
	Package struct {
		Name      string `json:"name"`
		Ecosystem string `json:"ecosystem"`
	} `json:"package"`
}

type osvAPIBatchRequest struct {
	Queries []osvAPIQuery `json:"queries"`
}

type osvAPIBatchResponse struct {
	Results []struct {
		Vulns []osvVulnerability `json:"vulns"`
	} `json:"results"`
}

func scanOsvAPI(ctx context.Context, files []File, strategy StackStrategy) (VulnReport, error) {
	queries := buildOsvAPIQueries(files, strategy)
	if len(queries) == 0 {
		return VulnReport{Scanned: true}, nil
	}

	report := VulnReport{
		Scanned: true,
		Items:   make([]Vulnerability, 0),
	}
	seen := make(map[string]struct{})
	client := &http.Client{Timeout: 30 * time.Second}
	enriched := make(map[string]osvVulnerability)

	for start := 0; start < len(queries); start += 100 {
		end := start + 100
		if end > len(queries) {
			end = len(queries)
		}

		batch, err := queryOsvAPIBatch(ctx, queries[start:end])
		if err != nil {
			return VulnReport{}, err
		}

		for index, result := range batch.Results {
			if index >= end-start {
				break
			}
			query := queries[start+index]
			packageName := formatOsvPackage(query.Package.Ecosystem, query.Package.Name, query.Version)
			for _, vuln := range result.Vulns {
				vuln, err := enrichOsvVulnerability(ctx, client, enriched, vuln)
				if err != nil {
					return VulnReport{}, err
				}

				key := packageName + ":" + vuln.ID
				if _, ok := seen[key]; ok {
					continue
				}
				seen[key] = struct{}{}

				title := strings.TrimSpace(vuln.Summary)
				if title == "" {
					title = vuln.ID
				}
				description := strings.TrimSpace(vuln.Details)
				if description == "" {
					description = vuln.ID
				}

				report.Items = append(report.Items, Vulnerability{
					Package:     packageName,
					Severity:    severityFromOsvVuln(vuln),
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

func buildOsvAPIQueries(files []File, strategy StackStrategy) []osvAPIQuery {
	queries := make([]osvAPIQuery, 0)
	seen := make(map[string]struct{})

	for _, path := range strategy.Files() {
		var content []byte
		for _, file := range files {
			if file.Path == path {
				content = file.Content
				break
			}
		}
		if len(content) == 0 {
			continue
		}

		dependencies, err := strategy.Parse(path, content)
		if err != nil {
			continue
		}

		for _, dependency := range dependencies {
			ecosystem, name, version, ok := osvPackageRef(dependency)
			if !ok {
				continue
			}
			key := ecosystem + ":" + name + "@" + version
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}

			query := osvAPIQuery{Version: version}
			query.Package.Name = name
			query.Package.Ecosystem = ecosystem
			queries = append(queries, query)
		}
	}

	return queries
}

func osvPackageRef(dependency Dependency) (ecosystem, name, version string, ok bool) {
	version = strings.TrimSpace(dependency.Version)
	if version == "" {
		return "", "", "", false
	}

	name = strings.TrimSpace(dependency.Name)
	if name == "" {
		return "", "", "", false
	}

	switch dependency.DependencyType {
	case DependencyTypeGradleLibrary, DependencyTypeGradlePlugin:
		return "Maven", name, version, true
	case DependencyTypeCocoaPods:
		return "SwiftURL", name, version, true
	case DependencyTypeBundler:
		return "RubyGems", name, version, true
	default:
		if strings.Contains(name, ":") {
			return "Maven", name, version, true
		}
		return "", "", "", false
	}
}

func queryOsvAPIBatch(ctx context.Context, queries []osvAPIQuery) (osvAPIBatchResponse, error) {
	payload, err := json.Marshal(osvAPIBatchRequest{Queries: queries})
	if err != nil {
		return osvAPIBatchResponse{}, fmt.Errorf("encode osv api request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, osvAPIBaseURL+"/v1/querybatch", bytes.NewReader(payload))
	if err != nil {
		return osvAPIBatchResponse{}, fmt.Errorf("create osv api request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return osvAPIBatchResponse{}, fmt.Errorf("osv api request failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return osvAPIBatchResponse{}, fmt.Errorf("read osv api response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return osvAPIBatchResponse{}, fmt.Errorf("osv api status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var response osvAPIBatchResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return osvAPIBatchResponse{}, fmt.Errorf("parse osv api response: %w", err)
	}

	return response, nil
}

func enrichOsvVulnerability(
	ctx context.Context,
	client *http.Client,
	cache map[string]osvVulnerability,
	vuln osvVulnerability,
) (osvVulnerability, error) {
	if vuln.ID == "" {
		return vuln, nil
	}
	if strings.TrimSpace(vuln.Summary) != "" && hasOsvSeveritySignal(vuln) {
		return vuln, nil
	}
	if cached, ok := cache[vuln.ID]; ok {
		return cached, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, osvAPIBaseURL+"/v1/vulns/"+vuln.ID, nil)
	if err != nil {
		return osvVulnerability{}, fmt.Errorf("create osv vuln request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return osvVulnerability{}, fmt.Errorf("osv vuln request failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return osvVulnerability{}, fmt.Errorf("read osv vuln response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return osvVulnerability{}, fmt.Errorf("osv vuln status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var full osvVulnerability
	if err := json.Unmarshal(body, &full); err != nil {
		return osvVulnerability{}, fmt.Errorf("parse osv vuln response: %w", err)
	}

	cache[vuln.ID] = full
	return full, nil
}
