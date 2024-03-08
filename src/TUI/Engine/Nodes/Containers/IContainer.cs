using TUI.Engine.Nodes.Attributes.Orientations;

namespace TUI.Engine.Nodes.Containers;

public interface IContainer : INode,
    IWithOrientation
{
    public Nodes GetNodes();
}