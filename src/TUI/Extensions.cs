using System.Text.Json.Nodes;
using System.Text.RegularExpressions;
using TUI.Domain;
using TUI.Settings;
using TUI.UserInterface;


namespace TUI;

public static class Extensions
{
    public static bool Have(this IEnumerable<string> array, string findValue)
    {
        return array.Any(item => item == findValue);
    }

    public static string Repeat(this char symbol, int repeatCount)
    {
        return new string(symbol, repeatCount);
    }

    public static int Width(this string text)
    {
        if (string.IsNullOrEmpty(text))
        {
            return 0;
        }

        var clearText = Regex.Replace(text, @"\S\[(\d{0,3}[;m][_]?){0,5}", "");
        var stringInfo = new System.Globalization.StringInfo(clearText);
        return stringInfo.LengthInTextElements;
    }

    public static string GetVersion(this JsonObject dependencies, DependencyDto dependencyDto)
    {
        dependencies.TryGetPropertyValue(dependencyDto.Name.ToLower(), out var version);
        var currentVersion = version?.GetValue<string>().ToVersion();
        if (currentVersion == null)
        {
            return "".Hint();
        }

        var conventionVersion = dependencyDto.Version.ToVersion();

        if (currentVersion > conventionVersion)
        {
            return currentVersion.ToString().Info();
        }

        if (currentVersion < conventionVersion)
        {
            if (currentVersion.Major == conventionVersion.Major)
            {
                return currentVersion.ToString().Warning();
            }

            return currentVersion.ToString().Error();
        }

        return currentVersion.ToString().Primary();
    }

    private static Version? ToVersion(this string textVersion)
    {
        var version = textVersion.Replace("^", "").Replace("~", "").Split(".");
        if (version.Length != 3)
            return null;
        var major = Convert.ToInt32(version[0]);
        var minor = Convert.ToInt32(version[1]);
        var patch = Convert.ToInt32(version[2].Split('-')[0]);
        return new Version(major, minor, patch);
    }
}