using TUI.Engine.Theme;

namespace TUI.Engine.Nodes.Attributes.Paddings;

public interface IPaddingable
{
    Padding? Padding { get; }

    void SetPadding(Level level);
    void SetPaddingLeft(Level level);
    void SetPaddingTop(Level level);
    void SetPaddingBottom(Level level);
    void SetPaddingRight(Level level);
}