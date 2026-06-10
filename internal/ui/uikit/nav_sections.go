package uikit

type NavSection struct {
	Screen  Screen
	Binding Binding
}

var NavSections = []NavSection{
	{Screen: ScreenDefault, Binding: KeyHome},
	{Screen: ScreenStacks, Binding: KeyStacks},
	{Screen: ScreenNamespaces, Binding: KeyNamespaces},
	{Screen: ScreenDependencies, Binding: KeyDependencies},
	{Screen: ScreenSources, Binding: KeySources},
	{Screen: ScreenProjects, Binding: KeyProjects},
	{Screen: ScreenPolicies, Binding: KeyPolicies},
	{Screen: ScreenView, Binding: KeyView},
	{Screen: ScreenSettings, Binding: KeySettings},
	{Screen: ScreenReleases, Binding: KeyReleases},
	{Screen: ScreenVulnerabilities, Binding: KeyVulnerabilities},
}

func NavSectionIndex(screen Screen) int {
	for index, section := range NavSections {
		if section.Screen == screen {
			return index
		}
	}

	return 0
}
