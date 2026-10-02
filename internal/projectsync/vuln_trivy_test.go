package projectsync

import "testing"

func TestParseTrivyReport(t *testing.T) {
	output := []byte(`{
		"Results": [
			{
				"Target": "Gemfile.lock",
				"Vulnerabilities": [
					{
						"VulnerabilityID": "CVE-2023-22796",
						"PkgName": "activesupport",
						"InstalledVersion": "6.1.7",
						"Severity": "HIGH",
						"Title": "DoS in Active Support",
						"Description": "Active Support denial of service"
					}
				]
			},
			{
				"Target": "README.md",
				"Vulnerabilities": [
					{
						"VulnerabilityID": "CVE-IGNORED",
						"PkgName": "ignored",
						"InstalledVersion": "1.0.0",
						"Severity": "HIGH",
						"Title": "ignored"
					}
				]
			}
		]
	}`)

	report, err := parseTrivyReport(output)
	if err != nil {
		t.Fatalf("parse trivy: %v", err)
	}
	if len(report.Items) != 1 {
		t.Fatalf("items = %d, want 1", len(report.Items))
	}
	if report.Items[0].Severity != VulnSeverityHigh {
		t.Fatalf("severity = %q", report.Items[0].Severity)
	}
	if report.Counts.High != 1 {
		t.Fatalf("counts = %#v", report.Counts)
	}
	if len(report.Coverage) != 1 || report.Coverage[0].Scanner != "Trivy" || report.Coverage[0].Targets != 1 {
		t.Fatalf("coverage = %#v", report.Coverage)
	}
}

func TestMergeVulnReportsDedupes(t *testing.T) {
	left := VulnReport{
		Scanned: true,
		Items: []Vulnerability{{
			Package:  "activesupport@6.1.7",
			Severity: VulnSeverityHigh,
			Title:    "DoS in Active Support",
		}},
		Counts: VulnCounts{High: 1},
	}
	right := VulnReport{
		Scanned: true,
		Items: []Vulnerability{{
			Package:  "activesupport@6.1.7",
			Severity: VulnSeverityHigh,
			Title:    "DoS in Active Support",
		}, {
			Package:  "addressable@2.8.1",
			Severity: VulnSeverityHigh,
			Title:    "CVE in addressable",
		}},
		Counts: VulnCounts{High: 2},
	}

	merged := mergeVulnReports(left, right)
	if len(merged.Items) != 2 {
		t.Fatalf("items = %d, want 2", len(merged.Items))
	}
	if merged.Counts.High != 2 {
		t.Fatalf("counts = %#v", merged.Counts)
	}
}

func TestHasTrivyLockfileInFiles(t *testing.T) {
	if !hasTrivyLockfileInFiles([]File{{Path: "Gemfile.lock", Content: []byte("specs:")}}) {
		t.Fatal("expected gemfile lock detection")
	}
	if hasTrivyLockfileInFiles([]File{{Path: "app/build.gradle.kts", Content: []byte("dependencies { }")}}) {
		t.Fatal("did not expect gradle file without lockfile")
	}
}
