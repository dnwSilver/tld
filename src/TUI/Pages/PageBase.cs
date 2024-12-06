using System.Diagnostics;
using TUI.Logs;

namespace TUI.Pages;

public abstract class PageBase : IPage
{
    public void Open()
    {
        Debugger.Log(0, "Event", $"Open page {GetType().UnderlyingSystemType.Name}\n");
        Log.Trace($"Open page {GetType().UnderlyingSystemType.Name}.");
        Bind();
        Initial();
        Render();
    }

    public abstract void Load();

    public abstract void Initial();

    public abstract void Render();

    public abstract void Bind();
}