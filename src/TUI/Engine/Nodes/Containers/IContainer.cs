using TUI.Engine.Nodes.Attributes.Orientations;
using TUI.Engine.Nodes.Attributes.Resizings;

namespace TUI.Engine.Nodes.Containers;

public interface IContainer : INode,
    IWithOrientation,
    IWithResizing
{
    public Nodes GetNodes();
}