using TUI.Engine.Nodes;
using TUI.Engine.Nodes.Attributes;
using TUI.Engine.Nodes.Components;
using TUI.Engine.Nodes.Containers;

namespace TUI.Engine.Rendering;

/// <summary>
/// 🍀
/// </summary>
public sealed class NodeCraftsman : IDrawable<INode>
{
    private readonly IDrawable<IComponent> _componentCraftsman;
    private readonly IDrawable<IContainer> _containerCraftsman;

    public NodeCraftsman(
        IDrawable<IComponent> componentCraftsman,
        IDrawable<IContainer> containerCraftsman)
    {
        _componentCraftsman = componentCraftsman;
        _containerCraftsman = containerCraftsman;
    }

    public Size Draw(INode node, Position pencil, Size maxSize) =>
        node switch
        {
            IContainer container => _containerCraftsman.Draw(container, pencil, maxSize),
            IComponent component => _componentCraftsman.Draw(component, pencil, maxSize),
            _ => throw new InvalidCastException("Unknown node type.")
        };
}