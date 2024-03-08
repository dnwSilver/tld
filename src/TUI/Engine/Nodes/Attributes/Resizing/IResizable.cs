using TUI.Engine.Nodes.Attributes.Orientations;

namespace TUI.Engine.Nodes.Attributes.Resizing;

public interface IResizable
{
    Resizing ResizingHorizontal { get; }

    Resizing ResizingVertical { get; }

    void SetAdaptive(Orientation orientation);
    void SetHug(Orientation orientation);
    void SetFixed(Orientation orientation, int value);
}