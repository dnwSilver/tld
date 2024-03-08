namespace TUI.Engine.Nodes.Attributes.Resizings;

public interface IWithResizing
{
    Resizing ResizingHorizontal { get; }

    Resizing ResizingVertical { get; }

    Size Fixed { get; }
}