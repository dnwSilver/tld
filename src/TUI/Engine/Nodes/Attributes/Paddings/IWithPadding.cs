using TUI.Engine.Theme;

namespace TUI.Engine.Nodes.Attributes.Paddings;

public interface IWithPadding<out TNode> where TNode : INode
{
    public Padding? Padding { get; }

    public TNode Set(Level padding);

    public TNode Set(
        Level? left = Level.None,
        Level? top = Level.None,
        Level? right = Level.None,
        Level? bottom = Level.None
    );
}