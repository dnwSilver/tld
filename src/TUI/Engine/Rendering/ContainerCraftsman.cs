using TUI.Engine.Nodes;
using TUI.Engine.Nodes.Attributes;
using TUI.Engine.Nodes.Attributes.Orientations;
using TUI.Engine.Nodes.Attributes.Resizing;
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

    public Size Draw(IContainer container, Position pencil, Size maxSize)
    {
        var controlNumber = 0;
        var nextNodePosition = pencil;
        var nodes = container.GetNodes();
        var sketchSize = container.GetSize(maxSize);
        Debug(nextNodePosition, nextNodePosition, maxSize);

        while (controlNumber < nodes.Count)
        {
            var node = nodes[controlNumber];
            nextNodePosition = DrawNode(node, container, nextNodePosition, sketchSize);
            controlNumber++;
        }

        return sketchSize;
    }

    private Position DrawNode(INode node, IContainer container, Position nodePosition, Size maxSize)
    {
        switch (node)
        {
            case IContainer childContainer:
                var containerSize = Draw(childContainer, nodePosition, maxSize);
                return GetNextNodePosition(container, containerSize, nodePosition);
            case IComponent childComponent:
                var componentSize = _componentCraftsman.Draw(childComponent, nodePosition, maxSize);
                return GetNextNodePosition(container, maxSize, nodePosition, componentSize);
            default:
                throw new InvalidCastException();
        }
    }

    private static Position GetNextNodePosition(
        IContainer container,
        Size defaultSize,
        Position position,
        Size? componentSize = null)
    {
        switch (container.Orientation)
        {
            case Orientation.Horizontal:
                var componentWidth = container.ResizingHorizontal switch
                {
                    Resizing.Adaptive => defaultSize.Width,
                    Resizing.Fixed => componentSize?.Width.Max(container.GetFixedSize().Width) ??
                                      container.GetFixedSize().Width,
                    _ => 0
                };
                return position with { Left = position.Left + componentWidth };
            case Orientation.Vertical:
                var componentHeight = container.ResizingVertical switch
                {
                    Resizing.Adaptive => defaultSize.Height,
                    Resizing.Fixed => componentSize?.Height.Max(container.GetFixedSize().Height) ??
                                      container.GetFixedSize().Height,
                    _ => 0
                };
                return position with { Top = position.Top + componentHeight };
            default:
                throw new InvalidCastException();
        }
    }
}