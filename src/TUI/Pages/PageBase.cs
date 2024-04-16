using System.Diagnostics;

namespace TUI.Pages;

public abstract class PageBase : IPage
{
    public void Open()
    {
        Debugger.Log(0, "Event", $"Open page ${GetType().UnderlyingSystemType.Name}\n");
        Bind();
        Render();
    }

    public abstract void Render();

    public abstract void Bind();
}