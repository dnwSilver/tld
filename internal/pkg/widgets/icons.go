package widgets

var Icons = newIconsRegistry()

func newIconsRegistry() *iconRegistry {
	return &iconRegistry{
		CopyRight: "\uf1f9", // 
	}
}

type iconRegistry struct {
	CopyRight string
}
