using Pastel;
using TUI.Engine.Theme;

namespace TUI.Engine;

public static class SymbolExtensions
{
    public static string Colorized(this string symbol) =>
        !SymbolColors.ContainsKey(symbol) ? symbol.Hint() : symbol.Pastel(SymbolColors[symbol]);

    private static readonly Dictionary<string, string> SymbolColors = new()
    {
        { Symbols.Git, "F14E32" },
        { Symbols.Site, "BF40BF" },
        { Symbols.GitLab, "E24329" },
        { Symbols.GitHub, "ADBAC7" },
        { Symbols.NetworkPublic, "00FFFF" },
        { Symbols.Api, "7F52FF" },
        { Symbols.DockerImage, "086DD7" },
        { Symbols.NpmPackage, "CB0000" },
        { Symbols.Seo, "4285F4" },
        { Symbols.Auth, "FFD700" },
    };
}