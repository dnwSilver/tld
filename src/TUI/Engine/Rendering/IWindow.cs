namespace TUI.Engine.Rendering;

public interface IWindow
{
    int Width { get; }
    int Height { get; }
    void SetCursorPosition(int left, int top);
    void Write(string value);
}