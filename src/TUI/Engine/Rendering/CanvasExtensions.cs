using TUI.Engine.Nodes.Attributes;

namespace TUI.Engine.Rendering;

public static class CanvasExtensions
{
    public static Size GetSize(this ICanvas canvas) => new(canvas.Width, canvas.Height);
}