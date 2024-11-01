using TUI.Engine.Attributes;
using TUI.Engine.Nodes;

namespace TUI.Engine.Rendering;

public interface IDrawable<in TItem> where TItem : INode
{
    Size Draw(TItem item, Position pencil, Size maxSize);
}