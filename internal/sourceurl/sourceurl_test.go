package sourceurl

import "testing"

func TestParseRequiresExplicitHTTPOrigin(t *testing.T) {
	for _, raw := range []string{"gitlab.internal", "//gitlab.internal", "https://", "https://user:pass@gitlab.internal", "ftp://gitlab.internal"} {
		if _, err := Parse(raw); err == nil {
			t.Errorf("accepted invalid source URL %q", raw)
		}
	}
	for _, raw := range []string{"https://gitlab.example.com", "http://127.0.0.1:8080"} {
		if _, err := Parse(raw); err != nil {
			t.Errorf("rejected valid source URL %q: %v", raw, err)
		}
	}
}

func TestRedirectOriginPolicy(t *testing.T) {
	first, _ := Parse("https://gitlab.example.com/group")
	for _, raw := range []string{"https://evil.example.com/path", "http://gitlab.example.com/path"} {
		next, _ := Parse(raw)
		if err := SameOrigin(first, next); err == nil {
			t.Errorf("accepted redirect to %q", raw)
		}
	}
	next, _ := Parse("https://gitlab.example.com/other")
	if err := SameOrigin(first, next); err != nil {
		t.Fatalf("same-origin redirect rejected: %v", err)
	}
}
