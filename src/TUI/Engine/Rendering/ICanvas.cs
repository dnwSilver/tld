using TUI.Engine.Nodes.Attributes;

namespace TUI.Engine.Rendering;

public interface ICanvas
{
    int Width { get; }
    int Height { get; }
    void SetPencil(int left, int top);
    void Paint(string value);
}