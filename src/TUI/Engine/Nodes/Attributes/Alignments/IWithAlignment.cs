namespace TUI.Engine.Nodes.Attributes.Alignments;

public interface IWithAlignment<out TNode> where TNode : INode
{
    public Alignment Alignment { get; }

    public TNode Set(AlignmentHorizontal alignmentHorizontal = AlignmentHorizontal.Left, Vertical vertical = Vertical.Top);
}