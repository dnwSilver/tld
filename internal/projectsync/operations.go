package projectsync

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	operationShareCICache           = "shareCICache"
	operationProcessModeOldestFirst = "processModeOldestFirst"
)

// releaseCandidateResourceGroupPrefix mirrors the application-rc-$CI_PROJECT_ID
// resource group created by the shared ci-react-site component.
const releaseCandidateResourceGroupPrefix = "application-rc-"

type OperationDefinition struct {
	ID     string
	Title  string
	Change string
}

var ProjectOperations = []OperationDefinition{
	{ID: operationShareCICache, Title: "Share CI cache with all branches", Change: "separated_caches = false (cross-branch cache sharing)"},
	{ID: operationProcessModeOldestFirst, Title: "Process RC queue oldest first", Change: "resource_group.process_mode = oldest_first"},
	{ID: operationProtectBranches, Title: "Protect master/dev branches", Change: "create missing protected branch rules"},
	{ID: operationProtectTags, Title: "Protect release/hotfix/version tags", Change: "create missing protected tag rules"},
}

type OperationService struct {
	SourceClient SourceClient
}

type OperationOutcome string

const (
	OperationOutcomeApplied   OperationOutcome = "applied"
	OperationOutcomePartial   OperationOutcome = "partial"
	OperationOutcomeUncertain OperationOutcome = "uncertain"
	OperationOutcomeFailed    OperationOutcome = "failed"
)

type OperationResult struct {
	Outcome OperationOutcome
	Applied int
	Total   int
}

func (s OperationService) Run(ctx context.Context, source Source, project Project, operationID string) error {
	_, err := s.run(ctx, source, project, operationID)
	return err
}

func (s OperationService) run(ctx context.Context, source Source, project Project, operationID string) (OperationResult, error) {
	if s.SourceClient == nil {
		return OperationResult{Outcome: OperationOutcomeFailed}, errors.New("operation source client is empty")
	}
	if project.ProviderID == "" {
		return OperationResult{Outcome: OperationOutcomeFailed}, errors.New("operation provider project id is empty")
	}
	if operationID == operationProtectBranches || operationID == operationProtectTags {
		return s.ensureProtectionResult(ctx, source, project, operationID)
	}
	client, ok := s.SourceClient.(ProjectOperationSourceClient)
	if !ok {
		return OperationResult{Outcome: OperationOutcomeFailed}, errors.New("project operations are supported only for gitlab")
	}

	result := OperationResult{Outcome: OperationOutcomeApplied, Applied: 1, Total: 1}
	switch operationID {
	case operationShareCICache:
		if err := client.SetSeparatedCaches(ctx, source, project, false); err != nil {
			result.Outcome, result.Applied = OperationOutcomeFailed, 0
			return result, err
		}
		return result, nil
	case operationProcessModeOldestFirst:
		resourceGroup, err := releaseCandidateResourceGroup(ctx, client, source, project)
		if err != nil {
			result.Outcome, result.Applied = OperationOutcomeFailed, 0
			return result, err
		}
		if err := client.SetResourceGroupProcessMode(ctx, source, project, resourceGroup, processModeOldestFirst); err != nil {
			result.Outcome, result.Applied = OperationOutcomeFailed, 0
			return result, err
		}
		return result, nil
	default:
		return OperationResult{Outcome: OperationOutcomeFailed}, fmt.Errorf("unsupported operation %q", operationID)
	}
}

// RunAndVerify reads the remote state back before reporting success to the UI.
// A failed readback is reported as an uncertain outcome because the mutation
// may already have been applied.
func (s OperationService) RunAndVerify(ctx context.Context, source Source, project Project, operationID string) (OperationResult, error) {
	result, runErr := s.run(ctx, source, project, operationID)
	if runErr != nil {
		return result, runErr
	}
	check := CheckService{SourceClient: s.SourceClient}
	var state CheckState
	var err error
	switch operationID {
	case operationShareCICache:
		client, ok := s.SourceClient.(CISettingsSourceClient)
		if !ok {
			result.Outcome = OperationOutcomeUncertain
			return result, errors.New("remote change applied, but CI settings readback is unavailable")
		}
		settings, readErr := client.CISettings(ctx, source, project)
		err = readErr
		state = checkState(!settings.SeparatedCaches)
	case operationProcessModeOldestFirst:
		state, err = check.runProcessModeCheck(ctx, source, project)
	case operationProtectBranches:
		branches, readErr := s.SourceClient.ProtectedBranches(ctx, source, project)
		err = readErr
		state = checkState(protectedBranchesValid(branches))
	case operationProtectTags:
		state, err = check.runProtectionCheck(ctx, source, project, checkProtectedTags)
	default:
		result.Outcome = OperationOutcomeUncertain
		return result, fmt.Errorf("unsupported operation %q", operationID)
	}
	if err != nil {
		result.Outcome = OperationOutcomeUncertain
		return result, fmt.Errorf("remote change applied, but readback failed: %w", err)
	}
	if state != CheckStatePass {
		result.Outcome = OperationOutcomeUncertain
		return result, fmt.Errorf("remote change applied, but readback did not confirm %s", operationID)
	}
	result.Outcome = OperationOutcomeApplied
	return result, nil
}

// The resource group name always contains the numeric project id, even when
// the provider id is configured as a group/repo path.
func releaseCandidateResourceGroup(ctx context.Context, client NumericProjectIDSourceClient, source Source, project Project) (string, error) {
	providerID := strings.TrimSpace(project.ProviderID)
	if numericID, err := strconv.ParseInt(providerID, 10, 64); err == nil {
		return releaseCandidateResourceGroupPrefix + strconv.FormatInt(numericID, 10), nil
	}

	numericID, err := client.NumericProjectID(ctx, source, project)
	if err != nil {
		return "", fmt.Errorf("resolve numeric project id: %w", err)
	}

	return releaseCandidateResourceGroupPrefix + strconv.FormatInt(numericID, 10), nil
}
