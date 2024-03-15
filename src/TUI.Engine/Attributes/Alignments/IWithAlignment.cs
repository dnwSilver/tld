namespace TUI.Engine.Attributes.Alignments;

public interface IWithAlignment
{
    internal Alignment Alignment { get; }

    void SetAlignment(Vertical vertical);
    
    void SetAlignment(Horizontal horizontal);
}