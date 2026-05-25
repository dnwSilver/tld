package projectsync

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

const (
	SourceTypeGitHub = "github"
	SourceTypeGitLab = "gitlab"
)

type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

func NewSourceClient(sourceType string, httpClient HTTPDoer) (SourceClient, error) {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 20 * time.Second}
	}

	switch strings.ToLower(strings.TrimSpace(sourceType)) {
	case SourceTypeGitHub:
		return GitHubClient{HTTPClient: httpClient}, nil
	case SourceTypeGitLab:
		return GitLabClient{HTTPClient: httpClient}, nil
	default:
		return nil, fmt.Errorf("unsupported source type %q", sourceType)
	}
}

type GitHubClient struct {
	HTTPClient HTTPDoer
}

type GitLabClient struct {
	HTTPClient HTTPDoer
}

func (c GitHubClient) ResolveHead(ctx context.Context, source Source, project Project) (Commit, error) {
	var repo struct {
		DefaultBranch string `json:"default_branch"`
	}
	if err := c.getJSON(ctx, source, githubAPIURL(source, "repos", project.ProviderID), &repo); err != nil {
		return Commit{}, err
	}
	if repo.DefaultBranch == "" {
		return Commit{}, fmt.Errorf("github default branch is empty for %s", project.ProviderID)
	}

	var commit struct {
		SHA string `json:"sha"`
	}
	if err := c.getJSON(ctx, source, githubAPIURL(source, "repos", project.ProviderID, "commits", repo.DefaultBranch), &commit); err != nil {
		return Commit{}, err
	}
	if commit.SHA == "" {
		return Commit{}, fmt.Errorf("github head commit is empty for %s", project.ProviderID)
	}

	return Commit{SHA: commit.SHA, ShortSHA: shortSHA(commit.SHA)}, nil
}

func (c GitHubClient) FetchFile(ctx context.Context, source Source, project Project, commitSHA string, filePath string) ([]byte, error) {
	requestURL := githubAPIURL(source, "repos", project.ProviderID, "contents", filePath)
	values := requestURL.Query()
	values.Set("ref", commitSHA)
	requestURL.RawQuery = values.Encode()

	var response struct {
		Content  string `json:"content"`
		Encoding string `json:"encoding"`
	}
	if err := c.getJSON(ctx, source, requestURL, &response); err != nil {
		return nil, err
	}
	if response.Encoding != "base64" {
		return nil, fmt.Errorf("unsupported github file encoding %q for %s", response.Encoding, filePath)
	}

	content, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(response.Content, "\n", ""))
	if err != nil {
		return nil, fmt.Errorf("decode github file %s: %w", filePath, err)
	}

	return content, nil
}

func (c GitHubClient) getJSON(ctx context.Context, source Source, requestURL url.URL, target any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL.String(), nil)
	if err != nil {
		return fmt.Errorf("create github request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	if source.PATToken != "" {
		req.Header.Set("Authorization", "Bearer "+source.PATToken)
	}

	return doJSON(c.HTTPClient, req, target)
}

func (c GitLabClient) ResolveHead(ctx context.Context, source Source, project Project) (Commit, error) {
	var repo struct {
		DefaultBranch string `json:"default_branch"`
	}
	if err := c.getJSON(ctx, source, gitlabProjectAPIURL(source, project.ProviderID), &repo); err != nil {
		return Commit{}, err
	}
	if repo.DefaultBranch == "" {
		return Commit{}, fmt.Errorf("gitlab default branch is empty for %s", project.ProviderID)
	}

	var commit struct {
		ID string `json:"id"`
	}
	if err := c.getJSON(ctx, source, gitlabProjectAPIURL(source, project.ProviderID, "repository", "commits", repo.DefaultBranch), &commit); err != nil {
		return Commit{}, err
	}
	if commit.ID == "" {
		return Commit{}, fmt.Errorf("gitlab head commit is empty for %s", project.ProviderID)
	}

	return Commit{SHA: commit.ID, ShortSHA: shortSHA(commit.ID)}, nil
}

func (c GitLabClient) FetchFile(ctx context.Context, source Source, project Project, commitSHA string, filePath string) ([]byte, error) {
	requestURL := gitlabProjectAPIURL(source, project.ProviderID, "repository", "files", filePath, "raw")
	values := requestURL.Query()
	values.Set("ref", commitSHA)
	requestURL.RawQuery = values.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create gitlab request: %w", err)
	}
	if source.PATToken != "" {
		req.Header.Set("PRIVATE-TOKEN", source.PATToken)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gitlab request failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("gitlab request failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read gitlab file: %w", err)
	}

	return body, nil
}

func (c GitLabClient) getJSON(ctx context.Context, source Source, requestURL url.URL, target any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL.String(), nil)
	if err != nil {
		return fmt.Errorf("create gitlab request: %w", err)
	}
	if source.PATToken != "" {
		req.Header.Set("PRIVATE-TOKEN", source.PATToken)
	}

	return doJSON(c.HTTPClient, req, target)
}

func doJSON(httpClient HTTPDoer, req *http.Request, target any) error {
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
}

func githubAPIURL(source Source, elements ...string) url.URL {
	base := parseBaseURL(source.URL)
	if base.Host == "" || strings.EqualFold(base.Host, "github.com") {
		base = url.URL{Scheme: "https", Host: "api.github.com"}
	} else {
		base.Path = path.Join(base.Path, "api", "v3")
	}

	return joinURL(base, elements...)
}

func gitlabAPIURL(source Source, elements ...string) url.URL {
	base := gitlabAPIBase(source)
	return joinURL(base, elements...)
}

func gitlabProjectAPIURL(source Source, projectID string, elements ...string) url.URL {
	base := gitlabAPIBase(source)
	basePath := base.Path
	baseRawPath := base.EscapedPath()
	rawElements := append([]string{"projects", projectID}, elements...)
	escapedElements := []string{"projects", url.PathEscape(projectID)}
	for _, element := range elements {
		escapedElements = append(escapedElements, url.PathEscape(element))
	}

	base.Path = path.Join(append([]string{basePath}, rawElements...)...)
	base.RawPath = path.Join(append([]string{baseRawPath}, escapedElements...)...)

	return base
}

func gitlabAPIBase(source Source) url.URL {
	base := parseBaseURL(source.URL)
	if base.Host == "" {
		base = url.URL{Scheme: "https", Host: "gitlab.com"}
	}
	base.Path = path.Join(base.Path, "api", "v4")

	return base
}

func parseBaseURL(raw string) url.URL {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Host == "" {
		return url.URL{}
	}

	return *parsed
}

func joinURL(base url.URL, elements ...string) url.URL {
	for _, element := range elements {
		base.Path = path.Join(base.Path, element)
	}

	return base
}
