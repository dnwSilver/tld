using System.Text;
using TUI.UserInterface;


namespace TUI.Controls;

public class Dashboard : IControl<string>
{
    public void Render(string title, Position position)
    {
        Console.SetCursorPosition(position.Left, position.Top);

        RenderTopLine(title);

        var marginTop = Theme.BorderWidth + Theme.Padding + position.Top;
        var dashboardHeight = Console.WindowHeight - Theme.BorderWidth;

        for (var top = marginTop; top < dashboardHeight; top++)
        {
            RenderMiddleLine();
        }

        RenderBottomLine();
    }

    private static void RenderMiddleLine()
    {
        Console.Write("│".Primary());
        Console.Write(new string(' ', Console.WindowWidth - Theme.BorderWidth * 2));
        Console.WriteLine("│".Primary());
    }

    private static void RenderBottomLine()
    {
        var lineWidth = Console.WindowWidth - Theme.BorderWidth * 2;
        Console.Write("└".Primary());
        Console.Write('─'.Repeat(lineWidth).Primary());
        Console.WriteLine("┘".Primary());
    }

    private static void RenderTopLine(string title)
    {
        var lineWidth =
                (Console.WindowWidth - title.Width() - Theme.BorderWidth * 2 - Theme.Padding * 2) /
                2;

        var topLine = new StringBuilder();
        topLine.Append("┌");
        topLine.Append('─'.Repeat(lineWidth));
        topLine.AppendFormat("{0}{1}{0}", ' '.Repeat(Theme.Padding), title);
        if (title.Width() % 2 == 1)
        {
            topLine.Append('─');
        }

        topLine.Append('─'.Repeat(lineWidth));
        topLine.Append("┐");
        Console.WriteLine(topLine.ToString().Primary());
    }
}