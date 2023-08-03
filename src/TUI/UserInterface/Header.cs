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

        RenderLogo();
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