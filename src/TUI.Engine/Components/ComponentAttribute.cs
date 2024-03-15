using TUI.Engine.Attributes.Alignments;
using TUI.Engine.Attributes.Paddings;
using TUI.Engine.Nodes;
using TUI.Engine.Theme;

namespace TUI.Engine.Components;

public abstract class ComponentAttribute : NodeBase, IComponent
{
    protected abstract Sketch DrawComponent();

    Sketch IComponent.MakeSketch() => DrawComponent();

    #region Alignments

    internal Alignment Alignment { get; private set; } = new(Defaults.HorizontalAlignment, Defaults.VerticalAlignment);

    Alignment IWithAlignment.Alignment => Alignment;

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

    internal Padding Padding { get; private set; } = new(Defaults.Padding);
    

    Padding IWithPadding.Padding => Padding;

    public void SetPadding(Level level) => Padding = new Padding(level);

    public void SetPaddingTop(Level level) => Padding = Padding with { Top = level };

    public void SetPaddingRight(Level level) => Padding = Padding with { Right = level };

    public void SetPaddingBottom(Level level) => Padding = Padding with { Bottom = level };

    public void SetPaddingLeft(Level level) => Padding = Padding with { Left = level };

    #endregion
}