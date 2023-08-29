using TUI.UserInterface;


namespace TUI.Controls;

public class Copyright : IControl<string>
{
    public void Render(string author, Position position, int? height = 0)
    {
        const string icon = "© ";
        Console.SetCursorPosition(position.Left - icon.Width(), position.Top);

        var copyright = $"{icon}{author}".Hint();
        Console.Write(copyright);
    }
}