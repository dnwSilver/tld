package version

import "testing"

func TestCompareExactRejectsRangesAndPrereleases(t *testing.T) {
	for _, value := range []string{"^1.2.3", "1.2.3-beta.1", "latest", "1.2.x", "1.2.3 || 2.0.0"} {
		if _, _, _, ok := CompareExact(value, "2.0.0"); ok {
			t.Errorf("accepted non-exact version %q", value)
		}
	}
	if comparison, leftMajor, rightMajor, ok := CompareExact("v1.2.10", "1.2.3"); !ok || comparison <= 0 || leftMajor != 1 || rightMajor != 1 {
		t.Fatalf("numeric comparison: %d, %d, %d, %v", comparison, leftMajor, rightMajor, ok)
	}
}
