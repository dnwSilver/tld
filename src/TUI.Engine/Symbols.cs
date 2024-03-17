using Pastel;
using TUI.Engine.Theme;

namespace TUI.Engine;

public static class Symbols
{
    public const string Space = " ";
    public const string Copyright = "©";
    public const string GitLab = "";
    public const string GitHub = "";
    public const string Git = "";
    public const string LineBreak = "\n";
    public const string NetworkPublic = "󰞉";
    public const string NetworkPrivate = "󰕑";
    public const string Undefined = "";
    public const string Site = "";
    public const string Api = "";
    public const string DockerImage = "";
    public const string NpmPackage = "";
    public const string SEO = "󰚩";
    public const string Auth = "";
    public const string NotFound = "";

    public static class Lines
    {
        public const string Vertical = "│";
        public const string Horizontal = "─";
    }

    public static class Angles
    {
        public const string RightTop = "┐";
        public const string LeftBottom = "└";
        public const string LeftTop = "┌";
        public const string RightBottom = "┘";
    }
}

public static class CharExtensions
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
        { Symbols.SEO, "4285F4" },
        { Symbols.Auth, "FFD700" },
    };
}