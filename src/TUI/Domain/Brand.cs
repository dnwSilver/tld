using Pastel;

namespace TUI.Domain;

public record Brand(string Name, string? Logo = null, string? Color = null)
{
    public string ColorLogo() => Logo?.Pastel(Color) ?? string.Empty;
};