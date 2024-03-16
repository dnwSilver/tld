using TUI.Engine.Attributes.Alignments;
using TUI.Engine.Attributes.Orientations;
using TUI.Engine.Components;
using TUI.Engine.Containers;
using TUI.Engine.Nodes;
using TUI.Engine.Theme;

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
    private INode _dashboard;

    public override Nodes GetNodes() =>
        new()
        {
            _header, _dashboard, _footer
        };

    public void AddDashboard(IComponent dashboard)
    {
        _dashboard = dashboard;
    }

    public void AddHeader(IContainer header)
    {
        header.SetFixed(Orientation.Vertical, 6);
        _header = header;
    }

    public void AddFooter(IComponent footer)
    {
        footer.SetFixed(Orientation.Vertical, 1);
        footer.SetPaddingRight(Level.Normal);
        footer.SetAlignment(Horizontal.Right);
        footer.SetAlignment(Vertical.Bottom);
        _footer = footer;
    }

    public string Render()
    {
        throw new NotImplementedException();
    }
}