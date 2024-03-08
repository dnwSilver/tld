using TUI.Engine.Nodes.Attributes;
using TUI.Engine.Nodes.Attributes.Alignments;
using TUI.Engine.Nodes.Attributes.Paddings;
using TUI.Engine.Nodes.Attributes.Resizing;
using TUI.Engine.Theme;


namespace TUI.Engine.Nodes.Components;


public abstract class ComponentBase : NodeBase, IComponent
{
    private Size _sketchSize;

    public abstract Sketch DrawComponent();

    public Sketch Draw()
    {
        var sketch = DrawComponent();
        _sketchSize = sketch.GetSize();
        return sketch;
    }

    public Resizing ResizingHorizontal { get; }

    // protected override Size GetAllowableSize() =>
        // new(
            // AllowableSize.Width <= _sketchSize.Width ? _sketchSize.Width : AllowableSize.Width,
            // AllowableSize.Height <= _sketchSize.Height ? _sketchSize.Height : AllowableSize.Height
        // );

    #region Alignments

    public Alignment Alignment { get; private set; } = new(Horizontal.Center, Vertical.Top);

    public void SetAlignment(Vertical vertical)
    {
        Alignment = Alignment with { Vertical = vertical };
    }

    public void SetAlignment(Horizontal horizontal)
    {
        Alignment = Alignment with { Horizontal = horizontal };
    }

    #endregion

    #region Paddings

    public Padding Padding { get; private set; } = new(Level.None);

    public void SetPadding(Level level) => Padding = new Padding(level);

    public void SetPaddingTop(Level level) => Padding = Padding with { Top = level };

    public void SetPaddingRight(Level level) => Padding = Padding with { Right = level };

    public void SetPaddingBottom(Level level) => Padding = Padding with { Bottom = level };

    public void SetPaddingLeft(Level level) => Padding = Padding with { Left = level };

    #endregion
}