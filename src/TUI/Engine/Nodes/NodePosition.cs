namespace TUI.Engine.Nodes;

public record NodePosition(int Left, int Top)
{
    public override string ToString() => $"L[{Left}] T[{Top}]";
}