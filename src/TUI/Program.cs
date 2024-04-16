using TUI.Pages;


Console.Clear();
Console.CursorVisible = false;

// var settings = Settings.Init();

var welcomePage = new WelcomePage();
welcomePage.Open();
Thread.Sleep(2000);

var dependenciesPage = new DependenciesPage();
dependenciesPage.Open();
// display.OpenDeps(settings.Projects[0]);

var key = new ConsoleKeyInfo('1', ConsoleKey.NoName, false, false, false);
var waitCommand = true;
do
{
    // if (key.Key == ConsoleKey.Q && !display.CommandLine.IsFocused)
    // {
    //     waitCommand = false;
    //     continue;
    // }
    //
    // if (display.CommandLine.IsFocused)
    // {
    //     switch (key.Key)
    //     {
    //         case ConsoleKey.Escape:
    //             display.CommandLine.IsFocused = false;
    //             break;
    //         default:
    //             Console.Write(key.KeyChar);
    //
    //             break;
    //     }
    // }
    // else
    // {
    //     switch (key.KeyChar)
    //     {
    //         case ':':
    //             display.OpenCommandLine();
    //             break;
    //     }
    //
    //     switch (key.Key)
    //     {
    //         case ConsoleKey.DownArrow:
    //             display.Next();
    //             break;
    //         case ConsoleKey.UpArrow:
    //             display.Previous();
    //             break;
    //         case ConsoleKey.E:
    //             display.Toggle();
    //             break;
    //     }
    // }

    key = Console.ReadKey(true);
} while (waitCommand);

Console.Clear();