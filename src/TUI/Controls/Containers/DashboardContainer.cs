using TUI.Controls.Components;
using TUI.Engine.Attributes.Orientations;
using TUI.Engine.Containers;
using TUI.Engine.Nodes;
using TUI.Engine.Theme;

namespace TUI.Controls.Containers;

public class DashboardContainer : ContainerBase
{
    private readonly Nodes _children = new();
    private readonly ContentContainer _content;
    
    public DashboardContainer()
    {
        var panel = new PanelComponent(" ".Info() + " Dependencies".Main());
        _content = new ContentContainer();
        _content.SetOrientationVertical();
        SetOrientationVertical();
        
        _children.Add(panel);
        _children.Add(_content);
    }
    
    public void AddChildren(IContainer node)
    {
        node.SetFixed(Orientation.Vertical, 1);
        _content.AddChildren(node);
    }
    
    public override Nodes GetNodes() => _children;
}