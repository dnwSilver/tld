using System.Text;
using TUI.Engine;
using TUI.Engine.Attributes;
using TUI.Engine.Attributes.Orientations;
using TUI.Engine.Components;

namespace TUI.Controls.Common;

public class StubComponent : ComponentBase
{
    private readonly Size _size;

    public StubComponent(Size size)
    {
        _size = size;
        SetFixed(Orientation.Horizontal, size.Width);
        SetFixed(Orientation.Vertical, size.Height);
    }

    protected override Sketch DrawComponent(Size minSize)
    {
        var builder = new StringBuilder();
        var height = 0;

        while (_size.Height > height)
        {
            builder.Append(Symbols.Space.Repeat(_size.Width));
            height++;
        }

        return new Sketch(builder.ToString());
    }
}