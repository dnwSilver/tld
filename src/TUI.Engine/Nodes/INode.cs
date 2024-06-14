using TUI.Engine.Attributes;
using TUI.Engine.Attributes.Resizings;
using TUI.Engine.Rendering.Canvas;

namespace TUI.Engine.Nodes;

public interface INode : IResizable
{
    DrawContext DrawContext { get; set; }
    StyleContext StyleContext { get; set; }
}

public record DrawContext(ICanvas Canvas, Position Pencil, Size MaxSize);

public record StyleContext(string? Foreground, string? Background = null);