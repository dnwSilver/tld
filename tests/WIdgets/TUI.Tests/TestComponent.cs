using TUI.Engine.Nodes.Components;

namespace Widgets.Tests;

internal class TestComponent : ComponentBase
{
    private string _content = "Lorem";

    public void SetContent(string content)
    {
        _content = content;
    }

    public override Sketch DrawComponent()
    {
        return new Sketch(_content);
    }
}