package projectsync

import "testing"

func TestParseOsvScanner(t *testing.T) {
	output := []byte(`{
		"results": [
			{
				"source": {
					"path": "/tmp/Podfile.lock",
					"type": "lockfile"
				},
				"packages": [
					{
						"package": {
							"name": "Alamofire",
							"version": "5.6.2",
							"ecosystem": "SwiftURL"
						},
						"vulnerabilities": [
							{
								"id": "GHSA-test-1234",
								"summary": "HTTP header injection",
								"details": "Malformed headers may be accepted.",
								"severity": [
									{
										"type": "CVSS_V3",
										"score": "8.1"
									}
								],
								"affected": [
									{
										"ranges": [
											{
												"type": "ECOSYSTEM",
												"events": [
													{"introduced": "5.0.0"},
													{"fixed": "5.7.0"}
												]
											}
										]
									}
								]
							},
							{
								"id": "CVE-2022-0001",
								"aliases": ["GHSA-test-1234"],
								"summary": "HTTP header injection",
								"severity": [
									{
										"type": "CVSS_V3",
										"score": "8.1"
									}
								]
							}
						],
						"groups": [
							{
								"ids": ["GHSA-test-1234", "CVE-2022-0001"]
							}
						]
					},
					{
						"package": {
							"name": "okhttp",
							"version": "4.9.0",
							"ecosystem": "Maven"
						},
						"vulnerabilities": [
							{
								"id": "GHSA-gradle-9999",
								"summary": "TLS bypass",
								"severity": [
									{
										"type": "CVSS_V3",
										"score": "3.5"
									}
								]
							}
						]
					}
				]
			}
		]
	}`)

	report, err := parseOsvScanner(output)
	if err != nil {
		t.Fatalf("parse osv-scanner: %v", err)
	}
	if !report.Scanned {
		t.Fatal("expected scanned report")
	}
	if report.Counts.Critical != 0 || report.Counts.High != 1 || report.Counts.Medium != 0 || report.Counts.Low != 1 {
		t.Fatalf("counts = %#v", report.Counts)
	}
	if len(report.Items) != 2 {
		t.Fatalf("items = %d, want 2", len(report.Items))
	}
	if report.Items[0].Severity != VulnSeverityHigh || report.Items[0].Title != "HTTP header injection" {
		t.Fatalf("first item = %#v", report.Items[0])
	}
	if report.Items[0].Range != "5.0.0 – < 5.7.0" {
		t.Fatalf("range = %q", report.Items[0].Range)
	}
	if report.Items[1].Severity != VulnSeverityLow {
		t.Fatalf("second item = %#v", report.Items[1])
	}
}

func TestResolveVulnStrategyMobile(t *testing.T) {
	android, err := ResolveVulnStrategy("Android")
	if err != nil {
		t.Fatalf("resolve android strategy: %v", err)
	}
	if _, ok := android.(AndroidVulnStrategy); !ok {
		t.Fatalf("android strategy = %T", android)
	}

	ios, err := ResolveVulnStrategy("iOS")
	if err != nil {
		t.Fatalf("resolve ios strategy: %v", err)
	}
	if _, ok := ios.(IOSVulnStrategy); !ok {
		t.Fatalf("ios strategy = %T", ios)
	}
}

func TestAndroidVulnStrategyRequiresGradleFile(t *testing.T) {
	_, err := AndroidVulnStrategy{}.Scan([]File{{Path: "settings.gradle.kts", Content: []byte("rootProject.name = \"demo\"")}})
	if err == nil {
		t.Fatal("expected error without gradle build file")
	}
}

func TestIOSVulnStrategyRequiresLockfile(t *testing.T) {
	_, err := IOSVulnStrategy{}.Scan([]File{{Path: "Podfile", Content: []byte("platform :ios, '13.0'\n")}})
	if err == nil {
		t.Fatal("expected error without ios lockfile")
	}
}
