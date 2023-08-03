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
        return new string(symbol, repeatCount);
    }

    public static int Width(this string text)
    {
        var clearText = Regex.Replace(text, @"\S\[(\d{0,3}[;m][_]?){0,5}", "");
        var stringInfo = new System.Globalization.StringInfo(clearText);
        return stringInfo.LengthInTextElements;
    }
}