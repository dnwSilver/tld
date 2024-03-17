using TUI.Engine.Theme;

namespace TUI.Controls.Components;

public static class VersionTypeExtensions
{
    public static string Colorize(this VersionType versionType, string value) =>
        versionType switch
        {
            VersionType.TooOld => value.Warning(),
            VersionType.ToNew => value.Info(),
            VersionType.SoGood => value.Hint(),
            VersionType.BeNice => value.Main(),
            VersionType.Convention => value.Main(),
            _ => value
        };
}