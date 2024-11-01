using TUI.Engine;


namespace TUI.UserInterface;

public static class Icons
{
    public static readonly Dictionary<string, string> Applications = new()
    {
        { Symbols.NpmPackage.Colorized(), "package" },
        { Symbols.DockerImage.Colorized(), "image" },
        { Symbols.Site.Colorized(), "site" },
        { Symbols.Api, "api" }
    };
}