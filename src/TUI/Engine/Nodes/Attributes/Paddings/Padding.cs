using TUI.Engine.Theme;

namespace TUI.Engine.Nodes.Attributes.Paddings;

public record Padding(
    Level? Left = Level.None,
    Level? Top = Level.None,
    Level? Right = Level.None,
    Level? Bottom = Level.None
)
{
    public Padding(Level padding) : this(Left: padding, Top: padding, Right: padding, Bottom: padding)
    {
    }
}