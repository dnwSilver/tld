package uikit

type Screen int

const (
	ScreenDefault Screen = iota
	ScreenStacks
	ScreenNamespaces
	ScreenDependencies
	ScreenProjects
	ScreenSources
	ScreenPolicies
)

type Stack struct {
	ID         int64
	Icon       string
	Name       string
	Color      string
	PolicyName string
}

type Namespace struct {
	ID       int64
	Icon     string
	Name     string
	Color    string
	PolicyID int64
}

type Dependency struct {
	ID        int64
	StackID   int64
	StackName string
	Icon      string
	Name      string
	Color     string
}

type Source struct {
	ID       int64
	Icon     string
	Name     string
	Color    string
	PATToken string
	URL      string
	Type     string
}

type Project struct {
	ID            int64
	ProjectID     string
	NamespaceID   int64
	NamespaceName string
	SourceID      int64
	SourceName    string
	StackID       int64
	StackName     string
	StackIcon     string
	StackColor    string
	Icon          string
	Name          string
	Color         string
}

type ProjectDependencyRun struct {
	CommitShortSHA string
	Status         string
	UpdatedAt      string
	Error          string
	HasValue       bool
}

type ProjectDependency struct {
	ID             int64
	Name           string
	Version        string
	DependencyType string
	SourceFile     string
}

type ProjectPane int

const (
	ProjectPaneProjects ProjectPane = iota
	ProjectPaneDependencies
)

type ProjectSyncStatus struct {
	ProjectID int64
	Message   string
	Running   bool
	Error     string
}

type Policy struct {
	ID              int64
	Name            string
	DependencyCount int
}

type PolicyValue struct {
	ID              int64
	PolicyID        int64
	DependencyID    int64
	DependencyIcon  string
	DependencyName  string
	DependencyColor string
	Version         string
}

type PolicyPane int

const (
	PolicyPanePolicies PolicyPane = iota
	PolicyPaneValues
)

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
	StackFormFieldPolicy
)

type StackForm struct {
	Open     bool
	Mode     StackFormMode
	StackID  int64
	Focus    StackFormField
	Icon     string
	Color    string
	Name     string
	PolicyID int64
	Policy   string
	CanSave  bool
	Error    string
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

type SourceFormField int

const (
	SourceFormFieldName SourceFormField = iota
	SourceFormFieldURL
	SourceFormFieldPATToken
	SourceFormFieldType
)

type SourceForm struct {
	Open     bool
	Mode     StackFormMode
	SourceID int64
	Focus    SourceFormField
	Name     string
	URL      string
	PATToken string
	Type     string
	CanSave  bool
	Error    string
}

type ProjectFormField int

const (
	ProjectFormFieldIcon ProjectFormField = iota
	ProjectFormFieldColor
	ProjectFormFieldName
	ProjectFormFieldProjectID
	ProjectFormFieldNamespace
	ProjectFormFieldSource
	ProjectFormFieldStack
)

type ProjectForm struct {
	Open        bool
	Mode        StackFormMode
	ID          int64
	ProjectID   string
	NamespaceID int64
	SourceID    int64
	StackID     int64
	Focus       ProjectFormField
	Icon        string
	Color       string
	Name        string
	CanSave     bool
	Error       string
}

type PolicyFormField int

const (
	PolicyFormFieldName PolicyFormField = iota
	PolicyFormFieldNamespace
)

type PolicyForm struct {
	Open        bool
	Mode        StackFormMode
	PolicyID    int64
	NamespaceID int64
	Focus       PolicyFormField
	Name        string
	CanSave     bool
	Error       string
}

type PolicyValueFormField int

const (
	PolicyValueFormFieldDependency PolicyValueFormField = iota
	PolicyValueFormFieldVersion
)

type PolicyValueForm struct {
	Open          bool
	Mode          StackFormMode
	PolicyValueID int64
	PolicyID      int64
	DependencyID  int64
	Focus         PolicyValueFormField
	Version       string
	CanSave       bool
	Error         string
}

type DeleteConfirm struct {
	Open    bool
	StackID int64
	Name    string
	Error   string
}
