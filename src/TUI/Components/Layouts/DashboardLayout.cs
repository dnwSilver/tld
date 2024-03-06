using TUI.Engine.Nodes;
using TUI.Engine.Nodes.Components;
using TUI.Engine.Nodes.Containers;

namespace TUI.Components.Layouts;

public class DashboardLayout : IContainer
{
    public Orientation Orientation { get; } = Orientation.Vertical;

    private INode _header;
    private INode _footer;

    public Nodes Nodes =>
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