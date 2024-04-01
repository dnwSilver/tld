using TUI.Controls.Statics;
using TUI.Engine.Attributes.Alignments;
using TUI.Engine.Components;
using TUI.Engine.Containers;
using TUI.Engine.Nodes;
using TUI.Engine.Theme;

namespace TUI.Controls.Containers;

public class FooterContainer : ContainerBase
{
    private readonly INode _breadcrumbs;

    public FooterContainer(IComponent breadcrumbs)
    {
        breadcrumbs.SetAlignment(Horizontal.Left);
        breadcrumbs.SetPaddingLeft(Level.Normal);
        _breadcrumbs = breadcrumbs;
    }

    public override Nodes GetNodes()
    {
        var copyright = new CopyrightComponent();
        copyright.SetAlignment(Horizontal.Right);
        copyright.SetPaddingRight(Level.Normal);
        return new Nodes { _breadcrumbs, copyright };
    }
}