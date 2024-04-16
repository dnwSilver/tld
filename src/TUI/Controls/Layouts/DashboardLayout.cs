using TUI.Engine.Attributes.Orientations;
using TUI.Engine.Containers;
using TUI.Engine.Nodes;

namespace TUI.Controls.Layouts;

public class DashboardLayout : ContainerBase, IContainer
{
    private readonly INode _header;
    private readonly INode _footer;
    private readonly INode _dashboard;

    public DashboardLayout(INode header, INode dashboard, INode footer)
    {
        SetOrientationVertical();
        SetAdaptive(Orientation.Horizontal);
        SetAdaptive(Orientation.Vertical);

        header.SetFixed(Orientation.Vertical, 6);
        footer.SetFixed(Orientation.Vertical, 1);

        _header = header;
        _footer = footer;
        _dashboard = dashboard;
    }

    public override Nodes GetNodes() =>
        new()
        {
            _header, _dashboard, _footer
        };
}