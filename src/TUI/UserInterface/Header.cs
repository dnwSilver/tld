using TUI.Controls;


namespace TUI.UserInterface;

public class Header : IControl
{
    public const int LogoWidth = 16;
    public const int Height = 6;

    public void Render(Position position)
    {
        Console.SetCursorPosition(position.Left, position.Top);

        for (var i = 1; i <= Height; i++)
        {
            Console.WriteLine(new string(' ', Console.WindowWidth - LogoWidth));
        }

        RenderHints();
        RenderHotkeys();
        RenderLogo();
    }

    private void RenderHotkeys()
    {
        Console.SetCursorPosition(30, Theme.Padding);
        Console.Write("󰬌 toggle header".Hint());
        Console.SetCursorPosition(30, Theme.Padding + 1);
        Console.Write("󰬘 quit".Hint());
        Console.SetCursorPosition(30, Theme.Padding + 2);
        Console.Write(" select previous".Hint());
        Console.SetCursorPosition(30, Theme.Padding + 3);
        Console.Write(" select next".Hint());
    }

    private void RenderHints()
    {
        Console.SetCursorPosition(0, Theme.Padding);
        Console.WriteLine(' '.Repeat(Theme.Padding) + "󰎔  Too new: ".Hint() + "1.20.0".Info());
        Console.WriteLine(' '.Repeat(Theme.Padding) + "  So good: ".Hint() + "1.20.0".Primary());
        Console.WriteLine(' '.Repeat(Theme.Padding) + "  Be nice: ".Hint() + "1.20.0".Warning());
        Console.WriteLine(' '.Repeat(Theme.Padding) + "󰬟  Too old: ".Hint() + "1.20.0".Error());
    }

    private void RenderLogo()
    {
        Console.SetCursorPosition(Console.WindowWidth - LogoWidth - Theme.Padding, 0);
        Console.WriteLine("  ╭━━━━┳╮".Primary() + "╱╱".Hint() + "╭━━━╮ ".Primary());
        Console.SetCursorPosition(Console.WindowWidth - LogoWidth - Theme.Padding, 1);
        Console.WriteLine("  ┃╭╮╭╮┃┃".Primary() + "╱╱".Hint() + "╰╮╭╮┃ ".Primary());
        Console.SetCursorPosition(Console.WindowWidth - LogoWidth - Theme.Padding, 2);
        Console.WriteLine("  ╰╯┃┃╰┫┃".Primary() + "╱╱╱".Hint() + "┃┃┃┃ ".Primary());
        Console.SetCursorPosition(Console.WindowWidth - LogoWidth - Theme.Padding, 3);
        Console.WriteLine("  ╱╱".Hint() + "┃┃".Primary() + "╱".Hint() + "┃┃".Primary() +
                          "╱".Hint() + "╭╮┃┃┃┃ ".Primary());
        Console.SetCursorPosition(Console.WindowWidth - LogoWidth - Theme.Padding, 4);
        Console.WriteLine(" ╱╱╱".Hint() + "┃┃".Primary() + "╱".Hint() + "┃╰━╯┣╯╰╯┃ ".Primary());
        Console.SetCursorPosition(Console.WindowWidth - LogoWidth - Theme.Padding, 5);
        Console.WriteLine("╱╱╱╱".Hint() + "╰╯".Primary() + "╱".Hint() + "╰━━━┻━━━╯ ".Primary());
    }
}