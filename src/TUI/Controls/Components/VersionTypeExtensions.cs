using TUI.Engine.Theme;

namespace TUI.Controls.Components;

public static class VersionTypeExtensions
{
    public static string Colorize(this VersionStatus versionStatus, string value) =>
        versionStatus switch
        {
            VersionStatus.TooOld => value.Warning(),
            VersionStatus.ToNew => value.Info(),
            VersionStatus.SoGood => value.Hint(),
            VersionStatus.BeNice => value.Main(),
            _ => value
        };
}