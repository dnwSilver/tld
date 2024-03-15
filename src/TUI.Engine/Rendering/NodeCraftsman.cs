using TUI.Engine.Attributes;
using TUI.Engine.Components;
using TUI.Engine.Containers;
using TUI.Engine.Nodes;

namespace TUI.Engine.Rendering;

internal sealed class DrawCraftsman : IDrawable<INode>
{
    private readonly IDrawable<IComponent> _componentCraftsman;
    private readonly IDrawable<IContainer> _containerCraftsman;

    public DrawCraftsman(
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