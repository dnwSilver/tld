using TUI.Engine.Attributes;
using TUI.Engine.Attributes.Orientations;
using TUI.Engine.Attributes.Resizings;
using TUI.Engine.Components;
using TUI.Engine.Containers;
using TUI.Engine.Nodes;

namespace TUI.Engine.Rendering;

internal sealed class ContainerCraftsman : CraftsmanBase, IDrawable<IContainer>
{
    private readonly IDrawable<IComponent> _componentCraftsman;

    public ContainerCraftsman(IDrawable<IComponent> componentCraftsman)
    {
        _componentCraftsman = componentCraftsman;
    }

    public Size Draw(IContainer container, Position pencil, Size maxSize)
    {
        var nodeNumber = 0;
        var nextNodePosition = pencil;
        var nodes = container.GetNodes();

        Debug(nextNodePosition, maxSize);

        while (nodeNumber < nodes.Count)
        {
            var node = nodes[nodeNumber];
            var nodeSize = node.GetSize(container, nodeNumber, maxSize);

            nextNodePosition = DrawNode(node, container, nextNodePosition, nodeSize);
            nodeNumber++;
        }

        return maxSize;
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
                return childComponent.IsRelative
                    ? GetNextNodePosition(container, maxSize, nodePosition, componentSize)
                    : nodePosition;
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