package uikit

type Screen int

const (
	ScreenDefault Screen = iota
	ScreenStacks
	ScreenNamespaces
	ScreenDependencies
)

type Stack struct {
	ID    int64
	Icon  string
	Name  string
	Color string
}

type Namespace struct {
	ID    int64
	Icon  string
	Name  string
	Color string
}

type Dependency struct {
	ID        int64
	StackID   int64
	StackName string
	Icon      string
	Name      string
	Color     string
}

type StackFormMode int

const (
	StackFormModeCreate StackFormMode = iota
	StackFormModeEdit
)

type StackFormField int

const (
	StackFormFieldIcon StackFormField = iota
	StackFormFieldColor
	StackFormFieldName
)

type StackForm struct {
	Open    bool
	Mode    StackFormMode
	StackID int64
	Focus   StackFormField
	Icon    string
	Color   string
	Name    string
	CanSave bool
	Error   string
}

type DependencyFormField int

const (
	DependencyFormFieldIcon DependencyFormField = iota
	DependencyFormFieldColor
	DependencyFormFieldName
	DependencyFormFieldStack
)

type DependencyForm struct {
	Open         bool
	Mode         StackFormMode
	DependencyID int64
	StackID      int64
	Focus        DependencyFormField
	Icon         string
	Color        string
	Name         string
	CanSave      bool
	Error        string
}

type DeleteConfirm struct {
	Open    bool
	StackID int64
	Name    string
	Error   string
}
