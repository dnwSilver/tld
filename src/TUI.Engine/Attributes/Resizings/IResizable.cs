using TUI.Engine.Attributes.Orientations;

namespace TUI.Engine.Attributes.Resizings;

public interface IResizable
{
    internal Resizing ResizingHorizontal { get; }

    internal Resizing ResizingVertical { get; }

    internal Size GetFixedSize();

    void SetAdaptive(Orientation orientation);

    void SetFixed(Orientation orientation, int value);
}