using TUI.Engine.Attributes.Alignments;
using TUI.Engine.Attributes.Orientations;
using TUI.Engine.Attributes.Resizings;

namespace TUI.Engine.Theme;

public static class Defaults
{
    public const Horizontal HorizontalAlignment = Horizontal.Center;

    public const Vertical VerticalAlignment = Vertical.Top;

    public const Level Padding = Level.None;

    public const Resizing HorizontalResizing = Resizing.Adaptive;

    public const Resizing VerticalResizing = Resizing.Adaptive;

    public const Orientation Orientation = TUI.Engine.Attributes.Orientations.Orientation.Horizontal;
}