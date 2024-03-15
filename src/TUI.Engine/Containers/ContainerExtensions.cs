using TUI.Engine.Attributes.Resizings;
using TUI.Engine.Nodes;

namespace TUI.Engine.Containers;

internal static class ContainerExtensions
{
    internal static IEnumerable<INode> GetFixedNodes(this IContainer container, int? takeNodeNumber = null)
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
}