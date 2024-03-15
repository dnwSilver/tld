using TUI.Engine.Attributes;
using TUI.Engine.Attributes.Alignments;
using TUI.Engine.Nodes;

namespace TUI.Engine.Components;

internal static class ComponentExtensions
{
    internal static Position CorrectContentPosition(this IComponent component,
        Position pencil,
        Size maxSize,
        Size sketchSize)
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