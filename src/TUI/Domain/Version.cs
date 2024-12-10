using TUI.Controls.Components;
using TUI.Engine;

namespace TUI.Domain;

public class Version
{
    public readonly int Major;
    public readonly int Minor;
    public readonly int Patch;
    public readonly VersionType Type;

    public Version(string version)
    {
        var parts = version.Split('.');

        if (parts.Length == 0)
        {
            return;
        }

        Major = Convert.ToInt32(parts[0].RemoveVersionPrefix());

        if (parts.Length == 1)
        {
            return;
        }

        Minor = Convert.ToInt32(parts[1]);

        if (parts.Length == 2)
        {
            return;
        }

        Patch = Convert.ToInt32(string.Join("", parts[2].TakeWhile(char.IsDigit)));

        var extension = parts[2].Replace(Patch.ToString(), "");

        Type = extension switch
        {
            not null when extension.Contains("rc") => VersionType.Candidate,
            not null when extension.Contains("beta") => VersionType.Beta,
            not null when extension.Contains("alpha") => VersionType.Alpha,
            not null when extension.Contains("canary") => VersionType.Canary,
            not null when extension.Contains("next") => VersionType.Next,
            _ => VersionType.Release
        };
    }
}