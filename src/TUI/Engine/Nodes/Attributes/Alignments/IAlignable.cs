namespace TUI.Engine.Nodes.Attributes.Alignments;

public interface IAlignable
{
    Alignment Alignment { get; }

    void SetAlignment(Vertical vertical);
    
    void SetAlignment(Horizontal horizontal);
}