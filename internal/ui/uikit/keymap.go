package uikit

type Binding struct {
	Keys   []string
	Label  string
	Symbol string
	Hint   string
}

func (b Binding) Matches(key string) bool {
	for _, k := range b.Keys {
		if k == key {
			return true
		}
	}

	return false
}

var (
	KeyHome            = Binding{Keys: []string{"0", "cmd+0", "alt+0"}, Label: "home", Symbol: SymbolDashboard, Hint: "0 home"}
	KeyStacks          = Binding{Keys: []string{"1", "cmd+1", "alt+1"}, Label: "stacks", Symbol: SymbolStack, Hint: "1 stacks"}
	KeyNamespaces      = Binding{Keys: []string{"2", "cmd+2", "alt+2"}, Label: "namespaces", Symbol: SymbolNamespace, Hint: "2 namespaces"}
	KeyDependencies    = Binding{Keys: []string{"3", "cmd+3", "alt+3"}, Label: "deps", Symbol: SymbolDependency, Hint: "3 deps"}
	KeySources         = Binding{Keys: []string{"4", "cmd+4", "alt+4"}, Label: "sources", Symbol: SymbolSource, Hint: "4 sources"}
	KeyProjects        = Binding{Keys: []string{"5", "cmd+5", "alt+5"}, Label: "projects", Symbol: SymbolProject, Hint: "5 projects"}
	KeyPolicies        = Binding{Keys: []string{"6", "cmd+6", "alt+6"}, Label: "policies", Symbol: SymbolPolicy, Hint: "6 policies"}
	KeyView            = Binding{Keys: []string{"7", "cmd+7", "alt+7"}, Label: "view", Symbol: SymbolDependency, Hint: "7 view"}
	KeySettings        = Binding{Keys: []string{"8", "cmd+8", "alt+8"}, Label: "settings", Symbol: SymbolSettings, Hint: "8 settings"}
	KeyReleases        = Binding{Keys: []string{"9", "cmd+9", "alt+9"}, Label: "releases", Symbol: SymbolReleases, Hint: "9 releases"}
	KeyVulnerabilities = Binding{Keys: []string{"v", "cmd+v", "alt+v"}, Label: "vulns", Symbol: SymbolVulnerabilities, Hint: "v vulns"}

	KeyAdd              = Binding{Keys: []string{"a"}, Label: "add", Symbol: SymbolAdd, Hint: "[a] add"}
	KeyEdit             = Binding{Keys: []string{"e"}, Label: "edit", Symbol: SymbolEdit, Hint: "[e] edit"}
	KeyClone            = Binding{Keys: []string{"c"}, Label: "clone", Symbol: SymbolAdd, Hint: "[c] clone"}
	KeyDelete           = Binding{Keys: []string{"d"}, Label: "delete", Symbol: SymbolDelete, Hint: "[d] delete"}
	KeyRefreshDeps      = Binding{Keys: []string{"r"}, Label: "refresh deps", Symbol: SymbolDependency, Hint: "[r] refresh row"}
	KeyRefreshAll       = Binding{Keys: []string{"R", "shift+r"}, Label: "refresh all", Symbol: SymbolDependency, Hint: "[R] refresh all"}
	KeyUpdatePins       = Binding{Keys: []string{"u"}, Label: "update pins", Symbol: SymbolDependency, Hint: "[u] update pins"}
	KeyVulnMode         = Binding{Keys: []string{"m"}, Label: "vulnerability mode", Symbol: SymbolVulnerabilities, Hint: "[m] prod/dev"}
	KeyPrev             = Binding{Keys: []string{"up", "k"}, Label: "prev", Symbol: SymbolSelectPrev, Hint: "[k] prev"}
	KeyNext             = Binding{Keys: []string{"down", "j"}, Label: "next", Symbol: SymbolSelectNext, Hint: "[j] next"}
	KeyStackPick        = Binding{Keys: []string{"left", "right", "h", "l"}, Label: "stack", Symbol: SymbolStack, Hint: "[h/l] Stack"}
	KeySourcePick       = Binding{Keys: []string{"left", "right", "h", "l"}, Label: "source", Symbol: SymbolSource, Hint: "[h/l] Source"}
	KeySourceTypePick   = Binding{Keys: []string{"left", "right", "h", "l"}, Label: "type", Symbol: SymbolSource, Hint: "[h/l] Type"}
	KeyRegistryPick     = Binding{Keys: []string{"left", "right", "h", "l"}, Label: "registry", Symbol: SymbolSource, Hint: "[h/l] Registry"}
	KeyRegistryKindPick = Binding{Keys: []string{"left", "right", "h", "l"}, Label: "kind", Symbol: SymbolSource, Hint: "[h/l] Kind"}
	KeyNamespacePick    = Binding{Keys: []string{"left", "right", "h", "l"}, Label: "namespace", Symbol: SymbolNamespace, Hint: "[h/l] Namespace"}
	KeyPolicyPick       = Binding{Keys: []string{"left", "right", "h", "l"}, Label: "policy", Symbol: SymbolPolicy, Hint: "[h/l] Policy"}
	KeyDependencyPick   = Binding{Keys: []string{"left", "right", "h", "l"}, Label: "dependency", Symbol: SymbolDependency, Hint: "[h/l] Dep"}

	KeyToggleHead = Binding{Keys: []string{"?", "shift+/"}, Label: "sections", Symbol: SymbolToggleHead, Hint: "[Shift+/] sections"}
	KeyQuit       = Binding{Keys: []string{"ctrl+c", "esc", "q"}, Label: "quit", Symbol: SymbolQuit, Hint: "quit"}

	KeyCancel     = Binding{Keys: []string{"esc"}, Label: "cancel", Hint: "[Esc] Cancel"}
	KeySave       = Binding{Keys: []string{"enter"}, Label: "save", Hint: "[Enter] Save"}
	KeyConfirmNo  = Binding{Keys: []string{"esc"}, Label: "no", Hint: "[Esc] No"}
	KeyConfirmYes = Binding{Keys: []string{"enter"}, Label: "yes", Hint: "[Enter] Yes"}
)
