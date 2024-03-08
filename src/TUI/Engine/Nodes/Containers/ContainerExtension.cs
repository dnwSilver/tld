using TUI.Engine.Nodes.Attributes;
using TUI.Engine.Nodes.Attributes.Alignments;
using TUI.Engine.Nodes.Attributes.Orientations;
using TUI.Engine.Nodes.Components;

namespace TUI.Engine.Nodes.Containers;

public static class ContainerExtension
{
    public static Size GetSize(this IContainer container, Size allowableSize)
    {
        var nodeCount = container.GetNodes().Count;
        var width = container.Orientation == Orientation.Horizontal
            ? allowableSize.Width / nodeCount
            : allowableSize.Width;
        var height = container.Orientation == Orientation.Vertical
            ? allowableSize.Height / nodeCount
            : allowableSize.Height;

        return new Size(width, height);
    }
}

public static class ComponentExtensions
{
    public static Position GetPosition(this IComponent component, Position sketchPosition, Size allowableSize,
        Size actualSize)
    {
        var left = sketchPosition.Left + (int)(component.Padding?.Left ?? 0) +
                   CompensationLeft(component.Alignment.AlignmentHorizontal, allowableSize, actualSize);
        var top = sketchPosition.Top + (int)(component.Padding?.Top ?? 0) +
                  CompensationTop(component.Alignment.Vertical, allowableSize, actualSize);
        return new Position(left, top);
    }

    private static int CompensationLeft(AlignmentHorizontal componentAlignmentHorizontal, Size defaultSize,
        Size realSize) =>
        componentAlignmentHorizontal switch
        {
            AlignmentHorizontal.Left => 0,
            AlignmentHorizontal.Center => (defaultSize.Width - realSize.Width) / 2,
            AlignmentHorizontal.Right => defaultSize.Width - realSize.Width,
            _ => 0
        };

    private static int CompensationTop(Vertical componentVertical, Size defaultSize, Size realSize)
        =>
            componentVertical switch
            {
                Vertical.Top => 0,
                Vertical.Center => (defaultSize.Height - realSize.Height) / 2,
                Vertical.Bottom => defaultSize.Height - realSize.Height,
                _ => 0
            };
}