using TUI.Engine.Nodes.Attributes;
using TUI.Engine.Nodes.Attributes.Alignments;
using TUI.Engine.Nodes.Attributes.Orientations;
using TUI.Engine.Nodes.Attributes.Resizing;
using TUI.Engine.Nodes.Components;

namespace TUI.Engine.Nodes.Containers;

public static class ContainerExtension
{
    public static Size GetSize(this IContainer container, Size allowableSize)
    {
        var nodeCount = container.GetNodes().Count;
        var width = container.ResizingHorizontal switch
        {
            Resizing.Adaptive => container.Orientation == Orientation.Horizontal
                ? allowableSize.Width / nodeCount
                : allowableSize.Width,
            Resizing.Fixed => container.GetFixedSize().Width,
            _ => throw new ArgumentOutOfRangeException()
        };
        var height = container.Orientation == Orientation.Vertical
            ? allowableSize.Height / nodeCount
            : allowableSize.Height;

        return new Size(width, height);
    }
}

public static class ComponentExtensions
{
    public static Position CorrectPosition(this IComponent component, Position pencil, Size maxSize, Size sketchSize)
    {
        var padding = component.Padding;
        var alignment = component.Alignment;
        var alignmentCompensationLeft = GetAlignmentCompensationLeft(alignment.Horizontal, maxSize, sketchSize);
        var alignmentCompensationTop = GetAlignmentCompensationTop(alignment.Vertical, maxSize, sketchSize);
        var left = pencil.Left + (int)padding.Left + alignmentCompensationLeft;
        var top = pencil.Top + (int)padding.Top + alignmentCompensationTop;
        return new Position(left, top);
    }

    private static int GetAlignmentCompensationLeft(Horizontal alignment, Size maxSize, Size sketchSize) =>
        alignment switch
        {
            Horizontal.Left => 0,
            Horizontal.Center => (maxSize.Width - sketchSize.Width) / 2,
            Horizontal.Right => maxSize.Width - sketchSize.Width,
            _ => 0
        };

    private static int GetAlignmentCompensationTop(Vertical alignment, Size maxSize, Size sketchSize) =>
        alignment switch
        {
            Vertical.Top => 0,
            Vertical.Center => (maxSize.Height - sketchSize.Height) / 2,
            Vertical.Bottom => maxSize.Height - sketchSize.Height,
            _ => 0
        };
}