namespace TUI.Engine.Nodes.Attributes;

public record Size(int Width, int Height)
{
    public override string ToString() => $"W[{Width}] H[{Height}]";
}