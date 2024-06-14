using System.Text;
using TUI.Engine;
using TUI.Engine.Attributes;
using TUI.Engine.Attributes.Orientations;
using TUI.Engine.Components;

namespace TUI.Controls.Common;

public class StubComponent : ComponentBase
{
    private readonly Size _size;
    private readonly string? _text;
    
    public StubComponent(Size size, string? text = null)
    {
        _size = size;
        _text = text;
        SetFixed(Orientation.Horizontal, size.Width);
        SetFixed(Orientation.Vertical, size.Height);
    }
    
    protected override Sketch DrawComponent(Size minSize)
    {
        var builder = new StringBuilder();
        var height = 0;
        
        
        while (_size.Height > height)
        {
            builder.Append(Symbols.Space.Repeat(_size.Width - (_text?.GetWidth() ?? 0)));
            height++;
        }
        
        if (_text is not null)
        {
            builder.Append(_text);
        }
        return new Sketch(builder.ToString());
    }
}