package registry

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	KindNPM       = "npm"
	KindGo        = "go"
	KindMaven     = "maven"
	KindRubyGems  = "rubygems"
	KindCocoaPods = "cocoapods"
)

var Kinds = []string{KindNPM, KindGo, KindMaven, KindRubyGems, KindCocoaPods}

var ErrPackageNotFound = errors.New("package not found")

type Source struct {
	URL   string
	Token string
	Kind  string
}

type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	HTTPClient HTTPDoer
}

func (c Client) LatestStable(ctx context.Context, source Source, packageName string) (string, error) {
	if strings.TrimSpace(source.URL) == "" {
		return "", errors.New("registry URL is empty")
	}
	packageName = strings.TrimSpace(packageName)
	if packageName == "" {
		return "", errors.New("registry package name is empty")
	}
	if c.HTTPClient == nil {
		c.HTTPClient = &http.Client{Timeout: 20 * time.Second}
	}

	var versions []string
	var err error
	switch source.Kind {
	case KindNPM:
		versions, err = c.npmVersions(ctx, source, normalizeNPMPackageName(packageName))
	case KindGo:
		versions, err = c.goVersions(ctx, source, packageName)
	case KindMaven:
		versions, err = c.mavenVersions(ctx, source, packageName)
	case KindRubyGems:
		versions, err = c.rubyGemVersions(ctx, source, packageName)
	case KindCocoaPods:
		versions, err = c.cocoaPodsVersions(ctx, source, packageName)
	default:
		return "", fmt.Errorf("unsupported registry kind %q", source.Kind)
	}
	if err != nil {
		return "", err
	}
	latest, found := latestStableVersion(versions)
	if !found {
		return "", fmt.Errorf("%w: %s has no stable versions", ErrPackageNotFound, packageName)
	}
	return latest, nil
}

func IsKind(value string) bool {
	for _, kind := range Kinds {
		if value == kind {
			return true
		}
	}
	return false
}

func (c Client) npmVersions(ctx context.Context, source Source, packageName string) ([]string, error) {
	endpoint := registryURL(source.URL, url.PathEscape(packageName))
	var metadata struct {
		Versions map[string]json.RawMessage `json:"versions"`
	}
	if err := c.getJSON(ctx, source, endpoint, &metadata); err != nil {
		return nil, err
	}
	versions := make([]string, 0, len(metadata.Versions))
	for version := range metadata.Versions {
		versions = append(versions, version)
	}
	return versions, nil
}

func (c Client) goVersions(ctx context.Context, source Source, module string) ([]string, error) {
	parts := strings.Split(module, "/")
	for index, part := range parts {
		parts[index] = url.PathEscape(escapeGoModulePart(part))
	}
	endpoint := registryURL(source.URL, strings.Join(parts, "/"), "@v", "list")
	body, err := c.get(ctx, source, endpoint, "text/plain")
	if err != nil {
		return nil, err
	}
	return strings.Fields(string(body)), nil
}

func (c Client) mavenVersions(ctx context.Context, source Source, coordinate string) ([]string, error) {
	group, artifact, found := strings.Cut(coordinate, ":")
	if !found || strings.TrimSpace(group) == "" || strings.TrimSpace(artifact) == "" || strings.Contains(artifact, ":") {
		return nil, fmt.Errorf("maven coordinate %q must be group:artifact", coordinate)
	}
	groupPath := strings.ReplaceAll(strings.TrimSpace(group), ".", "/")
	endpoint := registryURL(source.URL, groupPath, strings.TrimSpace(artifact), "maven-metadata.xml")
	body, err := c.get(ctx, source, endpoint, "application/xml")
	if err != nil {
		return nil, err
	}
	var metadata struct {
		Versioning struct {
			Release  string   `xml:"release"`
			Versions []string `xml:"versions>version"`
		} `xml:"versioning"`
	}
	if err := xml.Unmarshal(body, &metadata); err != nil {
		return nil, fmt.Errorf("decode maven metadata: %w", err)
	}
	if metadata.Versioning.Release != "" {
		metadata.Versioning.Versions = append(metadata.Versioning.Versions, metadata.Versioning.Release)
	}
	return metadata.Versioning.Versions, nil
}

func (c Client) rubyGemVersions(ctx context.Context, source Source, gem string) ([]string, error) {
	endpoint := registryURL(source.URL, "api", "v1", "versions", url.PathEscape(gem)+".json")
	var response []struct {
		Number     string `json:"number"`
		Prerelease bool   `json:"prerelease"`
	}
	if err := c.getJSON(ctx, source, endpoint, &response); err != nil {
		return nil, err
	}
	versions := make([]string, 0, len(response))
	for _, item := range response {
		if !item.Prerelease {
			versions = append(versions, item.Number)
		}
	}
	return versions, nil
}

func (c Client) cocoaPodsVersions(ctx context.Context, source Source, pod string) ([]string, error) {
	pod = strings.TrimSpace(strings.SplitN(pod, "/", 2)[0])
	digest := md5.Sum([]byte(pod))
	hash := hex.EncodeToString(digest[:])
	endpoint := registryURL(source.URL, fmt.Sprintf("all_pods_versions_%c_%c_%c.txt", hash[0], hash[1], hash[2]))
	body, err := c.get(ctx, source, endpoint, "text/plain")
	if err != nil {
		return nil, err
	}
	prefix := pod + "/"
	for _, line := range strings.Split(strings.ReplaceAll(string(body), "\r", ""), "\n") {
		if strings.HasPrefix(line, prefix) {
			return strings.Split(strings.TrimPrefix(line, prefix), "/"), nil
		}
	}
	return nil, fmt.Errorf("%w: %s", ErrPackageNotFound, pod)
}

func (c Client) getJSON(ctx context.Context, source Source, endpoint string, target any) error {
	body, err := c.get(ctx, source, endpoint, "application/json")
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("decode registry response: %w", err)
	}
	return nil
}

func (c Client) get(ctx context.Context, source Source, endpoint string, accept string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create registry request: %w", err)
	}
	if accept != "" {
		req.Header.Set("Accept", accept)
	}
	if source.Token != "" {
		req.Header.Set("Authorization", "Bearer "+source.Token)
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("registry request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusGone {
		return nil, ErrPackageNotFound
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("registry request failed with status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 16<<20))
	if err != nil {
		return nil, fmt.Errorf("read registry response: %w", err)
	}
	return body, nil
}

func registryURL(base string, elements ...string) string {
	return strings.TrimRight(strings.TrimSpace(base), "/") + "/" + strings.Join(elements, "/")
}

func normalizeNPMPackageName(name string) string {
	if strings.Contains(name, "/") && !strings.HasPrefix(name, "@") {
		return "@" + name
	}
	return name
}

func escapeGoModulePart(value string) string {
	var result strings.Builder
	for _, char := range value {
		switch {
		case char == '!':
			result.WriteString("!!")
		case char >= 'A' && char <= 'Z':
			result.WriteByte('!')
			result.WriteRune(char + ('a' - 'A'))
		default:
			result.WriteRune(char)
		}
	}
	return result.String()
}

var versionTokenPattern = regexp.MustCompile(`[0-9]+|[A-Za-z]+`)

func latestStableVersion(versions []string) (string, bool) {
	latest := ""
	for _, candidate := range versions {
		candidate = strings.TrimSpace(candidate)
		if !stableVersion(candidate) {
			continue
		}
		if latest == "" || compareVersions(candidate, latest) > 0 {
			latest = candidate
		}
	}
	return latest, latest != ""
}

func stableVersion(version string) bool {
	if strings.TrimSpace(version) == "" {
		return false
	}
	lower := strings.ToLower(version)
	for _, marker := range []string{"alpha", "beta", "-rc", ".rc", "snapshot", "preview", "canary", "nightly", "-dev", ".dev", "-m", "milestone", "eap"} {
		if strings.Contains(lower, marker) {
			return false
		}
	}
	return true
}

func compareVersions(left string, right string) int {
	leftTokens := versionTokenPattern.FindAllString(strings.TrimPrefix(strings.SplitN(left, "+", 2)[0], "v"), -1)
	rightTokens := versionTokenPattern.FindAllString(strings.TrimPrefix(strings.SplitN(right, "+", 2)[0], "v"), -1)
	length := len(leftTokens)
	if len(rightTokens) > length {
		length = len(rightTokens)
	}
	for index := 0; index < length; index++ {
		leftToken := "0"
		rightToken := "0"
		if index < len(leftTokens) {
			leftToken = leftTokens[index]
		}
		if index < len(rightTokens) {
			rightToken = rightTokens[index]
		}
		comparison := compareVersionToken(leftToken, rightToken)
		if comparison != 0 {
			return comparison
		}
	}
	return strings.Compare(left, right)
}

func compareVersionToken(left string, right string) int {
	leftNumeric := left != "" && left[0] >= '0' && left[0] <= '9'
	rightNumeric := right != "" && right[0] >= '0' && right[0] <= '9'
	if leftNumeric && rightNumeric {
		left = strings.TrimLeft(left, "0")
		right = strings.TrimLeft(right, "0")
		if left == "" {
			left = "0"
		}
		if right == "" {
			right = "0"
		}
		if len(left) != len(right) {
			if len(left) > len(right) {
				return 1
			}
			return -1
		}
		return strings.Compare(left, right)
	}
	if leftNumeric != rightNumeric {
		if leftNumeric {
			return 1
		}
		return -1
	}
	return strings.Compare(strings.ToLower(left), strings.ToLower(right))
}
