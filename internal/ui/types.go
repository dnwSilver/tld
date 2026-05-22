package ui

import "github.com/dnwSilver/tld/internal/ui/uikit"

type Screen = uikit.Screen

const (
	ScreenDefault      = uikit.ScreenDefault
	ScreenStacks       = uikit.ScreenStacks
	ScreenNamespaces   = uikit.ScreenNamespaces
	ScreenDependencies = uikit.ScreenDependencies
)

type Stack = uikit.Stack
type Namespace = uikit.Namespace
type Dependency = uikit.Dependency
type StackFormMode = uikit.StackFormMode

const (
	StackFormModeCreate = uikit.StackFormModeCreate
	StackFormModeEdit   = uikit.StackFormModeEdit
)

type StackFormField = uikit.StackFormField

const (
	StackFormFieldIcon  = uikit.StackFormFieldIcon
	StackFormFieldColor = uikit.StackFormFieldColor
	StackFormFieldName  = uikit.StackFormFieldName
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
type DeleteConfirm = uikit.DeleteConfirm

type Binding = uikit.Binding

var (
	KeyHome         = uikit.KeyHome
	KeyStacks       = uikit.KeyStacks
	KeyNamespaces   = uikit.KeyNamespaces
	KeyDependencies = uikit.KeyDependencies
	KeyAdd          = uikit.KeyAdd
	KeyEdit         = uikit.KeyEdit
	KeyDelete       = uikit.KeyDelete
	KeyPrev         = uikit.KeyPrev
	KeyNext         = uikit.KeyNext
	KeyStackPick    = uikit.KeyStackPick
	KeyQuit         = uikit.KeyQuit
)
