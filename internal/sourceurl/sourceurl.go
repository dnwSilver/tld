package sourceurl

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

func Parse(raw string) (*url.URL, error) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed == nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Hostname() == "" || parsed.User != nil || parsed.Fragment != "" {
		return nil, errors.New("source URL must be an absolute http(s) URL without credentials or fragment")
	}
	if parsed.Port() == "" && strings.HasSuffix(parsed.Host, ":") {
		return nil, errors.New("source URL has an invalid port")
	}
	return parsed, nil
}

func SameOrigin(first, next *url.URL) error {
	if first == nil || next == nil || !strings.EqualFold(first.Scheme, next.Scheme) || !strings.EqualFold(first.Host, next.Host) {
		return fmt.Errorf("request changed origin")
	}
	return nil
}
