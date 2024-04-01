using TUI.Engine.Attributes;
using TUI.Engine.Components;
using TUI.Engine.Theme;

namespace TUI.Controls.Components;

public class BreadCrumbsComponent : ComponentBase
{
    private readonly List<string> _crumbs = new() { " " };

    public BreadCrumbsComponent(params string[] crumbs)
    {
        _crumbs.AddRange(crumbs);
    }

    protected override Sketch DrawComponent(Size minSize)
    {
        return new Sketch(string.Join("  ".Hint(), _crumbs).Hint());
    }
}