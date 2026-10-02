package uikit

type ActionID string

const (
	ActionAdd               ActionID = "add"
	ActionEdit              ActionID = "edit"
	ActionClone             ActionID = "clone"
	ActionDelete            ActionID = "delete"
	ActionRefreshRow        ActionID = "refresh-row"
	ActionRefreshAll        ActionID = "refresh-all"
	ActionUpdatePins        ActionID = "update-pins"
	ActionOperations        ActionID = "operations"
	ActionSort              ActionID = "sort"
	ActionVulnMode          ActionID = "vuln-mode"
	ActionToggleFocus       ActionID = "toggle-focus"
	ActionOpenNavigation    ActionID = "open-navigation"
	ActionSearch            ActionID = "search"
	ActionReload            ActionID = "reload"
	ActionQuit              ActionID = "quit"
	ActionGoHome            ActionID = "go-home"
	ActionGoStacks          ActionID = "go-stacks"
	ActionGoNamespaces      ActionID = "go-namespaces"
	ActionGoDependencies    ActionID = "go-dependencies"
	ActionGoSources         ActionID = "go-sources"
	ActionGoProjects        ActionID = "go-projects"
	ActionGoPolicies        ActionID = "go-policies"
	ActionGoView            ActionID = "go-view"
	ActionGoSettings        ActionID = "go-settings"
	ActionGoReleases        ActionID = "go-releases"
	ActionGoVulnerabilities ActionID = "go-vulnerabilities"
)

type ScreenAction struct {
	ID      ActionID
	Binding Binding
}

var screenActions = map[Screen][]ScreenAction{
	ScreenDefault:      {{ID: ActionToggleFocus, Binding: KeyToggleFocus}},
	ScreenStacks:       crudActions(),
	ScreenNamespaces:   crudActions(),
	ScreenDependencies: crudActions(),
	ScreenSources:      crudActions(),
	ScreenProjects: append(crudActions(),
		ScreenAction{ID: ActionClone, Binding: KeyClone},
		ScreenAction{ID: ActionRefreshRow, Binding: KeyRefreshDeps},
		ScreenAction{ID: ActionRefreshAll, Binding: KeyRefreshAll},
		ScreenAction{ID: ActionToggleFocus, Binding: KeyToggleFocus},
	),
	ScreenPolicies: append(crudActions(), ScreenAction{ID: ActionUpdatePins, Binding: KeyUpdatePins}, ScreenAction{ID: ActionToggleFocus, Binding: KeyToggleFocus}),
	ScreenView: {
		{ID: ActionRefreshRow, Binding: KeyRefreshDeps},
		{ID: ActionRefreshAll, Binding: KeyRefreshAll},
		{ID: ActionToggleFocus, Binding: KeyToggleFocus},
	},
	ScreenSettings: {
		{ID: ActionRefreshRow, Binding: KeyRefreshDeps},
		{ID: ActionRefreshAll, Binding: KeyRefreshAll},
		{ID: ActionOperations, Binding: KeyOperations},
		{ID: ActionSort, Binding: KeySort},
	},
	ScreenReleases: {
		{ID: ActionRefreshRow, Binding: KeyRefreshDeps},
		{ID: ActionRefreshAll, Binding: KeyRefreshAll},
		{ID: ActionToggleFocus, Binding: KeyToggleFocus},
	},
	ScreenVulnerabilities: {
		{ID: ActionRefreshRow, Binding: KeyRefreshDeps},
		{ID: ActionRefreshAll, Binding: KeyRefreshAll},
		{ID: ActionVulnMode, Binding: KeyVulnMode},
		{ID: ActionToggleFocus, Binding: KeyToggleFocus},
	},
}

var globalActions = []ScreenAction{
	{ID: ActionOpenNavigation, Binding: KeyToggleHead},
	{ID: ActionSearch, Binding: KeySearch},
	{ID: ActionReload, Binding: KeyRetryLoad},
	{ID: ActionGoHome, Binding: KeyHome},
	{ID: ActionGoStacks, Binding: KeyStacks},
	{ID: ActionGoNamespaces, Binding: KeyNamespaces},
	{ID: ActionGoDependencies, Binding: KeyDependencies},
	{ID: ActionGoSources, Binding: KeySources},
	{ID: ActionGoProjects, Binding: KeyProjects},
	{ID: ActionGoPolicies, Binding: KeyPolicies},
	{ID: ActionGoView, Binding: KeyView},
	{ID: ActionGoSettings, Binding: KeySettings},
	{ID: ActionGoReleases, Binding: KeyReleases},
	{ID: ActionGoVulnerabilities, Binding: KeyVulnerabilities},
	{ID: ActionQuit, Binding: KeyQuit},
}

func crudActions() []ScreenAction {
	return []ScreenAction{
		{ID: ActionAdd, Binding: KeyAdd},
		{ID: ActionEdit, Binding: KeyEdit},
		{ID: ActionDelete, Binding: KeyDelete},
	}
}

func ScreenActions(screen Screen) []ScreenAction {
	actions := screenActions[screen]
	return append([]ScreenAction(nil), actions...)
}

func ScreenActionForKey(screen Screen, key string) (ActionID, bool) {
	for _, action := range screenActions[screen] {
		if action.Binding.Matches(key) {
			return action.ID, true
		}
	}
	return "", false
}

func GlobalActions() []ScreenAction {
	return append([]ScreenAction(nil), globalActions...)
}

func GlobalActionForKey(key string) (ActionID, bool) {
	for _, action := range globalActions {
		if action.Binding.Matches(key) {
			return action.ID, true
		}
	}
	return "", false
}

type ModalKind string

const (
	ModalNavigation ModalKind = "navigation"
	ModalOperations ModalKind = "operations"
)

type ModalActionID string

const (
	ModalActionCancel  ModalActionID = "cancel"
	ModalActionPrev    ModalActionID = "prev"
	ModalActionNext    ModalActionID = "next"
	ModalActionConfirm ModalActionID = "confirm"
)

func ModalActionForKey(kind ModalKind, key string) (ModalActionID, bool) {
	if KeyCancel.Matches(key) || (kind == ModalNavigation && KeyToggleHead.Matches(key)) || (kind == ModalOperations && KeyOperations.Matches(key)) {
		return ModalActionCancel, true
	}
	if KeyPrev.Matches(key) {
		return ModalActionPrev, true
	}
	if KeyNext.Matches(key) {
		return ModalActionNext, true
	}
	if KeySave.Matches(key) {
		return ModalActionConfirm, true
	}
	return "", false
}
