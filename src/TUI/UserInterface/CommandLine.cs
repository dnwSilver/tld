using TUI.Controls;


namespace TUI.UserInterface;

public class CommandLine : Dashboard
{
    public const int Height = 3;

    public void Render(Position position)
    {
        base.Render("Command", position, Height);

        Console.SetCursorPosition(position.Left + Theme.BorderWidth + Theme.Padding, position.Top + Theme.BorderWidth);
        Console.Write(">");
    }
}