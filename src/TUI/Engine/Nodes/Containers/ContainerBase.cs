using TUI.Engine.Nodes.Attributes;
using TUI.Engine.Nodes.Attributes.Orientations;

namespace TUI.Engine.Nodes.Containers;

public abstract class ContainerBase : NodeBase, IContainer
{
    public Orientation Orientation => Orientation.Horizontal;

    public Size GetSketchSize()
    {
        throw new NotImplementedException();
    }

    public abstract Nodes GetNodes();
}