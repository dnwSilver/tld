using TUI.Engine.Containers;
using TUI.Engine.Nodes;

namespace TUI.Engine.Tests.Stubs;

public class TestContainer : ContainerBase
{
    private readonly Nodes.Nodes _nodes = new();

    public override Nodes.Nodes GetNodes()
    {
        return _nodes;
    }

    public TestContainer SetNodes(params INode[] nodes)
    {
        _nodes.AddRange(nodes);
        return this;
    }
}