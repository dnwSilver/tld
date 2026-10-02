package ui

import "github.com/dnwSilver/tld/internal/ui/uikit"

// RenderState groups the data needed by each screen. The renderer reads only
// the active screen's state, so callers do not depend on a positional list of
// dozens of unrelated values.
type RenderState struct {
	Width           int
	Height          int
	Screen          uikit.Screen
	Dashboard       DashboardRenderState
	Stacks          StacksRenderState
	Namespaces      NamespacesRenderState
	Dependencies    DependenciesRenderState
	Projects        ProjectsRenderState
	Sources         SourcesRenderState
	Policies        PoliciesRenderState
	View            DependencyViewRenderState
	Settings        SettingsRenderState
	Releases        ReleasesRenderState
	Vulnerabilities VulnerabilitiesRenderState
	Overlay         OverlayRenderState
}

type DashboardRenderState struct {
	TokenRights       uikit.TokenRights
	AttentionRows     []uikit.DashboardAttentionRow
	Focus             uikit.DashboardPane
	SelectedProjectID int64
}

type StacksRenderState struct {
	Items      []uikit.Stack
	SelectedID int64
	Form       uikit.StackForm
	Delete     uikit.DeleteConfirm
}

type NamespacesRenderState struct {
	Items      []uikit.Namespace
	SelectedID int64
	Form       uikit.StackForm
	Delete     uikit.DeleteConfirm
}

type DependenciesRenderState struct {
	Items      []uikit.Dependency
	SelectedID int64
	Form       uikit.DependencyForm
	Delete     uikit.DeleteConfirm
}

type ProjectsRenderState struct {
	Items                []uikit.Project
	SelectedID           int64
	Dependencies         []uikit.ProjectDependency
	SelectedDependencyID int64
	LatestRun            uikit.ProjectDependencyRun
	SyncStatus           uikit.ProjectSyncStatus
	Focus                uikit.ProjectPane
	Form                 uikit.ProjectForm
	Delete               uikit.DeleteConfirm
}

type SourcesRenderState struct {
	Items      []uikit.Source
	SelectedID int64
	Form       uikit.SourceForm
	Delete     uikit.DeleteConfirm
}

type PoliciesRenderState struct {
	Items           []uikit.Policy
	SelectedID      int64
	Values          []uikit.PolicyValue
	SelectedValueID int64
	Focus           uikit.PolicyPane
	UpdateStatus    uikit.SettingsStatus
	Form            uikit.PolicyForm
	ValueForm       uikit.PolicyValueForm
	UpdateForm      uikit.PolicyUpdateForm
	Delete          uikit.DeleteConfirm
}

type DependencyViewRenderState struct {
	View              uikit.DependencyView
	SelectedProjectID int64
	ColumnOffset      int
}

type SettingsRenderState struct {
	Columns           []uikit.ProjectCheck
	Rows              []uikit.ProjectCheckRow
	SelectedProjectID int64
	TableState        uikit.SettingsTableState
	Status            uikit.SettingsStatus
}

type ReleasesRenderState struct {
	Rows              []uikit.ReleaseRow
	SelectedProjectID int64
	Period            uikit.ReleasePeriod
	Status            uikit.SettingsStatus
}

type VulnerabilitiesRenderState struct {
	Rows              []uikit.VulnProjectRow
	SelectedProjectID int64
	Items             []uikit.VulnerabilityItem
	SelectedItemIndex int
	Focus             uikit.VulnPane
	Mode              uikit.VulnMode
	Status            uikit.SettingsStatus
}

type OverlayRenderState struct {
	SearchEditing     bool
	SearchQuery       string
	SearchCount       int
	NavigationOpen    bool
	NavigationIndex   int
	Operations        []uikit.SettingsOperation
	OperationsOpen    bool
	OperationsIndex   int
	OperationsConfirm bool
	GlobalError       string
	LoadState         uikit.LoadState
}
