using TUI.Engine.Attributes;
using TUI.Engine.Components;

namespace TUI.Engine.Tests.Stubs;

public class TestComponent : ComponentBase
{
    private string _content = "Lorem";

    public void SetContent(string content)
    {
        _content = content;
    }

    protected override Sketch DrawComponent(Size minSize)
    {
        return new Sketch(_content);
    }
}