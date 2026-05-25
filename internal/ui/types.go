package ui

import "github.com/dnwSilver/tld/internal/ui/uikit"

type Screen = uikit.Screen

const (
	ScreenDefault      = uikit.ScreenDefault
	ScreenStacks       = uikit.ScreenStacks
	ScreenNamespaces   = uikit.ScreenNamespaces
	ScreenDependencies = uikit.ScreenDependencies
	ScreenProjects     = uikit.ScreenProjects
	ScreenSources      = uikit.ScreenSources
	ScreenPolicies     = uikit.ScreenPolicies
)

type Stack = uikit.Stack
type Namespace = uikit.Namespace
type Dependency = uikit.Dependency
type Project = uikit.Project
type ProjectDependencyRun = uikit.ProjectDependencyRun
type ProjectDependency = uikit.ProjectDependency
type ProjectPane = uikit.ProjectPane
type ProjectSyncStatus = uikit.ProjectSyncStatus
type Source = uikit.Source
type Policy = uikit.Policy
type PolicyValue = uikit.PolicyValue
type PolicyPane = uikit.PolicyPane

const (
	ProjectPaneProjects     = uikit.ProjectPaneProjects
	ProjectPaneDependencies = uikit.ProjectPaneDependencies
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
	DependencyFormFieldIcon  = uikit.DependencyFormFieldIcon
	DependencyFormFieldColor = uikit.DependencyFormFieldColor
	DependencyFormFieldName  = uikit.DependencyFormFieldName
	DependencyFormFieldStack = uikit.DependencyFormFieldStack
)

type DependencyForm = uikit.DependencyForm
type SourceFormField = uikit.SourceFormField

const (
	SourceFormFieldName     = uikit.SourceFormFieldName
	SourceFormFieldURL      = uikit.SourceFormFieldURL
	SourceFormFieldPATToken = uikit.SourceFormFieldPATToken
	SourceFormFieldType     = uikit.SourceFormFieldType
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
	PolicyValueFormFieldVersion    = uikit.PolicyValueFormFieldVersion
)

type PolicyValueForm = uikit.PolicyValueForm
type DeleteConfirm = uikit.DeleteConfirm

type Binding = uikit.Binding

var (
	KeyHome           = uikit.KeyHome
	KeyStacks         = uikit.KeyStacks
	KeyNamespaces     = uikit.KeyNamespaces
	KeyDependencies   = uikit.KeyDependencies
	KeyProjects       = uikit.KeyProjects
	KeySources        = uikit.KeySources
	KeyPolicies       = uikit.KeyPolicies
	KeyAdd            = uikit.KeyAdd
	KeyEdit           = uikit.KeyEdit
	KeyDelete         = uikit.KeyDelete
	KeyRefreshDeps    = uikit.KeyRefreshDeps
	KeyPrev           = uikit.KeyPrev
	KeyNext           = uikit.KeyNext
	KeyStackPick      = uikit.KeyStackPick
	KeySourcePick     = uikit.KeySourcePick
	KeySourceTypePick = uikit.KeySourceTypePick
	KeyNamespacePick  = uikit.KeyNamespacePick
	KeyPolicyPick     = uikit.KeyPolicyPick
	KeyDependencyPick = uikit.KeyDependencyPick
	KeyQuit           = uikit.KeyQuit
)
