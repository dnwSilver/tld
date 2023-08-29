using System.Text;
using TUI.UserInterface;


namespace TUI.Controls;

public class Dashboard : IControl<string>
{
    public bool IsFocused { get; set; }

    public void Render(string title, Position position, int? height = 0)
    {
        Console.SetCursorPosition(position.Left, position.Top);

        RenderTopLine(title, IsFocused);

        var marginTop = position.Top;

        var dashboardHeight = height == 0 ? Console.WindowHeight - marginTop : height + Theme.Padding * 2;

        for (var top = marginTop;
             top < dashboardHeight + marginTop - Theme.BorderWidth * 2 - Theme.Padding * 2;
             top++)
            RenderMiddleLine(IsFocused);

        RenderBottomLine(IsFocused);
    }

    private static void RenderMiddleLine(bool isFocused)
    {
        Console.Write("│".Primary(isFocused));
        Console.Write(new string(' ', Console.WindowWidth - Theme.BorderWidth * 2));
        Console.WriteLine("│".Primary(isFocused));
    }

    private static void RenderBottomLine(bool isFocused)
    {
        var lineWidth = Console.WindowWidth - Theme.BorderWidth * 2;
        Console.Write("└".Primary(isFocused));
        Console.Write('─'.Repeat(lineWidth).Primary(isFocused));
        Console.WriteLine("┘".Primary(isFocused));
    }

    private static void RenderTopLine(string title, bool isFocused)
    {
        var lineWidth =
                (Console.WindowWidth - title.Width() - Theme.BorderWidth * 2 - Theme.Padding * 2) /
                2;

        var topLine = new StringBuilder();
        topLine.Append("┌");
        topLine.Append('─'.Repeat(lineWidth));
        topLine.AppendFormat("{0}{1}{0}", ' '.Repeat(Theme.Padding), title);
        if (title.Width() % 2 == 1) topLine.Append('─');

        topLine.Append('─'.Repeat(lineWidth));
        topLine.Append("┐");
        Console.WriteLine(topLine.ToString().Primary(isFocused));
    }
}