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
	ScreenView
	ScreenSettings
	ScreenReleases
	ScreenVulnerabilities
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
	ID                 int64
	StackID            int64
	StackName          string
	Icon               string
	Name               string
	Color              string
	RegistryName       string
	RegistryID         int64
	RegistrySourceName string
}

type Source struct {
	ID           int64
	Icon         string
	Name         string
	Color        string
	PATToken     string
	URL          string
	Type         string
	RegistryKind string
}

type Project struct {
	ID              int64
	ProjectID       string
	NamespaceID     int64
	NamespaceName   string
	SourceID        int64
	SourceName      string
	StackID         int64
	StackName       string
	StackIcon       string
	StackColor      string
	Icon            string
	Name            string
	Color           string
	Freezing        bool
	EndOfLife       bool
	DependencyCount int
}

type DashboardPane int

const (
	DashboardPaneSummary DashboardPane = iota
	DashboardPaneAttention
)

type TokenRights struct {
	Checked    bool
	Maintainer bool
}

type DashboardAttentionRow struct {
	ProjectID        int64
	ProjectName      string
	Critical         int
	High             int
	Major            int
	MinorPatch       int
	SettingsErrors   int
	SettingsWarnings int
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
	Current   int
	Total     int
}

type DependencyView struct {
	StackID   int64
	StackName string
	Columns   []DependencyViewColumn
	Rows      []DependencyViewRow
}

type DependencyViewColumn struct {
	DependencyID  int64
	Icon          string
	Name          string
	Color         string
	PolicyVersion string
}

type DependencyViewRow struct {
	ProjectID        int64
	ProjectIcon      string
	ProjectName      string
	ProjectColor     string
	ProjectFreezing  bool
	ProjectEndOfLife bool
	Versions         map[int64]string
}

type Policy struct {
	ID              int64
	Name            string
	DependencyCount int
}

type PolicyValue struct {
	ID                 int64
	PolicyID           int64
	DependencyID       int64
	DependencyIcon     string
	DependencyName     string
	DependencyColor    string
	StackID            int64
	StackName          string
	RegistryName       string
	RegistryID         int64
	RegistrySourceName string
	Version            string
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
	DependencyFormFieldRegistry
	DependencyFormFieldPackage
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
	RegistryName string
	RegistryID   int64
	CanSave      bool
	Error        string
}

type SourceFormField int

const (
	SourceFormFieldName SourceFormField = iota
	SourceFormFieldURL
	SourceFormFieldPATToken
	SourceFormFieldType
	SourceFormFieldRegistryKind
)

type SourceForm struct {
	Open         bool
	Mode         StackFormMode
	SourceID     int64
	Focus        SourceFormField
	Name         string
	URL          string
	PATToken     string
	Type         string
	RegistryKind string
	CanSave      bool
	Error        string
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
	ProjectFormFieldFreezing
	ProjectFormFieldEndOfLife
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
	Freezing    bool
	EndOfLife   bool
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
	PolicyValueFormFieldRegistry
	PolicyValueFormFieldVersion
)

type PolicyValueForm struct {
	Open          bool
	Mode          StackFormMode
	PolicyValueID int64
	PolicyID      int64
	DependencyID  int64
	RegistryID    int64
	Focus         PolicyValueFormField
	Version       string
	CanSave       bool
	CanUpdate     bool
	Error         string
}

type PolicyUpdateFormField int

const (
	PolicyUpdateFormFieldStack PolicyUpdateFormField = iota
	PolicyUpdateFormFieldRegistry
)

type PolicyUpdateForm struct {
	Open       bool
	PolicyID   int64
	StackID    int64
	RegistryID int64
	Focus      PolicyUpdateFormField
	CanStart   bool
	Error      string
}

type DeleteConfirm struct {
	Open    bool
	StackID int64
	Name    string
	Error   string
}

type ProjectCheck struct {
	ID    string
	Title string
}

type CheckState string

const (
	CheckStateUnknown       CheckState = "unknown"
	CheckStatePass          CheckState = "pass"
	CheckStateWarning       CheckState = "warning"
	CheckStateFail          CheckState = "fail"
	CheckStateNotApplicable CheckState = "not-applicable"
)

type ProjectCheckRow struct {
	ProjectID        int64
	ProjectIcon      string
	ProjectName      string
	ProjectColor     string
	ProjectFreezing  bool
	ProjectEndOfLife bool
	Results          map[string]CheckState
	Versions         map[string]string
}

type SettingsStatus struct {
	Message string
	Running bool
	Error   string
	Current int
	Total   int
}

type SettingsOperation struct {
	ID    string
	Title string
}

type ReleasePeriod int

const (
	ReleasePeriodYear ReleasePeriod = iota
	ReleasePeriodHalf
	ReleasePeriodQuarter
)

func (p ReleasePeriod) Months() int {
	switch p {
	case ReleasePeriodHalf:
		return 6
	case ReleasePeriodQuarter:
		return 3
	default:
		return 12
	}
}

func (p ReleasePeriod) Title() string {
	switch p {
	case ReleasePeriodHalf:
		return "6 months"
	case ReleasePeriodQuarter:
		return "3 months"
	default:
		return "year"
	}
}

func (p ReleasePeriod) Next() ReleasePeriod {
	switch p {
	case ReleasePeriodYear:
		return ReleasePeriodHalf
	case ReleasePeriodHalf:
		return ReleasePeriodQuarter
	default:
		return ReleasePeriodYear
	}
}

type ReleaseMonth struct {
	Label     string
	SlotCount int
	Marks     []bool
}

type ReleaseStatusIndicator string

const (
	ReleaseStatusUnknown      ReleaseStatusIndicator = "unknown"
	ReleaseStatusUntaggedMain ReleaseStatusIndicator = "untagged-main"
	ReleaseStatusDevAhead     ReleaseStatusIndicator = "dev-ahead"
	ReleaseStatusDevBehind    ReleaseStatusIndicator = "dev-behind"
)

type ReleaseRow struct {
	ProjectID        int64
	ProjectIcon      string
	ProjectName      string
	ProjectColor     string
	ProjectFreezing  bool
	ProjectEndOfLife bool
	HasReleases      bool
	Months           []ReleaseMonth
	Statuses         []ReleaseStatusIndicator
}

type VulnCounts struct {
	Critical int
	High     int
	Medium   int
	Low      int
	None     int
}

type VulnerabilityItem struct {
	Package     string
	Severity    string
	Title       string
	Description string
	Range       string
}

type VulnProjectRow struct {
	ProjectID        int64
	ProjectIcon      string
	ProjectName      string
	ProjectColor     string
	ProjectFreezing  bool
	ProjectEndOfLife bool
	Scanned          bool
	Counts           VulnCounts
}

type VulnMode int

const (
	VulnModeProd VulnMode = iota
	VulnModeDev
)

func (m VulnMode) Title() string {
	if m == VulnModeDev {
		return "dev"
	}
	return "prod"
}

func (m VulnMode) Next() VulnMode {
	if m == VulnModeProd {
		return VulnModeDev
	}
	return VulnModeProd
}

type VulnPane int

const (
	VulnPaneProjects VulnPane = iota
	VulnPaneDetails
)
