namespace TUI.Engine.Rendering;

public class ConsoleCanvas : ICanvas
{
    public int Width => Console.WindowWidth;
    public int Height => Console.WindowHeight;
    public void SetPencil(int left, int top) => Console.SetCursorPosition(left, top);
    public void Paint(string value) => Console.Write(value);
}