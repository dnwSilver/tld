using TUI.Engine.Nodes;
using TUI.Engine.Nodes.Attributes;
using TUI.Engine.Nodes.Attributes.Orientations;
using TUI.Engine.Nodes.Components;
using TUI.Engine.Nodes.Containers;

namespace TUI.Engine.Rendering;

public sealed class ContainerCraftsman : CraftsmanBase, IDrawable<IContainer>
{
    private readonly IDrawable<IComponent> _componentCraftsman;

    public ContainerCraftsman(IDrawable<IComponent> componentCraftsman)
    {
        _componentCraftsman = componentCraftsman;
    }


    public Size Draw(IContainer container, Position sketchPosition, Size allowableSize)
    {
        var sketchSize = container.GetSize(allowableSize);
        var controlNumber = 0;
        var nodes = container.GetNodes();

        while (controlNumber < nodes.Count)
        {
            var node = nodes[controlNumber];
            sketchPosition = DrawNode(node, container.Orientation, sketchPosition, sketchSize);
            controlNumber++;
        }

        Debug(sketchPosition, sketchPosition, allowableSize);
        return sketchSize;
    }

    private Position DrawNode(INode node, Orientation orientation, Position sketchPosition, Size sketchSize)
    {
        switch (node)
        {
            case IContainer childContainer:
                Draw(childContainer, sketchPosition, sketchSize);
                return GetNextNodePosition(orientation, sketchSize, sketchPosition);
            case IComponent childComponent:
                var componentSize = _componentCraftsman.Draw(childComponent, sketchPosition, sketchSize);
                return GetNextNodePosition(orientation, sketchSize, sketchPosition, componentSize);
            default:
                throw new InvalidCastException();
        }
    }

    private static Position GetNextNodePosition(
        Orientation orientation,
        Size defaultSize,
        Position position,
        Size? componentSize = null)
    {
        switch (orientation)
        {
            case Orientation.Horizontal:
                var componentWidth = componentSize?.Width.Max(defaultSize.Width) ?? defaultSize.Width;
                return position with { Left = position.Left + componentWidth };
            case Orientation.Vertical:
                var componentHeight = componentSize?.Height.Max(defaultSize.Height) ?? defaultSize.Height;
                return position with { Top = position.Top + componentHeight };
            default:
                throw new InvalidCastException();
        }
    }
}