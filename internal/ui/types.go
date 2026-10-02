package ui

import "github.com/dnwSilver/tld/internal/ui/uikit"

type Screen = uikit.Screen
type LoadState = uikit.LoadState
type LoadPhase = uikit.LoadPhase
type ModalKind = uikit.ModalKind
type ModalActionID = uikit.ModalActionID
type ActionID = uikit.ActionID
type ScreenAction = uikit.ScreenAction

const (
	LoadPhaseIdle    = uikit.LoadPhaseIdle
	LoadPhaseLoading = uikit.LoadPhaseLoading
	LoadPhaseReady   = uikit.LoadPhaseReady
	LoadPhaseStale   = uikit.LoadPhaseStale
	LoadPhaseError   = uikit.LoadPhaseError

	ScreenDefault         = uikit.ScreenDefault
	ScreenStacks          = uikit.ScreenStacks
	ScreenNamespaces      = uikit.ScreenNamespaces
	ScreenDependencies    = uikit.ScreenDependencies
	ScreenProjects        = uikit.ScreenProjects
	ScreenSources         = uikit.ScreenSources
	ScreenPolicies        = uikit.ScreenPolicies
	ScreenView            = uikit.ScreenView
	ScreenSettings        = uikit.ScreenSettings
	ScreenReleases        = uikit.ScreenReleases
	ScreenVulnerabilities = uikit.ScreenVulnerabilities
)

type Stack = uikit.Stack
type Namespace = uikit.Namespace
type Dependency = uikit.Dependency
type Project = uikit.Project
type DashboardPane = uikit.DashboardPane
type DashboardAttentionRow = uikit.DashboardAttentionRow
type TokenRights = uikit.TokenRights
type ProjectRight = uikit.ProjectRight
type ProjectDependencyRun = uikit.ProjectDependencyRun
type ProjectDependency = uikit.ProjectDependency
type ProjectPane = uikit.ProjectPane
type ProjectSyncStatus = uikit.ProjectSyncStatus
type DependencyView = uikit.DependencyView
type DependencyViewColumn = uikit.DependencyViewColumn
type DependencyViewRow = uikit.DependencyViewRow
type Source = uikit.Source
type Policy = uikit.Policy
type PolicyValue = uikit.PolicyValue
type PolicyPane = uikit.PolicyPane

const (
	ProjectPaneProjects     = uikit.ProjectPaneProjects
	ProjectPaneDependencies = uikit.ProjectPaneDependencies
)

const (
	DashboardPaneSummary   = uikit.DashboardPaneSummary
	DashboardPaneAttention = uikit.DashboardPaneAttention
)

const (
	PolicyPanePolicies = uikit.PolicyPanePolicies
	PolicyPaneValues   = uikit.PolicyPaneValues
)

type StackFormMode = uikit.StackFormMode

const (
	StackFormModeCreate = uikit.StackFormModeCreate
	StackFormModeEdit   = uikit.StackFormModeEdit
)

type StackFormField = uikit.StackFormField

const (
	StackFormFieldIcon   = uikit.StackFormFieldIcon
	StackFormFieldColor  = uikit.StackFormFieldColor
	StackFormFieldName   = uikit.StackFormFieldName
	StackFormFieldPolicy = uikit.StackFormFieldPolicy
)

type StackForm = uikit.StackForm
type DependencyFormField = uikit.DependencyFormField

const (
	DependencyFormFieldIcon     = uikit.DependencyFormFieldIcon
	DependencyFormFieldColor    = uikit.DependencyFormFieldColor
	DependencyFormFieldName     = uikit.DependencyFormFieldName
	DependencyFormFieldRegistry = uikit.DependencyFormFieldRegistry
	DependencyFormFieldPackage  = uikit.DependencyFormFieldPackage
	DependencyFormFieldStack    = uikit.DependencyFormFieldStack
)

type DependencyForm = uikit.DependencyForm
type SourceFormField = uikit.SourceFormField

const (
	SourceFormFieldName         = uikit.SourceFormFieldName
	SourceFormFieldURL          = uikit.SourceFormFieldURL
	SourceFormFieldPATToken     = uikit.SourceFormFieldPATToken
	SourceFormFieldType         = uikit.SourceFormFieldType
	SourceFormFieldRegistryKind = uikit.SourceFormFieldRegistryKind
)

type SourceForm = uikit.SourceForm
type ProjectFormField = uikit.ProjectFormField

const (
	ProjectFormFieldIcon      = uikit.ProjectFormFieldIcon
	ProjectFormFieldColor     = uikit.ProjectFormFieldColor
	ProjectFormFieldName      = uikit.ProjectFormFieldName
	ProjectFormFieldProjectID = uikit.ProjectFormFieldProjectID
	ProjectFormFieldNamespace = uikit.ProjectFormFieldNamespace
	ProjectFormFieldSource    = uikit.ProjectFormFieldSource
	ProjectFormFieldStack     = uikit.ProjectFormFieldStack
	ProjectFormFieldFreezing  = uikit.ProjectFormFieldFreezing
	ProjectFormFieldEndOfLife = uikit.ProjectFormFieldEndOfLife
)

type ProjectForm = uikit.ProjectForm
type PolicyFormField = uikit.PolicyFormField

const (
	PolicyFormFieldName      = uikit.PolicyFormFieldName
	PolicyFormFieldNamespace = uikit.PolicyFormFieldNamespace
)

type PolicyForm = uikit.PolicyForm
type PolicyValueFormField = uikit.PolicyValueFormField

const (
	PolicyValueFormFieldDependency = uikit.PolicyValueFormFieldDependency
	PolicyValueFormFieldRegistry   = uikit.PolicyValueFormFieldRegistry
	PolicyValueFormFieldVersion    = uikit.PolicyValueFormFieldVersion
)

type PolicyValueForm = uikit.PolicyValueForm
type PolicyUpdateFormField = uikit.PolicyUpdateFormField

const (
	PolicyUpdateFormFieldStack    = uikit.PolicyUpdateFormFieldStack
	PolicyUpdateFormFieldRegistry = uikit.PolicyUpdateFormFieldRegistry
)

type PolicyUpdateForm = uikit.PolicyUpdateForm
type DeleteConfirm = uikit.DeleteConfirm
type ProjectCheck = uikit.ProjectCheck
type CheckState = uikit.CheckState
type ProjectCheckRow = uikit.ProjectCheckRow
type SettingsTableState = uikit.SettingsTableState
type SettingsStatus = uikit.SettingsStatus
type SettingsOperation = uikit.SettingsOperation
type ReleaseRow = uikit.ReleaseRow
type ReleaseMonth = uikit.ReleaseMonth
type ReleasePeriod = uikit.ReleasePeriod
type ReleaseStatusIndicator = uikit.ReleaseStatusIndicator
type VulnCounts = uikit.VulnCounts
type VulnerabilityItem = uikit.VulnerabilityItem
type VulnProjectRow = uikit.VulnProjectRow
type VulnMode = uikit.VulnMode
type VulnPane = uikit.VulnPane

const (
	VulnModeProd = uikit.VulnModeProd
	VulnModeDev  = uikit.VulnModeDev
)

const (
	VulnPaneProjects = uikit.VulnPaneProjects
	VulnPaneDetails  = uikit.VulnPaneDetails
)

const (
	ReleasePeriodYear    = uikit.ReleasePeriodYear
	ReleasePeriodHalf    = uikit.ReleasePeriodHalf
	ReleasePeriodQuarter = uikit.ReleasePeriodQuarter
)

const (
	ReleaseStatusUnknown      = uikit.ReleaseStatusUnknown
	ReleaseStatusUntaggedMain = uikit.ReleaseStatusUntaggedMain
	ReleaseStatusDevAhead     = uikit.ReleaseStatusDevAhead
	ReleaseStatusDevBehind    = uikit.ReleaseStatusDevBehind
)

const (
	CheckStateUnknown       = uikit.CheckStateUnknown
	CheckStatePass          = uikit.CheckStatePass
	CheckStateWarning       = uikit.CheckStateWarning
	CheckStateFail          = uikit.CheckStateFail
	CheckStateNotApplicable = uikit.CheckStateNotApplicable
)

type Binding = uikit.Binding

var (
	KeyHome                 = uikit.KeyHome
	KeyStacks               = uikit.KeyStacks
	KeyNamespaces           = uikit.KeyNamespaces
	KeyDependencies         = uikit.KeyDependencies
	KeyProjects             = uikit.KeyProjects
	KeySources              = uikit.KeySources
	KeyPolicies             = uikit.KeyPolicies
	KeyView                 = uikit.KeyView
	KeySettings             = uikit.KeySettings
	KeyReleases             = uikit.KeyReleases
	KeyVulnerabilities      = uikit.KeyVulnerabilities
	KeyAdd                  = uikit.KeyAdd
	KeyEdit                 = uikit.KeyEdit
	KeyClone                = uikit.KeyClone
	KeyDelete               = uikit.KeyDelete
	KeyRefreshDeps          = uikit.KeyRefreshDeps
	KeyRefreshAll           = uikit.KeyRefreshAll
	KeyUpdatePins           = uikit.KeyUpdatePins
	KeyOperations           = uikit.KeyOperations
	KeySort                 = uikit.KeySort
	KeyColumnPick           = uikit.KeyColumnPick
	KeyVulnMode             = uikit.KeyVulnMode
	KeyToggleFocus          = uikit.KeyToggleFocus
	KeyPrev                 = uikit.KeyPrev
	KeyNext                 = uikit.KeyNext
	KeyStackPick            = uikit.KeyStackPick
	KeySourcePick           = uikit.KeySourcePick
	KeySourceTypePick       = uikit.KeySourceTypePick
	KeyRegistryPick         = uikit.KeyRegistryPick
	KeyRegistryKindPick     = uikit.KeyRegistryKindPick
	KeyNamespacePick        = uikit.KeyNamespacePick
	KeyPolicyPick           = uikit.KeyPolicyPick
	KeyDependencyPick       = uikit.KeyDependencyPick
	KeyToggleHead           = uikit.KeyToggleHead
	KeyRetryLoad            = uikit.KeyRetryLoad
	KeyQuit                 = uikit.KeyQuit
	ActionAdd               = uikit.ActionAdd
	ActionEdit              = uikit.ActionEdit
	ActionClone             = uikit.ActionClone
	ActionDelete            = uikit.ActionDelete
	ActionRefreshRow        = uikit.ActionRefreshRow
	ActionRefreshAll        = uikit.ActionRefreshAll
	ActionUpdatePins        = uikit.ActionUpdatePins
	ActionOperations        = uikit.ActionOperations
	ActionSort              = uikit.ActionSort
	ActionVulnMode          = uikit.ActionVulnMode
	ActionToggleFocus       = uikit.ActionToggleFocus
	ActionOpenNavigation    = uikit.ActionOpenNavigation
	ActionSearch            = uikit.ActionSearch
	ActionReload            = uikit.ActionReload
	ActionQuit              = uikit.ActionQuit
	ActionGoHome            = uikit.ActionGoHome
	ActionGoStacks          = uikit.ActionGoStacks
	ActionGoNamespaces      = uikit.ActionGoNamespaces
	ActionGoDependencies    = uikit.ActionGoDependencies
	ActionGoSources         = uikit.ActionGoSources
	ActionGoProjects        = uikit.ActionGoProjects
	ActionGoPolicies        = uikit.ActionGoPolicies
	ActionGoView            = uikit.ActionGoView
	ActionGoSettings        = uikit.ActionGoSettings
	ActionGoReleases        = uikit.ActionGoReleases
	ActionGoVulnerabilities = uikit.ActionGoVulnerabilities
	ModalNavigation         = uikit.ModalNavigation
	ModalOperations         = uikit.ModalOperations
	ModalActionCancel       = uikit.ModalActionCancel
	ModalActionPrev         = uikit.ModalActionPrev
	ModalActionNext         = uikit.ModalActionNext
	ModalActionConfirm      = uikit.ModalActionConfirm
)

var GlobalActions = uikit.GlobalActions
var GlobalActionForKey = uikit.GlobalActionForKey
var ModalActionForKey = uikit.ModalActionForKey

func ScreenActions(screen Screen) []ScreenAction { return uikit.ScreenActions(screen) }
func ScreenActionForKey(screen Screen, key string) (ActionID, bool) {
	return uikit.ScreenActionForKey(screen, key)
}

type NavSection = uikit.NavSection

var NavSections = uikit.NavSections

func NavSectionIndex(screen Screen) int {
	return uikit.NavSectionIndex(screen)
}
