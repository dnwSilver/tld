namespace TUI.Engine.Rendering;

public static class IntegerExtension
{
    public static int Max(this int value, int maxValue) => value <= maxValue ? value : maxValue;
    public static int Min(this int value, int minValue) => value > minValue ? value : minValue;
}