package app

import (
	"github.com/dnwSilver/tld/internal/projectsync"
	"github.com/dnwSilver/tld/internal/ui"
	"testing"
	"time"
)

func TestReleaseMonthsKeepHotfixKinds(t *testing.T) {
	now := time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)
	for _, period := range []ui.ReleasePeriod{ui.ReleasePeriodQuarter, ui.ReleasePeriodYear} {
		months := buildReleaseEventMonths(now, []projectsync.Release{
			{CreatedAt: now, Kind: projectsync.ReleaseKindRelease},
			{CreatedAt: now.AddDate(0, 0, 1), Kind: projectsync.ReleaseKindHotfix},
		}, period)
		current := months[len(months)-1]
		releaseSlot, hotfixSlot := daySlot(15, period), daySlot(16, period)
		if !current.Marks[releaseSlot] || !current.HotfixMarks[hotfixSlot] || current.Marks[hotfixSlot] || current.HotfixMarks[releaseSlot] {
			t.Fatalf("%v: %#v", period, current)
		}
	}
}

func TestReleaseCountsUsePeriodAndCountOverlappingEvents(t *testing.T) {
	now := time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)
	for _, period := range []ui.ReleasePeriod{ui.ReleasePeriodQuarter, ui.ReleasePeriodYear} {
		end := monthStart(now).AddDate(0, 1, 0)
		start := end.AddDate(0, -period.Months(), 0)
		releases := []projectsync.Release{
			{CreatedAt: start, Kind: projectsync.ReleaseKindRelease},
			{CreatedAt: now, Kind: projectsync.ReleaseKindRelease},
			{CreatedAt: now, Kind: projectsync.ReleaseKindRelease},
			{CreatedAt: now, Kind: projectsync.ReleaseKindHotfix},
			{CreatedAt: end.Add(-time.Nanosecond), Kind: projectsync.ReleaseKindHotfix},
			{CreatedAt: start.Add(-time.Nanosecond), Kind: projectsync.ReleaseKindRelease},
			{CreatedAt: end, Kind: projectsync.ReleaseKindHotfix},
			{Kind: projectsync.ReleaseKindHotfix},
		}
		releaseCount, hotfixCount := countReleases(now, releases, period)
		if releaseCount != 3 || hotfixCount != 2 {
			t.Fatalf("%v: %d releases, %d hotfixes", period, releaseCount, hotfixCount)
		}
	}
}
