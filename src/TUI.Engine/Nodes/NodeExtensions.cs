using TUI.Engine.Attributes;
using TUI.Engine.Attributes.Resizings;
using TUI.Engine.Containers;

namespace TUI.Engine.Nodes;

internal static class NodeExtensions
{
    internal static Size GetSize(this INode node, IContainer parentContainer, int nodeNumber, Size allowableSize)
    {
        var width = node.GetWidth(parentContainer, allowableSize.Width);
        var height = node.GetHeight(parentContainer, allowableSize.Height, nodeNumber);
        return new Size(width, height);
    }
}