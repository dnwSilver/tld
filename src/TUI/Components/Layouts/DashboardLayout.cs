using TUI.Engine.Attributes.Orientations;
using TUI.Engine.Components;
using TUI.Engine.Containers;
using TUI.Engine.Nodes;

namespace TUI.Components.Layouts;

public class DashboardLayout : ContainerBase, IContainer
{
    public DashboardLayout()
    {
        SetOrientationVertical();
        SetAdaptive(Orientation.Horizontal);
        SetAdaptive(Orientation.Vertical);
    }

    private INode _header;
    private INode _footer;

    public override Nodes GetNodes() =>
        new()
        {
            _header, _footer
        };

    public DashboardLayout AddHeader(IContainer header)
    {
        _header = header;
        return this;
    }

    public DashboardLayout AddFooter(IComponent footer)
    {
        _footer = footer;
        return this;
    }

    public string Render()
    {
        throw new NotImplementedException();
    }
}