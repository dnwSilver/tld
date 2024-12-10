namespace TUI.Logs;

enum Level
{
    Trace = 60,
    Debug = 50,
    Info = 40,
    Warning = 30,
    Error = 20,
    Fatal = 10,
}

public static class Log
{
    private static readonly int LogLevel = (int)Level.Info;
    private static string Now => DateTime.Now.ToString("yyyy-MM-dd HH:mm:ss");
    public static void Trace(string message)
    {
        if (LogLevel < 60) return;
        Write("🌚", message);
    }

    public static void Debug(string message)
    {
        if (LogLevel < 50) return;
        Write("🦎", message);
    }

    public static void Info(string message)
    {
        if (LogLevel < 40) return;
        Write("🦋", message);
    }

    public static void Warning(string message)
    {
        if (LogLevel < 30) return;
        Write("🍋", message);
    }

    public static void Error(string message)
    {
        if (LogLevel < 20) return;
        Write("🐞", message);
    }

    public static void Fatal(string message)
    {
        if (LogLevel < 10) return;
        Write("💀", message);
    }

    public static void Write(string icon, string message)
    {
        // /tld/src/TUI/file.log
        var file = "file.log";
        var line = string.Join('|', Now, icon, message);
        File.AppendAllText(file, line + Environment.NewLine);
    }
}