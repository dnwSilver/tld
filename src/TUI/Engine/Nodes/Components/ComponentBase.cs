using TUI.Engine.Nodes.Attributes.Alignments;
using TUI.Engine.Nodes.Attributes.Paddings;
using TUI.Engine.Theme;

namespace TUI.Engine.Nodes.Components;

public abstract class ComponentBase : List<IComponent>, IComponent
{
    public abstract Content Render();

    #region Alignments

    public Alignment Alignment { get; private set; } = new(Horizontal.Center, Vertical.Top);

    public IComponent Set(Horizontal horizontal = Horizontal.Left, Vertical vertical = Vertical.Top)
    {
        Alignment = new Alignment(horizontal, vertical);
        return this;
    }

    #endregion

    #region Paddings

    public Padding Padding { get; private set; } = new(Level.None);

    public IComponent Set(Level padding)
    {
        Padding = new Padding(padding);
        return this;
    }

    public IComponent Set(
        Level? left = Level.None,
        Level? top = Level.None,
        Level? right = Level.None,
        Level? bottom = Level.None
    )
    {
        Padding = new Padding(left, top, right, bottom);
        return this;
    }

    #endregion
}