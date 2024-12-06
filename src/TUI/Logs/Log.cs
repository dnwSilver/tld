namespace TUI.Logs;

public static class Log
{
    private static string Now => DateTime.Now.ToString("yyyy-MM-dd HH:mm:ss");
    public static void Trace(string message) => Write("🌚", message);
    public static void Debug(string message) => Write("🦎", message);
    public static void Info(string message) => Write("🦋", message);
    public static void Warning(string message) => Write("🍋", message);
    public static void Error(string message) => Write("🐞", message);
    public static void Fatal(string message) => Write("💀", message);

    public static void Write(string icon, string message)
    {
        // /tld/src/TUI/file.log
        var file = "file.log";
        var line = string.Join('|', Now, icon, message);
        File.AppendAllText(file, line + Environment.NewLine);
    }
}