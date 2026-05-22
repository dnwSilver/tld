package ui

import "github.com/dnwSilver/tld/internal/ui/uikit"

type Screen = uikit.Screen

const (
	ScreenDefault      = uikit.ScreenDefault
	ScreenStacks       = uikit.ScreenStacks
	ScreenNamespaces   = uikit.ScreenNamespaces
	ScreenDependencies = uikit.ScreenDependencies
	ScreenSources      = uikit.ScreenSources
	ScreenPolicies     = uikit.ScreenPolicies
)

type Stack = uikit.Stack
type Namespace = uikit.Namespace
type Dependency = uikit.Dependency
type Source = uikit.Source
type Policy = uikit.Policy
type PolicyValue = uikit.PolicyValue
type PolicyPane = uikit.PolicyPane

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
	KeySources        = uikit.KeySources
	KeyPolicies       = uikit.KeyPolicies
	KeyAdd            = uikit.KeyAdd
	KeyEdit           = uikit.KeyEdit
	KeyDelete         = uikit.KeyDelete
	KeyPrev           = uikit.KeyPrev
	KeyNext           = uikit.KeyNext
	KeyStackPick      = uikit.KeyStackPick
	KeySourceTypePick = uikit.KeySourceTypePick
	KeyNamespacePick  = uikit.KeyNamespacePick
	KeyPolicyPick     = uikit.KeyPolicyPick
	KeyDependencyPick = uikit.KeyDependencyPick
	KeyQuit           = uikit.KeyQuit
)
