using TUI.Engine.Attributes.Orientations;
using TUI.Engine.Containers;

namespace TUI.Engine.Attributes.Resizings;

internal static class ResizingExtensions
{
    internal static int GetHeight(this IResizable node, IContainer container, int maxHeight, int nodeIndex)
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

    internal static int GetWidth(this IResizable node, IContainer container, int maxWidth)
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