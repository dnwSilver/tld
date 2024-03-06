namespace TUI.Engine.Nodes.Attributes.Alignments;

public interface IWithAlignment<out TNode> where TNode : INode
{
    public Alignment Alignment { get; }

    public TNode Set(Horizontal horizontal = Horizontal.Left, Vertical vertical = Vertical.Top);
}