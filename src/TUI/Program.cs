using TUI.Pages;


Console.Clear();
Console.CursorVisible = false;

// var settings = Settings.Init();

var welcomePage = new WelcomePage();
welcomePage.Open();
Thread.Sleep(500);

var dependenciesPage = new DependenciesPage();
dependenciesPage.Open();


ConsoleKeyInfo? key = null;

var waitCommand = true;
do
{
    if (key?.Key == ConsoleKey.Q)
    {
        waitCommand = false;
        continue;
    }
    
    if (key?.Key == ConsoleKey.R)
    {
        dependenciesPage.LoadDependencies();
        key = null;
        continue;
    }
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