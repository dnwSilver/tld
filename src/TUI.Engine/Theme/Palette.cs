using Pastel;

namespace TUI.Engine.Theme;

public static class Palette
{
    public const string HoverColor = "292928";
    public const string PrimaryColor = "84BA64";
    public const string HintColor = "71797E";
    public const string DisableColor = "303030";
    public const string ErrorColor = "CA3433";
    public const string WarningColor = "EC9706";
    public const string InfoColor = "25799F";
    
    public static string Main(this string currentText, bool isFocused = true) =>
        isFocused
            ? currentText.Pastel(PrimaryColor)
            : Hint(currentText);
    
    public static string Hint(this string currentText) => currentText.Pastel(HintColor);
    
    public static string Hint(this char currentText) => currentText.ToString().Pastel(HintColor);
    
    public static string Disable(this string currentText) => currentText.RemoveColors().Pastel(HintColor);
    
    public static string Warning(this string currentText) => currentText.Pastel(WarningColor);
    
    public static string Error(this string currentText) => currentText.Pastel(ErrorColor);
    
    public static string Info(this string currentText) => currentText.Pastel(InfoColor);
    
    public static string Info(this char currentText) => currentText.ToString().Pastel(InfoColor);
    
    public static string Info(this int currentText) => currentText.ToString().Pastel(InfoColor);
}