namespace TUI.Engine.Rendering;

public class ConsoleWindow : IWindow
{
    public int Width => Console.WindowWidth;
    public int Height => Console.WindowHeight;
    public void SetCursorPosition(int left, int top) => Console.SetCursorPosition(left, top);
    public void Write(string value) => Console.Write(value);
}