using TUI.Engine.Nodes;
using TUI.Engine.Nodes.Attributes;
using TUI.Engine.Nodes.Components;
using TUI.Engine.Nodes.Containers;

namespace TUI.Engine.Rendering;

public sealed class NodeCraftsman
{
    private readonly ICanvas _canvas;
    private readonly IDrawable<IComponent> _componentCraftsman;
    private readonly IDrawable<IContainer> _containerCraftsman;

    public NodeCraftsman(
        ICanvas canvas,
        IDrawable<IComponent> componentCraftsman,
        IDrawable<IContainer> containerCraftsman)
    {
        _canvas = canvas;
        _componentCraftsman = componentCraftsman;
        _containerCraftsman = containerCraftsman;
    }

    public void Draw(INode node)
    {
        var windowSize = new Size(_canvas.Width, _canvas.Height);
        var defaultPosition = new Position(0, 0);

        switch (node)
        {
            case IContainer container:
                _containerCraftsman.Draw(container, defaultPosition, windowSize);
                break;
            case IComponent component:
                _componentCraftsman.Draw(component, defaultPosition, windowSize);
                break;
        }
    }
}