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
