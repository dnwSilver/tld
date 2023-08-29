using Pastel;


namespace TUI.UserInterface;

public static class Palette
{
    public const string HoverColor = "292928";
    public const string PrimaryColor = "84BA64";
    public const string HintColor = "71797E";
    public const string ErrorColor = "CA3433";
    public const string WarningColor = "EC9706";
    public const string InfoColor = "25799F";

    public static string Primary(this string currentText, bool isFocused = true)
    {
        return isFocused
                ? currentText.Pastel(PrimaryColor)
                : Hint(currentText);
    }

    public static string Hint(this string currentText)
    {
        return currentText.Pastel(HintColor);
    }

    public static string Disable(this string currentText)
    {
        return currentText.RemoveColors().Pastel(HintColor);
    }

    public static string Warning(this string currentText)
    {
        return currentText.Pastel(WarningColor);
    }

    public static string Error(this string currentText)
    {
        return currentText.Pastel(ErrorColor);
    }

    public static string Info(this string currentText)
    {
        return currentText.Pastel(InfoColor);
    }
}