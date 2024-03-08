using System.Diagnostics;
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

        while (controlNumber < container.GetNodes().Count)
        {
            var node = container.GetNodes()[controlNumber];

            sketchPosition = RenderNode(node, container.Orientation, sketchSize, sketchPosition);
            controlNumber++;
        }

        Debug(sketchPosition, sketchPosition, allowableSize);
        return sketchSize;
    }

    private Position RenderNode(INode node, Orientation orientation, Size defaultSize, Position position)
    {
        switch (node)
        {
            case IContainer container when orientation == Orientation.Horizontal:
                Draw(container, position, defaultSize);
                return position with
                {
                    Left = position.Left + defaultSize.Width
                };
            case IContainer container when orientation == Orientation.Vertical:
                Draw(container, position, defaultSize);
                return position with
                {
                    Top = position.Top + defaultSize.Height
                };
            case IComponent component when orientation == Orientation.Horizontal:
                var componentWidth = _componentCraftsman.Draw(component, position, defaultSize).Width;
                return position with
                {
                    Left = position.Left + (defaultSize.Width <= componentWidth ? componentWidth : defaultSize.Width)
                };
            case IComponent component when orientation == Orientation.Vertical:
                var componentHeight = _componentCraftsman.Draw(component, position, defaultSize).Height;
                return position with
                {
                    Top = position.Top + (defaultSize.Height <= componentHeight ? componentHeight : defaultSize.Height)
                };
            default:
                throw new InvalidCastException();
        }
    }
}