using TUI.Engine.Nodes;

namespace TUI.Engine.Attributes;

public readonly record struct Size(int Width, int Height)
{
    public static Size operator -(Size a, Position b) => new(a.Width - b.Left, a.Height - b.Top);

    public override string ToString() => $"W[{Width}] H[{Height}]";
}