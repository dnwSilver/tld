using TUI.Controls;
using TUI.Domain;


Console.Clear();
Console.CursorVisible = false;

var settings = Settings.Init();

var display = new Display();
display.OpenDeps(settings.Projects[0]);

var key = new ConsoleKeyInfo('1', ConsoleKey.NoName, false, false, false);
var waitCommand = true;
do
{
    if (key.Key == ConsoleKey.Q && !display.CommandLine.IsFocused)
    {
        waitCommand = false;
        continue;
    }

    if (display.CommandLine.IsFocused)
    {
        switch (key.Key)
        {
            case ConsoleKey.Escape:
                display.CommandLine.IsFocused = false;
                break;
            default:
                Console.Write(key.KeyChar);

                break;
        }
    }
    else
    {
        switch (key.KeyChar)
        {
            case ':':
                display.OpenCommandLine();
                break;
        }

        switch (key.Key)
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
    }

    key = Console.ReadKey(true);
} while (waitCommand);

Console.Clear();