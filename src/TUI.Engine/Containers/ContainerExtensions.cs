using TUI.Engine.Attributes.Resizings;
using TUI.Engine.Components;
using TUI.Engine.Nodes;

namespace TUI.Engine.Containers;

internal static class ContainerExtensions
{
    private static readonly Func<INode, bool> AbsoluteNodes = node => node is IComponent { IsRelative: false };

    private static readonly Func<INode, bool> FixedNodes = node => node.ResizingVertical == Resizing.Fixed;

    internal static IEnumerable<INode> GetAbsoluteNodes(this IContainer container, int? takeNodeNumber = null)
    {
        if (takeNodeNumber is not null)
        {
            return container
                .GetNodes()
                .Take(takeNodeNumber.Value + 1)
                .Where(AbsoluteNodes);
        }

        return container
            .GetNodes()
            .Where(AbsoluteNodes);
    }

    internal static IEnumerable<INode> GetFixedNodes(this IContainer container, int? takeNodeNumber = null)
    {
        if (takeNodeNumber is not null)
        {
            return container
                .GetNodes()
                .Take(takeNodeNumber.Value + 1)
                .Where(FixedNodes);
        }

        return container
            .GetNodes()
            .Where(FixedNodes);
    }
}