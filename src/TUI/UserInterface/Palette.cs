using Pastel;


namespace TUI.UserInterface;

public static class Palette
{
    public const string HoverColor = "292928";
    public const string PrimaryColor = "84BA64";
    public const string HintColor = "71797E";
    public static string Primary(this string currentText) => currentText.Pastel(PrimaryColor);
    public static string Hint(this string currentText) => currentText.Pastel(HintColor);
}