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
	SourceTypeGitHub    = "github"
	SourceTypeGitLab    = "gitlab"
	SourceTypeGitea     = "gitea"
	SourceTypeBitbucket = "bitbucket"
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
	case SourceTypeGitea:
		return GiteaClient{HTTPClient: httpClient}, nil
	case SourceTypeBitbucket:
		return BitbucketClient{HTTPClient: httpClient}, nil
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

func (c GitHubClient) HasBranch(ctx context.Context, source Source, project Project, branch string) (bool, error) {
	requestURL := githubAPIURL(source, "repos", project.ProviderID, "branches", branch)
	return branchExists(c.HTTPClient, func() (*http.Request, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL.String(), nil)
		if err != nil {
			return nil, fmt.Errorf("create github request: %w", err)
		}
		req.Header.Set("Accept", "application/vnd.github+json")
		if source.PATToken != "" {
			req.Header.Set("Authorization", "Bearer "+source.PATToken)
		}
		return req, nil
	})
}

func (c GitHubClient) CompareBranches(ctx context.Context, source Source, project Project, baseBranch string, headBranch string) (BranchDivergence, error) {
	var raw struct {
		AheadBy  int `json:"ahead_by"`
		BehindBy int `json:"behind_by"`
	}
	if err := c.getJSON(ctx, source, githubAPIURL(source, "repos", project.ProviderID, "compare", baseBranch+"..."+headBranch), &raw); err != nil {
		return BranchDivergence{}, err
	}
	return BranchDivergence{Ahead: raw.AheadBy, Behind: raw.BehindBy}, nil
}

func (c GitHubClient) DefaultBranch(ctx context.Context, source Source, project Project) (string, error) {
	var repo struct {
		DefaultBranch string `json:"default_branch"`
	}
	if err := c.getJSON(ctx, source, githubAPIURL(source, "repos", project.ProviderID), &repo); err != nil {
		return "", err
	}
	if repo.DefaultBranch == "" {
		return "", fmt.Errorf("github default branch is empty for %s", project.ProviderID)
	}
	return repo.DefaultBranch, nil
}

func (c GitHubClient) ProtectedBranches(context.Context, Source, Project) ([]ProtectedBranch, error) {
	return nil, fmt.Errorf("protected branches check is supported only for gitlab")
}

func (c GitHubClient) Tags(context.Context, Source, Project) ([]Tag, error) {
	return []Tag{}, nil
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
	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrFileNotFound
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("gitlab request failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read gitlab file: %w", err)
	}

	return body, nil
}

func (c GitLabClient) HasBranch(ctx context.Context, source Source, project Project, branch string) (bool, error) {
	requestURL := gitlabProjectAPIURL(source, project.ProviderID, "repository", "branches", branch)
	return branchExists(c.HTTPClient, func() (*http.Request, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL.String(), nil)
		if err != nil {
			return nil, fmt.Errorf("create gitlab request: %w", err)
		}
		if source.PATToken != "" {
			req.Header.Set("PRIVATE-TOKEN", source.PATToken)
		}
		return req, nil
	})
}

func (c GitLabClient) CompareBranches(ctx context.Context, source Source, project Project, baseBranch string, headBranch string) (BranchDivergence, error) {
	ahead, err := c.compareBranchCommitCount(ctx, source, project, baseBranch, headBranch)
	if err != nil {
		return BranchDivergence{}, err
	}
	behind, err := c.compareBranchCommitCount(ctx, source, project, headBranch, baseBranch)
	if err != nil {
		return BranchDivergence{}, err
	}
	return BranchDivergence{Ahead: ahead, Behind: behind}, nil
}

func (c GitLabClient) compareBranchCommitCount(ctx context.Context, source Source, project Project, fromBranch string, toBranch string) (int, error) {
	requestURL := gitlabProjectAPIURL(source, project.ProviderID, "repository", "compare")
	values := requestURL.Query()
	values.Set("from", fromBranch)
	values.Set("to", toBranch)
	values.Set("per_page", "100")
	requestURL.RawQuery = values.Encode()

	var raw struct {
		Commits []struct{} `json:"commits"`
	}
	if err := c.getJSON(ctx, source, requestURL, &raw); err != nil {
		return 0, err
	}
	return len(raw.Commits), nil
}

func (c GitLabClient) DefaultBranch(ctx context.Context, source Source, project Project) (string, error) {
	var repo struct {
		DefaultBranch string `json:"default_branch"`
	}
	if err := c.getJSON(ctx, source, gitlabProjectAPIURL(source, project.ProviderID), &repo); err != nil {
		return "", err
	}
	if repo.DefaultBranch == "" {
		return "", fmt.Errorf("gitlab default branch is empty for %s", project.ProviderID)
	}
	return repo.DefaultBranch, nil
}

func (c GitLabClient) Tags(ctx context.Context, source Source, project Project) ([]Tag, error) {
	requestURL := gitlabProjectAPIURL(source, project.ProviderID, "repository", "tags")
	values := requestURL.Query()
	values.Set("per_page", "100")
	requestURL.RawQuery = values.Encode()

	var raw []struct {
		Name   string `json:"name"`
		Commit struct {
			CreatedAt time.Time `json:"created_at"`
		} `json:"commit"`
	}
	if err := c.getJSON(ctx, source, requestURL, &raw); err != nil {
		return nil, err
	}

	tags := make([]Tag, 0, len(raw))
	for _, item := range raw {
		tags = append(tags, Tag{Name: item.Name, CreatedAt: item.Commit.CreatedAt})
	}

	return tags, nil
}

func (c GitLabClient) ProtectedBranches(ctx context.Context, source Source, project Project) ([]ProtectedBranch, error) {
	var raw []struct {
		Name             string `json:"name"`
		AllowForcePush   bool   `json:"allow_force_push"`
		PushAccessLevels []struct {
			AccessLevel int `json:"access_level"`
		} `json:"push_access_levels"`
		MergeAccessLevels []struct {
			AccessLevel int `json:"access_level"`
		} `json:"merge_access_levels"`
	}
	if err := c.getJSON(ctx, source, gitlabProjectAPIURL(source, project.ProviderID, "protected_branches"), &raw); err != nil {
		return nil, err
	}

	branches := make([]ProtectedBranch, 0, len(raw))
	for _, item := range raw {
		branch := ProtectedBranch{Name: item.Name, AllowForcePush: item.AllowForcePush}
		for _, level := range item.PushAccessLevels {
			branch.PushAccessLevels = append(branch.PushAccessLevels, level.AccessLevel)
		}
		for _, level := range item.MergeAccessLevels {
			branch.MergeAccessLevels = append(branch.MergeAccessLevels, level.AccessLevel)
		}
		branches = append(branches, branch)
	}

	return branches, nil
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

type GiteaClient struct {
	HTTPClient HTTPDoer
}

func (c GiteaClient) ResolveHead(ctx context.Context, source Source, project Project) (Commit, error) {
	defaultBranch, err := c.DefaultBranch(ctx, source, project)
	if err != nil {
		return Commit{}, err
	}

	var commit struct {
		SHA string `json:"sha"`
	}
	if err := c.getJSON(ctx, source, giteaAPIURL(source, "repos", project.ProviderID, "commits", defaultBranch), &commit); err != nil {
		return Commit{}, err
	}
	if commit.SHA == "" {
		return Commit{}, fmt.Errorf("gitea head commit is empty for %s", project.ProviderID)
	}

	return Commit{SHA: commit.SHA, ShortSHA: shortSHA(commit.SHA)}, nil
}

func (c GiteaClient) FetchFile(ctx context.Context, source Source, project Project, commitSHA string, filePath string) ([]byte, error) {
	requestURL := giteaAPIURL(source, "repos", project.ProviderID, "contents", filePath)
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
		return nil, fmt.Errorf("unsupported gitea file encoding %q for %s", response.Encoding, filePath)
	}

	content, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(response.Content, "\n", ""))
	if err != nil {
		return nil, fmt.Errorf("decode gitea file %s: %w", filePath, err)
	}

	return content, nil
}

func (c GiteaClient) HasBranch(ctx context.Context, source Source, project Project, branch string) (bool, error) {
	requestURL := giteaAPIURL(source, "repos", project.ProviderID, "branches", branch)
	return branchExists(c.HTTPClient, func() (*http.Request, error) {
		return c.newRequest(ctx, source, requestURL)
	})
}

func (c GiteaClient) CompareBranches(ctx context.Context, source Source, project Project, baseBranch string, headBranch string) (BranchDivergence, error) {
	var raw struct {
		AheadBy  int `json:"ahead_by"`
		BehindBy int `json:"behind_by"`
	}
	if err := c.getJSON(ctx, source, giteaAPIURL(source, "repos", project.ProviderID, "compare", baseBranch+"..."+headBranch), &raw); err != nil {
		return BranchDivergence{}, err
	}
	return BranchDivergence{Ahead: raw.AheadBy, Behind: raw.BehindBy}, nil
}

func (c GiteaClient) DefaultBranch(ctx context.Context, source Source, project Project) (string, error) {
	var repo struct {
		DefaultBranch string `json:"default_branch"`
	}
	if err := c.getJSON(ctx, source, giteaAPIURL(source, "repos", project.ProviderID), &repo); err != nil {
		return "", err
	}
	if repo.DefaultBranch == "" {
		return "", fmt.Errorf("gitea default branch is empty for %s", project.ProviderID)
	}
	return repo.DefaultBranch, nil
}

func (c GiteaClient) ProtectedBranches(context.Context, Source, Project) ([]ProtectedBranch, error) {
	return nil, fmt.Errorf("protected branches check is supported only for gitlab")
}

func (c GiteaClient) Tags(ctx context.Context, source Source, project Project) ([]Tag, error) {
	requestURL := giteaAPIURL(source, "repos", project.ProviderID, "tags")
	values := requestURL.Query()
	values.Set("limit", "100")
	requestURL.RawQuery = values.Encode()

	var raw []struct {
		Name   string `json:"name"`
		Commit struct {
			Created time.Time `json:"created"`
		} `json:"commit"`
	}
	if err := c.getJSON(ctx, source, requestURL, &raw); err != nil {
		return nil, err
	}

	tags := make([]Tag, 0, len(raw))
	for _, item := range raw {
		tags = append(tags, Tag{Name: item.Name, CreatedAt: item.Commit.Created})
	}

	return tags, nil
}

func (c GiteaClient) getJSON(ctx context.Context, source Source, requestURL url.URL, target any) error {
	req, err := c.newRequest(ctx, source, requestURL)
	if err != nil {
		return err
	}
	return doJSON(c.HTTPClient, req, target)
}

func (c GiteaClient) newRequest(ctx context.Context, source Source, requestURL url.URL) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create gitea request: %w", err)
	}
	if source.PATToken != "" {
		req.Header.Set("Authorization", "token "+source.PATToken)
	}
	return req, nil
}

type BitbucketClient struct {
	HTTPClient HTTPDoer
}

func (c BitbucketClient) ResolveHead(ctx context.Context, source Source, project Project) (Commit, error) {
	defaultBranch, err := c.DefaultBranch(ctx, source, project)
	if err != nil {
		return Commit{}, err
	}

	var commit struct {
		Hash string `json:"hash"`
	}
	if err := c.getJSON(ctx, source, bitbucketRepoAPIURL(source, project.ProviderID, "commit", defaultBranch), &commit); err != nil {
		return Commit{}, err
	}
	if commit.Hash == "" {
		return Commit{}, fmt.Errorf("bitbucket head commit is empty for %s", project.ProviderID)
	}

	return Commit{SHA: commit.Hash, ShortSHA: shortSHA(commit.Hash)}, nil
}

func (c BitbucketClient) FetchFile(ctx context.Context, source Source, project Project, commitSHA string, filePath string) ([]byte, error) {
	requestURL := bitbucketRepoAPIURL(source, project.ProviderID, "src", commitSHA, filePath)

	req, err := c.newRequest(ctx, source, requestURL)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("bitbucket request failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrFileNotFound
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("bitbucket request failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read bitbucket file: %w", err)
	}

	return body, nil
}

func (c BitbucketClient) HasBranch(ctx context.Context, source Source, project Project, branch string) (bool, error) {
	requestURL := bitbucketRepoAPIURL(source, project.ProviderID, "refs", "branches", branch)
	return branchExists(c.HTTPClient, func() (*http.Request, error) {
		return c.newRequest(ctx, source, requestURL)
	})
}

func (c BitbucketClient) CompareBranches(ctx context.Context, source Source, project Project, baseBranch string, headBranch string) (BranchDivergence, error) {
	ahead, err := c.bitbucketBranchCommitCount(ctx, source, project, headBranch, baseBranch)
	if err != nil {
		return BranchDivergence{}, err
	}
	behind, err := c.bitbucketBranchCommitCount(ctx, source, project, baseBranch, headBranch)
	if err != nil {
		return BranchDivergence{}, err
	}
	return BranchDivergence{Ahead: ahead, Behind: behind}, nil
}

func (c BitbucketClient) bitbucketBranchCommitCount(ctx context.Context, source Source, project Project, includeBranch string, excludeBranch string) (int, error) {
	requestURL := bitbucketRepoAPIURL(source, project.ProviderID, "commits", includeBranch)
	values := requestURL.Query()
	values.Set("exclude", excludeBranch)
	values.Set("pagelen", "100")
	requestURL.RawQuery = values.Encode()

	total := 0
	for {
		var raw struct {
			Values []struct{} `json:"values"`
			Next   string     `json:"next"`
		}
		if err := c.getJSON(ctx, source, requestURL, &raw); err != nil {
			return 0, err
		}
		total += len(raw.Values)
		if raw.Next == "" {
			return total, nil
		}
		nextURL, err := url.Parse(raw.Next)
		if err != nil {
			return 0, fmt.Errorf("parse bitbucket next page: %w", err)
		}
		requestURL = *nextURL
	}
}

func (c BitbucketClient) DefaultBranch(ctx context.Context, source Source, project Project) (string, error) {
	var repo struct {
		MainBranch struct {
			Name string `json:"name"`
		} `json:"mainbranch"`
	}
	if err := c.getJSON(ctx, source, bitbucketRepoAPIURL(source, project.ProviderID), &repo); err != nil {
		return "", err
	}
	if repo.MainBranch.Name == "" {
		return "", fmt.Errorf("bitbucket default branch is empty for %s", project.ProviderID)
	}
	return repo.MainBranch.Name, nil
}

func (c BitbucketClient) ProtectedBranches(context.Context, Source, Project) ([]ProtectedBranch, error) {
	return nil, fmt.Errorf("protected branches check is supported only for gitlab")
}

func (c BitbucketClient) Tags(ctx context.Context, source Source, project Project) ([]Tag, error) {
	requestURL := bitbucketRepoAPIURL(source, project.ProviderID, "refs", "tags")
	values := requestURL.Query()
	values.Set("pagelen", "100")
	requestURL.RawQuery = values.Encode()

	var raw struct {
		Values []struct {
			Name   string `json:"name"`
			Target struct {
				Date time.Time `json:"date"`
			} `json:"target"`
		} `json:"values"`
	}
	if err := c.getJSON(ctx, source, requestURL, &raw); err != nil {
		return nil, err
	}

	tags := make([]Tag, 0, len(raw.Values))
	for _, item := range raw.Values {
		tags = append(tags, Tag{Name: item.Name, CreatedAt: item.Target.Date})
	}

	return tags, nil
}

func (c BitbucketClient) getJSON(ctx context.Context, source Source, requestURL url.URL, target any) error {
	req, err := c.newRequest(ctx, source, requestURL)
	if err != nil {
		return err
	}
	return doJSON(c.HTTPClient, req, target)
}

func (c BitbucketClient) newRequest(ctx context.Context, source Source, requestURL url.URL) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create bitbucket request: %w", err)
	}
	if source.PATToken != "" {
		req.Header.Set("Authorization", "Bearer "+source.PATToken)
	}
	return req, nil
}

func branchExists(httpClient HTTPDoer, build func() (*http.Request, error)) (bool, error) {
	req, err := build()
	if err != nil {
		return false, err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return true, nil
	}

	return false, fmt.Errorf("request failed with status %d", resp.StatusCode)
}

func doJSON(httpClient HTTPDoer, req *http.Request, target any) error {
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode == http.StatusNotFound {
		return ErrFileNotFound
	}
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

func giteaAPIURL(source Source, elements ...string) url.URL {
	base := parseBaseURL(source.URL)
	if base.Host == "" {
		base = url.URL{Scheme: "https", Host: "gitea.com"}
	}
	base.Path = path.Join(base.Path, "api", "v1")

	return joinURL(base, elements...)
}

func bitbucketAPIBase(source Source) url.URL {
	base := parseBaseURL(source.URL)
	if base.Host == "" || strings.EqualFold(base.Host, "bitbucket.org") {
		base = url.URL{Scheme: "https", Host: "api.bitbucket.org"}
	}
	base.Path = path.Join(base.Path, "2.0")

	return base
}

func bitbucketRepoAPIURL(source Source, projectID string, elements ...string) url.URL {
	base := bitbucketAPIBase(source)
	parts := strings.SplitN(strings.TrimSpace(projectID), "/", 2)
	if len(parts) != 2 {
		return joinURL(base, append([]string{"repositories", projectID}, elements...)...)
	}
	rawElements := append([]string{"repositories", parts[0], parts[1]}, elements...)
	return joinURL(base, rawElements...)
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
