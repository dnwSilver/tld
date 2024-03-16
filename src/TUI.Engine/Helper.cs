using System.Runtime.CompilerServices;
using Pastel;
using TUI.Engine.Attributes;
using TUI.Engine.Nodes;

[assembly: InternalsVisibleTo("TUI.Engine.Tests", AllInternalsVisible = true)]

namespace TUI.Engine;

public static class Helper
{
    private static readonly Queue<ConsoleColor> Colors = new();

    private static void Init()
    {
        Colors.Enqueue(ConsoleColor.DarkYellow);
        Colors.Enqueue(ConsoleColor.DarkMagenta);
        Colors.Enqueue(ConsoleColor.DarkGreen);
        Colors.Enqueue(ConsoleColor.DarkCyan);
        Colors.Enqueue(ConsoleColor.DarkBlue);
        Colors.Enqueue(ConsoleColor.DarkRed);
        Colors.Enqueue(ConsoleColor.Cyan);
        Colors.Enqueue(ConsoleColor.Yellow);
    }

    public static void ShowBackground(Position position, Size size)
    {
        // return;
        if (!Colors.Any())
        {
            Init();
        }

        var color = Colors.Dequeue();

        var top = position.Top;
        var height = 0;

        while (height < size.Height)
        {
            var left = position.Left;
            var width = 0;
            while (width < size.Width)
            {
                Console.SetCursorPosition(left, top);
                Console.Write(Symbols.Space.ToString().PastelBg(color));
                width++;
                left++;
            }

            height++;
            top++;
        }
    }
}