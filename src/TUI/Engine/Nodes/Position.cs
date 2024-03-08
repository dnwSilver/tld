namespace TUI.Engine.Nodes;

public record Position(int Left, int Top)
{
    public override string ToString() => $"L[{Left}] T[{Top}]";
}