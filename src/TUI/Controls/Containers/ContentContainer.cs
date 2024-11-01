using TUI.Engine.Containers;
using TUI.Engine.Nodes;

namespace TUI.Controls.Containers;

public class ContentContainer : ContainerBase
{
    private readonly Nodes _children = new();

    public void AddChildren(INode node)
    {
        _children.Add(node);
    }

    public override Nodes GetNodes()
    {
        return _children;
    }
}