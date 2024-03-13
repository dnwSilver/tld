using TUI.Engine.Nodes.Attributes;
using TUI.Engine.Nodes.Attributes.Alignments;
using TUI.Engine.Nodes.Attributes.Orientations;
using TUI.Engine.Nodes.Attributes.Resizing;
using TUI.Engine.Nodes.Components;
using TUI.Engine.Rendering;

namespace TUI.Engine.Nodes.Containers;

public static class ContainerExtension
{
    public static Size GetSize(this INode node, IContainer parentContainer, int nodeNumber, Size allowableSize)
    {
        var width = GetWidth(node, parentContainer, allowableSize.Width);
        var height = GetHeight(node, parentContainer, allowableSize.Height, nodeNumber);
        return new Size(width, height);
    }

    private static IEnumerable<INode> GetFixedNodes(this IContainer container, int? takeNodeNumber = null)
    {
        if (takeNodeNumber is not null)
        {
            return container
                .GetNodes()
                .Take(takeNodeNumber.Value + 1)
                .Where(n => n.ResizingVertical == Resizing.Fixed);
        }

        return container
            .GetNodes()
            .Where(n => n.ResizingVertical == Resizing.Fixed);
    }

    private static int GetHeight(IResizable node, IContainer container, int maxHeight, int nodeIndex)
    {
        if (node.ResizingVertical == Resizing.Fixed)
        {
            return node.GetFixedSize().Height;
        }

        if (container.Orientation == Orientation.Horizontal)
        {
            return maxHeight;
        }

        var fixedNodes = container.GetFixedNodes().ToArray();

        var fixedHeight = fixedNodes.Sum(s => s.GetFixedSize().Height);
        var allowableHeight = maxHeight - fixedHeight;

        var allowableCount = container.GetNodes().Count - fixedNodes.Length;
        var nodeHeight = (allowableHeight / allowableCount).Min(1);
        var nodeNumber = nodeIndex + 1 - container.GetFixedNodes(nodeIndex).Sum(c => c.GetFixedSize().Height);

        if (allowableHeight - nodeNumber * nodeHeight < nodeHeight)
        {
            return allowableHeight + nodeHeight - nodeNumber * nodeHeight;
        }

        return nodeHeight;
    }

    private static int GetWidth(IResizable node, IContainer container, int maxWidth)
    {
        if (node.ResizingHorizontal == Resizing.Fixed)
        {
            return node.GetFixedSize().Width;
        }

        if (container.Orientation == Orientation.Vertical)
        {
            return maxWidth;
        }

        var fixedNodes = container
            .GetNodes()
            .Where(n => n.ResizingHorizontal == Resizing.Fixed).ToArray();

        var allowableWidth = maxWidth - fixedNodes.Sum(s => s.GetFixedSize().Width);
        var allowableCount = container.GetNodes().Count - fixedNodes.Length;

        return allowableWidth / allowableCount;
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