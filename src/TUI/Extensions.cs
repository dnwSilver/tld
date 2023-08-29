using System.Globalization;
using System.Text.RegularExpressions;


namespace TUI;

public static class Extensions
{
    public static bool Have(this IEnumerable<string> array, string findValue)
    {
        return array.Any(item => item == findValue);
    }

    public static string Repeat(this char symbol, int repeatCount)
    {
        return repeatCount < 0 ? "" : new string(symbol, repeatCount);
    }

    public static string RemoveColors(this string text)
    {
        return Regex.Replace(text, @"\S\[(\d{0,3}[;m]_?){0,5}", "");
    }

    public static int Width(this string text)
    {
        if (string.IsNullOrEmpty(text)) return 0;

        var clearText = text.RemoveColors();
        var stringInfo = new StringInfo(clearText);
        return stringInfo.LengthInTextElements;
    }

    public static Version? ToVersion(this string textVersion)
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