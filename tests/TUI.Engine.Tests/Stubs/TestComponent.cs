using TUI.Engine.Components;

namespace TUI.Engine.Tests.Stubs;

public class TestComponent : ComponentAttribute
{
    private string _content = "Lorem";

    public void SetContent(string content)
    {
        _content = content;
    }

    protected override Sketch DrawComponent()
    {
        return new Sketch(_content);
    }
}