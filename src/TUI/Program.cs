using TUI.Controls;
using Settings = TUI.Domain.Settings;


Console.Clear();
Console.CursorVisible = false;

var settings = Settings.Init();

var display = new Display();
display.OpenDeps(settings.Projects[0]);

var hotKey = ConsoleKey.NoName;
do
{
    switch (hotKey)
    {
        case ConsoleKey.DownArrow:
            display.Next();
            break;
        case ConsoleKey.UpArrow:
            display.Previous();
            break;
        case ConsoleKey.E:
            display.ToggleHeader();
            break;
    }

    hotKey = Console.ReadKey(intercept: true).Key;
} while (hotKey != ConsoleKey.Q);

Console.Clear();