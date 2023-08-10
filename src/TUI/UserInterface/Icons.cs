using Pastel;


namespace TUI.UserInterface;

public static class Icons
{
    public static string GitLab => GetIcon("",         "E24329");
    public static string GitHub => GetIcon("",         "ADBAC7");
    public static string Git => GetIcon("",            "F14E32");
    public static string NetworkPublic => GetIcon("󰞉", "00FFFF");
    public static string NetworkPrivate => GetIcon("󰕑");
    public static string Undefined => GetIcon("");
    public static string Site => GetIcon("",         "BF40BF");
    public static string Api => GetIcon("",          "7F52FF");
    public static string DockerImage => GetIcon("󰡨", "086DD7");
    public static string NpmPackage => GetIcon("",   "CB0000");
    public static string SEO => GetIcon("󰚩",         "4285F4");
    public static string Auth => GetIcon("",         "FFD700");
    public static string NotFound => GetIcon("");

    public readonly static Dictionary<string, string> Applications = new()
    {
        { NpmPackage, "package" },
        { DockerImage, "image" },
        { Site, "site" },
        { Api, "api" },
    };

    private static string GetIcon(string icon, string? activeColor = null) =>
            icon.Pastel(activeColor ?? Palette.HintColor);
}