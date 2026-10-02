package projectsync

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const (
	checkProtectedBranches   = "protectedBranches"
	checkProtectedTags       = "protectedTags"
	operationProtectBranches = "protectBranches"
	operationProtectTags     = "protectTags"
)

var requiredProtectedBranches = []string{"master", "dev"}
var requiredProtectedTags = []string{"release/*.*.*", "hotfix/*.*.*", "v*.*.*"}

type ProtectionSourceClient interface {
	ProtectedTags(context.Context, Source, Project) ([]string, error)
	ProtectBranch(context.Context, Source, Project, string) error
	ProtectTag(context.Context, Source, Project, string) error
}

func (s CheckService) runProtectionCheck(ctx context.Context, source Source, project Project, checkID string) (CheckState, error) {
	if !strings.EqualFold(strings.TrimSpace(source.Type), SourceTypeGitLab) {
		return CheckStateNotApplicable, nil
	}
	names, required, err := protectionNames(ctx, s.SourceClient, source, project, checkID == checkProtectedTags)
	if err != nil {
		return CheckStateFail, err
	}
	return checkState(len(missingProtection(names, required)) == 0), nil
}

func protectionNames(ctx context.Context, client SourceClient, source Source, project Project, tags bool) ([]string, []string, error) {
	if tags {
		c, ok := client.(ProtectionSourceClient)
		if !ok {
			return nil, nil, fmt.Errorf("protected tags source client is not supported")
		}
		names, err := c.ProtectedTags(ctx, source, project)
		return names, requiredProtectedTags, err
	}
	branches, err := client.ProtectedBranches(ctx, source, project)
	names := make([]string, 0, len(branches))
	for _, branch := range branches {
		names = append(names, branch.Name)
	}
	return names, requiredProtectedBranches, err
}

func missingProtection(names, required []string) []string {
	existing := make(map[string]bool, len(names))
	for _, name := range names {
		existing[name] = true
	}
	var missing []string
	for _, name := range required {
		if !existing[name] {
			missing = append(missing, name)
		}
	}
	return missing
}

func (s OperationService) ensureProtection(ctx context.Context, source Source, project Project, operationID string) error {
	_, err := s.ensureProtectionResult(ctx, source, project, operationID)
	return err
}

func (s OperationService) ensureProtectionResult(ctx context.Context, source Source, project Project, operationID string) (OperationResult, error) {
	if !strings.EqualFold(strings.TrimSpace(source.Type), SourceTypeGitLab) {
		return OperationResult{Outcome: OperationOutcomeFailed}, fmt.Errorf("project operations are supported only for gitlab")
	}
	client, ok := s.SourceClient.(ProtectionSourceClient)
	if !ok {
		return OperationResult{Outcome: OperationOutcomeFailed}, fmt.Errorf("protection source client is not supported")
	}
	tags := operationID == operationProtectTags
	names, required, err := protectionNames(ctx, s.SourceClient, source, project, tags)
	if err != nil {
		return OperationResult{Outcome: OperationOutcomeFailed}, err
	}
	missing := missingProtection(names, required)
	result := OperationResult{Outcome: OperationOutcomeApplied, Total: len(missing)}
	for _, name := range missing {
		if tags {
			err = client.ProtectTag(ctx, source, project, name)
		} else {
			err = client.ProtectBranch(ctx, source, project, name)
		}
		if err != nil {
			result.Outcome = OperationOutcomeFailed
			if result.Applied > 0 {
				result.Outcome = OperationOutcomePartial
			}
			return result, fmt.Errorf("protect %s after %d successful changes: %w", name, result.Applied, err)
		}
		result.Applied++
	}
	return result, nil
}

func (c GitLabClient) ProtectedTags(ctx context.Context, source Source, project Project) ([]string, error) {
	var names []string
	for page := 1; page <= maxTagPages; page++ {
		requestURL := gitlabProjectAPIURL(source, project.ProviderID, "protected_tags")
		values := requestURL.Query()
		values.Set("per_page", "100")
		values.Set("page", fmt.Sprint(page))
		requestURL.RawQuery = values.Encode()
		var raw []struct {
			Name string `json:"name"`
		}
		if err := c.getJSON(ctx, source, requestURL, &raw); err != nil {
			return nil, err
		}
		for _, item := range raw {
			names = append(names, item.Name)
		}
		if len(raw) < 100 {
			return names, nil
		}
	}
	return nil, fmt.Errorf("gitlab protected tags exceed page budget")
}

func (c GitLabClient) ProtectBranch(ctx context.Context, source Source, project Project, name string) error {
	rule := protectedBranchRules[name]
	values := url.Values{}
	values.Set("name", name)
	values.Set("push_access_level", fmt.Sprint(rule.PushAccessLevels[0]))
	values.Set("merge_access_level", fmt.Sprint(rule.MergeAccessLevels[0]))
	values.Set("allow_force_push", "false")
	return c.writeForm(ctx, source, http.MethodPost, gitlabProjectAPIURL(source, project.ProviderID, "protected_branches"), values)
}

func (c GitLabClient) ProtectTag(ctx context.Context, source Source, project Project, name string) error {
	values := url.Values{}
	values.Set("name", name)
	values.Set("create_access_level", fmt.Sprint(accessLevelMaintainer))
	return c.writeForm(ctx, source, http.MethodPost, gitlabProjectAPIURL(source, project.ProviderID, "protected_tags"), values)
}
