using TUI.Engine.Theme;

namespace TUI.Controls.Components;

public static class VersionExtensions
{
    public static string ToImage(this VersionType versionType)
        =>
            versionType switch
            {
                VersionType.Alpha => "󰀫",
                VersionType.Beta => "󰂡",
                VersionType.Candidate => "󰑣",
                VersionType.Canary => "󱗆",
                VersionType.Next => "󰒭",
                VersionType.Unstable => "󱓉",
                _ => ""
            };

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