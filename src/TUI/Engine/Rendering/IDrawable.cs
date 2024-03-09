using TUI.Engine.Nodes;
using TUI.Engine.Nodes.Attributes;

namespace TUI.Engine.Rendering;

public interface IDrawable<in TItem> where TItem : INode
{
    Size Draw(TItem item, Position pencil, Size maxSize);
}