using TUI.Logs;
using TUI.Pages;


Console.Clear();
Console.CursorVisible = false;

Log.Info("Run application.");

var welcomePage = WelcomePage.Instance;
welcomePage.Open();
Thread.Sleep(500);

IPage currentPage = DependenciesPage.Instance;
currentPage.Open();


ConsoleKeyInfo? key = null;

var waitCommand = true;
do
{
    switch (key?.Key)
    {
        case ConsoleKey.Q:
            waitCommand = false;
            Log.Trace("Run command quit.");
            break;
        case ConsoleKey.R:
            key = null;
            Log.Trace("Run command load deps.");
            currentPage.Load();
            break;
        case ConsoleKey.D1:
            key = null;
            currentPage = DependenciesPage.Instance;
            Console.Clear();
            currentPage.Render();
            break;
        case ConsoleKey.D0:
            key = null;
            currentPage = WelcomePage.Instance;
            Console.Clear();
            currentPage.Render();
            break;
        default:
            key = Console.ReadKey(true);
            break;
    }
} while (waitCommand);

Log.Info("Quit application.");
Console.Clear();
Console.CursorVisible = true;