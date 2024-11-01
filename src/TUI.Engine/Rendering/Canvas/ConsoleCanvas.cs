using TUI.Engine.Attributes;
using TUI.Engine.Nodes;

namespace TUI.Engine.Rendering.Canvas;

public class ConsoleCanvas : ICanvas
{
    private readonly DrawCraftsman _drawCraftsman;

    public Size Size { get; } = new(Console.WindowWidth, Console.WindowHeight);

    public ConsoleCanvas()
    {
        var componentCraftsman = new ComponentCraftsman(this);
        var containerCraftsman = new ContainerCraftsman(componentCraftsman);
        _drawCraftsman = new DrawCraftsman(componentCraftsman, containerCraftsman);
    }

    public void SetPencil(Position pencilPosition)
    {
        Console.SetCursorPosition(pencilPosition.Left, pencilPosition.Top);
    }

    public void Paint(string value) => Console.Write(value);

    public void Draw(INode node)
    {
        _drawCraftsman.Draw(node, Position.Default, Size);
    }

    public void Draw(INode node, Position pencil, Size maxSize)
    {
        _drawCraftsman.Draw(node, pencil, maxSize);
    }
}