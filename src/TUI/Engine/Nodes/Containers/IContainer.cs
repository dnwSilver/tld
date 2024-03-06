namespace TUI.Engine.Nodes.Containers;

public interface IContainer : INode
{
    public Orientation Orientation { get; }

    public Nodes Nodes { get; }
}