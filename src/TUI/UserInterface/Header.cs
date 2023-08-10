using TUI.Controls;


namespace TUI.UserInterface;

public class Header : IControl
{
    public const int LogoWidth = 16;
    public const int Height = 6;
    public const int MaxHeaderBlocksWidth = 16;


    private readonly Dictionary<string, string> _hotKeys = new()
    {
        { "", "select prev" },
        { "", "select next" },
        { "󰬌", "toggle head" },
        { "󰬘", "quit" },
    };

    private readonly Dictionary<string, string> _hints = new()
    {
        { "󰎔", "too new".Info() },
        { "", "so good" },
        { "", "be nice".Primary() },
        { "󰬟", "too old".Warning() }
    };

    private readonly Dictionary<string, string> _tags = new()
    {
        { Icons.Auth, "Auth" },
        { Icons.NetworkPublic, "WWW" },
        { Icons.SEO, "SEO" },
        { Icons.GitLab, "VCS" },
    };

    public void Render(Position position)
    {
        Console.SetCursorPosition(position.Left, position.Top);

        for (var i = 1; i <= Height; i++)
        {
            Console.WriteLine(new string(' ', Console.WindowWidth - LogoWidth));
        }

        RenderBlock(0, _hints);
        RenderBlock(1, _tags);
        RenderBlock(2, Icons.Applications);
        RenderBlock(3, _hotKeys);
        RenderLogo();
    }

    private static void RenderBlock(int blockNumber, Dictionary<string, string> items)
    {
        var leftPadding = Theme.Padding + blockNumber * MaxHeaderBlocksWidth;

        var hotKeyNumber = 0;
        foreach (var item in items)
        {
            Console.SetCursorPosition(leftPadding, Theme.Padding + hotKeyNumber++);
            Console.Write((item.Key + " " + item.Value).Hint());
        }
    }


    private static void RenderLogo()
    {
        Console.SetCursorPosition(Console.WindowWidth - LogoWidth - Theme.Padding, 0);
        Console.WriteLine("  ╭━━━━┳╮".Primary() + "╱╱".Hint() + "╭━━━╮ ".Primary());
        Console.SetCursorPosition(Console.WindowWidth - LogoWidth - Theme.Padding, 1);
        Console.WriteLine("  ┃╭╮╭╮┃┃".Primary() + "╱╱".Hint() + "╰╮╭╮┃ ".Primary());
        Console.SetCursorPosition(Console.WindowWidth - LogoWidth - Theme.Padding, 2);
        Console.WriteLine("  ╰╯┃┃╰┫┃".Primary() + "╱╱╱".Hint() + "┃┃┃┃ ".Primary());
        Console.SetCursorPosition(Console.WindowWidth - LogoWidth - Theme.Padding, 3);
        Console.WriteLine("  ╱╱".Hint() + "┃┃".Primary() + "╱".Hint() + "┃┃".Primary() + "╱".Hint() +
                          "╭╮┃┃┃┃ ".Primary());
        Console.SetCursorPosition(Console.WindowWidth - LogoWidth - Theme.Padding, 4);
        Console.WriteLine(" ╱╱╱".Hint() + "┃┃".Primary() + "╱".Hint() + "┃╰━╯┣╯╰╯┃ ".Primary());
        Console.SetCursorPosition(Console.WindowWidth - LogoWidth - Theme.Padding, 5);
        Console.WriteLine("╱╱╱╱".Hint() + "╰╯".Primary() + "╱".Hint() + "╰━━━┻━━━╯ ".Primary());
    }
}