package ui

import "github.com/dnwSilver/tld/internal/ui/uikit"

type Screen = uikit.Screen

const (
	ScreenDefault = uikit.ScreenDefault
	ScreenStacks  = uikit.ScreenStacks
)

type Stack = uikit.Stack
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
type DeleteConfirm = uikit.DeleteConfirm
