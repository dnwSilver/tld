namespace TUI.Engine.Nodes;

public record Position(int Left, int Top)
{
    public static readonly Position Default = new(0, 0);
    public override string ToString() => $"L[{Left}] T[{Top}]";
}