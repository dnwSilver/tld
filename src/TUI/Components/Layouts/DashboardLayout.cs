using TUI.Engine.Nodes;
using TUI.Engine.Nodes.Attributes.Orientations;
using TUI.Engine.Nodes.Attributes.Resizing;
using TUI.Engine.Nodes.Components;
using TUI.Engine.Nodes.Containers;

namespace TUI.Components.Layouts;

public class DashboardLayout : ContainerBase, IContainer
{
    public new Orientation Orientation => Orientation.Vertical;

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

    public Resizing ResizingHorizontal => Resizing.Adaptive;

    public Resizing ResizingVertical => Resizing.Adaptive;
}