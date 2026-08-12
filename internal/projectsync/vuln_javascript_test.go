package projectsync

import (
	"slices"
	"testing"
)

func TestParseNpmAudit(t *testing.T) {
	output := []byte(`{
		"auditReportVersion": 2,
		"vulnerabilities": {
			"lodash": {
				"name": "lodash",
				"severity": "high",
				"range": "<4.17.21",
				"via": [
					{
						"source": 1234,
						"name": "lodash",
						"dependency": "lodash",
						"title": "Prototype Pollution",
						"url": "https://example.com/advisory",
						"severity": "high",
						"range": "<4.17.21",
						"overview": "Prototype pollution in lodash"
					}
				]
			}
		},
		"metadata": {
			"vulnerabilities": {
				"info": 1,
				"low": 2,
				"moderate": 3,
				"high": 4,
				"critical": 5
			}
		}
	}`)

	report, err := parseNpmAudit(output)
	if err != nil {
		t.Fatalf("parse npm audit: %v", err)
	}
	if !report.Scanned {
		t.Fatal("expected scanned report")
	}
	if report.Counts.Critical != 5 || report.Counts.High != 4 || report.Counts.Medium != 3 || report.Counts.Low != 2 || report.Counts.None != 1 {
		t.Fatalf("counts = %#v", report.Counts)
	}
	if len(report.Items) != 1 {
		t.Fatalf("items = %d, want 1", len(report.Items))
	}
	if report.Items[0].Title != "Prototype Pollution" || report.Items[0].Severity != VulnSeverityHigh {
		t.Fatalf("item = %#v", report.Items[0])
	}
}

func TestResolveVulnStrategyJavaScript(t *testing.T) {
	strategy, err := ResolveVulnStrategy("JavaScript", VulnScanModeDev)
	if err != nil {
		t.Fatalf("resolve strategy: %v", err)
	}
	javascript, ok := strategy.(JavaScriptVulnStrategy)
	if !ok {
		t.Fatalf("strategy = %T, want JavaScriptVulnStrategy", strategy)
	}
	if javascript.Mode != VulnScanModeDev {
		t.Fatalf("mode = %q, want %q", javascript.Mode, VulnScanModeDev)
	}
}

func TestNpmAuditArgsOmitDevDependenciesInProd(t *testing.T) {
	prod := npmAuditArgs(VulnScanModeProd)
	if !slices.Equal(prod, []string{"audit", "--omit=dev", "--json"}) {
		t.Fatalf("prod args = %v", prod)
	}

	dev := npmAuditArgs(VulnScanModeDev)
	if !slices.Equal(dev, []string{"audit", "--json"}) {
		t.Fatalf("dev args = %v", dev)
	}

	defaultMode := npmAuditArgs("")
	if !slices.Equal(defaultMode, prod) {
		t.Fatalf("default args = %v, want prod args %v", defaultMode, prod)
	}
}
