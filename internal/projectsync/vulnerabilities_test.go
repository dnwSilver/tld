package projectsync

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/dnwSilver/tld/internal/storage"
)

func TestVulnerabilityReportsAreCachedPerMode(t *testing.T) {
	ctx := context.Background()
	store, err := storage.Open(ctx, filepath.Join(t.TempDir(), "tld.db"), "secret")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer func() {
		_ = store.Close()
	}()

	source := Source{Type: storage.SourceTypeGitHub}
	project := Project{ProviderID: "owner/repo"}
	prodReport := VulnReport{
		Scanned: true, Outcome: "complete", Revision: "abcdef1234567890", ScannedAt: time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC),
		Counts: VulnCounts{High: 1},
		Items:  []Vulnerability{{Package: "runtime", Severity: VulnSeverityHigh}},
	}
	devReport := VulnReport{
		Scanned: true,
		Counts:  VulnCounts{High: 2},
		Items: []Vulnerability{
			{Package: "runtime", Severity: VulnSeverityHigh},
			{Package: "dev-only", Severity: VulnSeverityHigh},
		},
	}

	prodService := VulnScanService{Cache: store.Cache(), Mode: VulnScanModeProd}
	if err := prodService.cacheReport(ctx, source, project, prodReport); err != nil {
		t.Fatalf("cache prod report: %v", err)
	}
	failedAt := prodReport.ScannedAt.Add(time.Hour)
	if err := prodService.cacheAttempt(ctx, source, project, VulnScanAttempt{Outcome: VulnScanFailed, At: failedAt}); err != nil {
		t.Fatalf("cache failed attempt: %v", err)
	}
	devService := VulnScanService{Cache: store.Cache(), Mode: VulnScanModeDev}
	if err := devService.cacheReport(ctx, source, project, devReport); err != nil {
		t.Fatalf("cache dev report: %v", err)
	}

	loadedProd, err := prodService.LoadProject(ctx, source, project)
	if err != nil {
		t.Fatalf("load prod report: %v", err)
	}
	if len(loadedProd.Items) != 1 || loadedProd.Items[0].Package != "runtime" || loadedProd.Outcome != "complete" || loadedProd.Revision != prodReport.Revision || !loadedProd.ScannedAt.Equal(prodReport.ScannedAt) || loadedProd.LastAttempt.Outcome != VulnScanFailed || !loadedProd.LastAttempt.At.Equal(failedAt) {
		t.Fatalf("prod report = %#v", loadedProd)
	}

	loadedDev, err := devService.LoadProject(ctx, source, project)
	if err != nil {
		t.Fatalf("load dev report: %v", err)
	}
	if len(loadedDev.Items) != 2 || loadedDev.Items[1].Package != "dev-only" {
		t.Fatalf("dev report = %#v", loadedDev)
	}
	batch, err := prodService.LoadProjects(ctx, []VulnProjectRef{
		{Source: source, Project: Project{ProviderID: "missing"}},
		{Source: source, Project: project},
	})
	if err != nil {
		t.Fatalf("load report batch: %v", err)
	}
	if len(batch) != 2 || batch[0].Scanned || !batch[1].Scanned || batch[1].Counts.High != 1 {
		t.Fatalf("batch reports = %#v", batch)
	}
}

func TestUnsupportedScanPersistsLastAttemptOutcome(t *testing.T) {
	ctx := context.Background()
	store, err := storage.Open(ctx, filepath.Join(t.TempDir(), "unsupported.db"), "secret")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	service := VulnScanService{Cache: store.Cache(), SourceClient: &fakeSourceClient{}}
	source := Source{Type: storage.SourceTypeGitHub}
	project := Project{ProviderID: "owner/repo", StackName: "unknown-stack"}
	if _, err := service.RunProject(ctx, source, project, nil); !errors.Is(err, ErrUnsupportedVulnStrategy) {
		t.Fatalf("error = %v", err)
	}
	report, err := service.LoadProject(ctx, source, project)
	if err != nil {
		t.Fatal(err)
	}
	if report.Scanned || report.LastAttempt.Outcome != VulnScanUnsupported || report.LastAttempt.At.IsZero() {
		t.Fatalf("report = %#v", report)
	}
}
