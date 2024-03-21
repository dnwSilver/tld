using Pastel;

namespace TUI.Controls.Components;

public record Brand(string Name, string Logo, string Color)
{
    public string ColorLogo() => Logo.Pastel(Color);
};