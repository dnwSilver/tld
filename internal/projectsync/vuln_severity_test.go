package projectsync

import "testing"

func TestSeverityFromCVSSVector(t *testing.T) {
	severity := severityFromCVSSVector("CVSS:4.0/AV:N/AC:L/AT:P/PR:N/UI:N/VC:H/VI:N/VA:N/SC:N/SI:N/SA:N")
	if severity != VulnSeverityHigh {
		t.Fatalf("severity = %q, want high", severity)
	}
}

func TestSeverityFromDatabaseSpecific(t *testing.T) {
	severity := severityFromOsvVuln(osvVulnerability{
		DatabaseSpecific: struct {
			Severity string `json:"severity"`
		}{Severity: "HIGH"},
	})
	if severity != VulnSeverityHigh {
		t.Fatalf("severity = %q, want high", severity)
	}
}
