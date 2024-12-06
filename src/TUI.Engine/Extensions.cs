using System.Globalization;
using System.Text;
using System.Text.RegularExpressions;

namespace TUI.Engine;

public static class Extensions
{
    public static int Max(this int value, int maxValue)
    {
        return value <= maxValue ? value : maxValue;
    }

    public static int Min(this int value, int minValue)
    {
        return value > minValue ? value : minValue;
    }

    public static bool Have(this IEnumerable<string> array, string findValue)
    {
        return array.Any(item => item == findValue);
    }

    public static string Repeat(this string value, int count)
    {
        return count < 0 ? string.Empty : new StringBuilder(value.Length * count).Insert(0, value, count).ToString();
    }

    public static string RemoveColors(this string text)
    {
        return Regex.Replace(text, @"\S\[(\d{0,3}[;m]_?){0,5}", "");
    }

    public static int GetWidth(this string text)
    {
        if (string.IsNullOrEmpty(text)) return 0;

        var clearText = text.RemoveColors();
        var stringInfo = new StringInfo(clearText);
        return stringInfo.LengthInTextElements;
    }

    public static Version ToVersion(this string textVersion)
    {
        var version = textVersion.Replace("^", "").Replace("v", "").Replace("~", "").Split(".");
        var major = Convert.ToInt32(version[0]);
        var minor = Convert.ToInt32(version[1]);
        var patch = Convert.ToInt32(version[2].Split('-')[0]);
        return new Version(major, minor, patch);
    }
}
