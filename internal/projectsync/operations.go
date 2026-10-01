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
	ID    string
	Title string
}

var ProjectOperations = []OperationDefinition{
	{ID: operationShareCICache, Title: "Share CI cache with all branches"},
	{ID: operationProcessModeOldestFirst, Title: "Process RC queue oldest first"},
	{ID: operationProtectBranches, Title: "Protect master/dev branches"},
	{ID: operationProtectTags, Title: "Protect release/hotfix/version tags"},
}

type OperationService struct {
	SourceClient SourceClient
}

func (s OperationService) Run(ctx context.Context, source Source, project Project, operationID string) error {
	if s.SourceClient == nil {
		return errors.New("operation source client is empty")
	}
	if project.ProviderID == "" {
		return errors.New("operation provider project id is empty")
	}
	if operationID == operationProtectBranches || operationID == operationProtectTags {
		return s.ensureProtection(ctx, source, project, operationID)
	}
	client, ok := s.SourceClient.(ProjectOperationSourceClient)
	if !ok {
		return errors.New("project operations are supported only for gitlab")
	}

	switch operationID {
	case operationShareCICache:
		return client.SetSeparatedCaches(ctx, source, project, false)
	case operationProcessModeOldestFirst:
		resourceGroup, err := releaseCandidateResourceGroup(ctx, client, source, project)
		if err != nil {
			return err
		}
		return client.SetResourceGroupProcessMode(ctx, source, project, resourceGroup, processModeOldestFirst)
	default:
		return fmt.Errorf("unsupported operation %q", operationID)
	}
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
