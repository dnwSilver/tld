using TUI.Engine.Attributes.Orientations;
using TUI.Engine.Nodes;

namespace TUI.Engine.Containers;

public interface IContainer : INode, IWithOrientation
{
    public Nodes.Nodes GetNodes();
}