using TUI.Engine.Attributes;
using TUI.Engine.Nodes;

namespace TUI.Engine.Rendering.Canvas;

public interface ICanvas
{
    Size Size { get; }

    void SetPencil(Position pencilPosition);

    void Paint(string value);

    void Draw(INode node);

    void Draw(INode node, Position pencil, Size maxSize);
}