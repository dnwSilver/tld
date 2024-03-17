using TUI.Engine;


namespace TUI.UserInterface;

public static class Icons
{
    public static readonly Dictionary<string, string> Applications = new()
    {
        { Symbols.NpmPackage, "package" },
        { Symbols.DockerImage, "image" },
        { Symbols.Site, "site" },
        { Symbols.Api, "api" }
    };
}