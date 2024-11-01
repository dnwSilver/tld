using TUI.Engine.Theme;

namespace TUI.Engine.Attributes.Paddings;

public interface IWithPadding
{
    Padding Padding { get; }

    void SetPadding(Level level);
    void SetPaddingLeft(Level level);
    void SetPaddingTop(Level level);
    void SetPaddingBottom(Level level);
    void SetPaddingRight(Level level);
}