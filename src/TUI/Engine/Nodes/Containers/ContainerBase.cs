using TUI.Engine.Nodes.Attributes.Orientations;

namespace TUI.Engine.Nodes.Containers;

public abstract class ContainerBase : NodeBase, IContainer
{
    public Orientation Orientation => Orientation.Horizontal;

    public abstract Nodes GetNodes();
}